package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"dbm-lite/config"
	"dbm-lite/internal/adapter"
	"dbm-lite/internal/database"
	"dbm-lite/internal/dbtype"
	"dbm-lite/internal/model"
	"dbm-lite/internal/model/datasource_auth"
	"dbm-lite/pkg/crypto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InternalAuthService struct {
	cache sync.Map
}

func NewInternalAuthService() *InternalAuthService {
	return &InternalAuthService{}
}

func (s *InternalAuthService) generateID() string {
	return "dauth_" + time.Now().Format("20060102150405") + "_" + uuid.New().String()[:8]
}

func (s *InternalAuthService) CreateInternalUser(req struct {
	DatasourceID string
	Username     string
	Host         string
	Password     string
	Remark       string
}, operatorID, operatorName string) (*datasource_auth.DatasourceInternalUser, error) {
	if req.Host == "" {
		req.Host = "%"
	}

	ds, err := GetDatasourceByID(req.DatasourceID)
	if err != nil {
		return nil, err
	}

	adapter := adapter.NewPermissionAdapter(ds.DBType)

	if ds.DBType != model.DBTypeSQLite {
		password := req.Password
		if password == "" {
			password = uuid.New().String()[:12]
		}

		params := &dbtype.ConnectionParams{
			Type:     ds.DBType,
			Host:     ds.Host,
			Port:     ds.Port,
			Username: ds.Username,
			Password: ds.Password,
			Charset:  ds.Charset,
			Timezone: ds.Timezone,
		}
		conn, err := dbtype.Connect(req.DatasourceID, params)
		if err != nil {
			return nil, err
		}
		defer conn.DB.Close()

		if err := adapter.CreateUser(conn.DB, req.Username, req.Host, password); err != nil {
			return nil, err
		}

		encryptedPwd, err := crypto.EncryptAES(password, config.App.AESKey)
		if err != nil {
			return nil, err
		}

		user := &datasource_auth.DatasourceInternalUser{
			ID:           s.generateID(),
			DatasourceID: req.DatasourceID,
			Username:     req.Username,
			Host:         req.Host,
			Password:     encryptedPwd,
			Status:       datasource_auth.UserStatusActive,
			Remark:       req.Remark,
			CreatedAt:    time.Now(),
			CreatedBy:    operatorID,
			UpdatedAt:    time.Now(),
		}

		if err := database.DB.Create(user).Error; err != nil {
			return nil, err
		}

		s.LogAudit(operatorID, operatorName, "create_user", req.Username, req.DatasourceID, datasource_auth.AuditResultSuccess, "")
		return user, nil
	}

	user := &datasource_auth.DatasourceInternalUser{
		ID:           s.generateID(),
		DatasourceID: req.DatasourceID,
		Username:     req.Username,
		Host:         req.Host,
		Status:       datasource_auth.UserStatusActive,
		Remark:       req.Remark,
		CreatedAt:    time.Now(),
		CreatedBy:    operatorID,
		UpdatedAt:    time.Now(),
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, err
	}

	s.LogAudit(operatorID, operatorName, "create_user", req.Username, req.DatasourceID, datasource_auth.AuditResultSuccess, "")
	return user, nil
}

func (s *InternalAuthService) ListInternalUsers(datasourceID, keyword, status string, page, pageSize int) ([]datasource_auth.DatasourceInternalUser, int64, error) {
	query := database.DB.Model(&datasource_auth.DatasourceInternalUser{})
	if datasourceID != "" {
		query = query.Where("datasource_id = ?", datasourceID)
	}
	if keyword != "" {
		query = query.Where("username LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []datasource_auth.DatasourceInternalUser
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *InternalAuthService) GetInternalUser(id string) (*datasource_auth.DatasourceInternalUser, error) {
	var user datasource_auth.DatasourceInternalUser
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *InternalAuthService) UpdateInternalUser(id string, req struct {
	Host   string
	Remark string
}, operatorID, operatorName string) error {
	user, err := s.GetInternalUser(id)
	if err != nil {
		return err
	}

	if user.IsBuiltIn {
		return fmt.Errorf("内置账号不允许修改")
	}

	update := map[string]interface{}{"updated_at": time.Now()}
	if req.Host != "" {
		update["host"] = req.Host
	}
	if req.Remark != "" {
		update["remark"] = req.Remark
	}

	if err := database.DB.Model(&user).Updates(update).Error; err != nil {
		return err
	}

	s.LogAudit(operatorID, operatorName, "update_user", user.Username, user.DatasourceID, datasource_auth.AuditResultSuccess, "")
	return nil
}

func (s *InternalAuthService) DeleteInternalUser(id string, operatorID, operatorName string) error {
	user, err := s.GetInternalUser(id)
	if err != nil {
		return err
	}

	if user.IsBuiltIn {
		return fmt.Errorf("内置账号不允许删除")
	}

	ds, err := GetDatasourceByID(user.DatasourceID)
	if err != nil {
		return err
	}

	if ds.DBType != model.DBTypeSQLite {
		params := &dbtype.ConnectionParams{
			Type:     ds.DBType,
			Host:     ds.Host,
			Port:     ds.Port,
			Username: ds.Username,
			Password: ds.Password,
			Charset:  ds.Charset,
			Timezone: ds.Timezone,
		}
		conn, err := dbtype.Connect(user.DatasourceID, params)
		if err != nil {
			return err
		}
		defer conn.DB.Close()

		adapter := adapter.NewPermissionAdapter(ds.DBType)
		if err := adapter.DropUser(conn.DB, user.Username, user.Host); err != nil {
			return err
		}
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		return err
	}

	s.invalidateCache(user.DatasourceID)
	s.LogAudit(operatorID, operatorName, "delete_user", user.Username, user.DatasourceID, datasource_auth.AuditResultSuccess, "")
	return nil
}

func (s *InternalAuthService) ResetPassword(id string, newPassword string, operatorID, operatorName string) error {
	user, err := s.GetInternalUser(id)
	if err != nil {
		return err
	}

	if user.IsBuiltIn {
		return fmt.Errorf("内置账号不允许重置密码")
	}

	ds, err := GetDatasourceByID(user.DatasourceID)
	if err != nil {
		return err
	}

	if ds.DBType != model.DBTypeSQLite {
		if newPassword == "" {
			newPassword = uuid.New().String()[:12]
		}

		params := &dbtype.ConnectionParams{
			Type:     ds.DBType,
			Host:     ds.Host,
			Port:     ds.Port,
			Username: ds.Username,
			Password: ds.Password,
			Charset:  ds.Charset,
			Timezone: ds.Timezone,
		}
		conn, err := dbtype.Connect(user.DatasourceID, params)
		if err != nil {
			return err
		}
		defer conn.DB.Close()

		adapter := adapter.NewPermissionAdapter(ds.DBType)
		if err := adapter.AlterUser(conn.DB, user.Username, user.Host, newPassword); err != nil {
			return err
		}

		encryptedPwd, err := crypto.EncryptAES(newPassword, config.App.AESKey)
		if err != nil {
			return err
		}

		if err := database.DB.Model(&user).Update("password", encryptedPwd).Update("updated_at", time.Now()).Error; err != nil {
			return err
		}
	}

	s.invalidateCache(user.DatasourceID)
	s.LogAudit(operatorID, operatorName, "reset_password", user.Username, user.DatasourceID, datasource_auth.AuditResultSuccess, "")
	return nil
}

func (s *InternalAuthService) ToggleUserStatus(id string, enable bool, operatorID, operatorName string) error {
	user, err := s.GetInternalUser(id)
	if err != nil {
		return err
	}

	if user.IsBuiltIn {
		return fmt.Errorf("内置账号不允许修改状态")
	}

	ds, err := GetDatasourceByID(user.DatasourceID)
	if err != nil {
		return err
	}

	if ds.DBType != model.DBTypeSQLite {
		params := &dbtype.ConnectionParams{
			Type:     ds.DBType,
			Host:     ds.Host,
			Port:     ds.Port,
			Username: ds.Username,
			Password: ds.Password,
			Charset:  ds.Charset,
			Timezone: ds.Timezone,
		}
		conn, err := dbtype.Connect(user.DatasourceID, params)
		if err != nil {
			return err
		}
		defer conn.DB.Close()

		adapter := adapter.NewPermissionAdapter(ds.DBType)
		if err := adapter.EnableUser(conn.DB, user.Username, user.Host, enable); err != nil {
			return err
		}
	}

	newStatus := datasource_auth.UserStatusActive
	if !enable {
		newStatus = datasource_auth.UserStatusInactive
	}

	if err := database.DB.Model(&user).Update("status", newStatus).Update("updated_at", time.Now()).Error; err != nil {
		return err
	}

	s.invalidateCache(user.DatasourceID)
	s.LogAudit(operatorID, operatorName, "toggle_user_status", user.Username, user.DatasourceID, datasource_auth.AuditResultSuccess, fmt.Sprintf("enable=%v", enable))
	return nil
}

func (s *InternalAuthService) CreateInternalRole(req struct {
	DatasourceID string
	Name         string
	Description  string
}, operatorID, operatorName string) (*datasource_auth.DatasourceInternalRole, error) {
	role := &datasource_auth.DatasourceInternalRole{
		ID:          s.generateID(),
		DatasourceID: req.DatasourceID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := database.DB.Create(role).Error; err != nil {
		return nil, err
	}

	s.LogAudit(operatorID, operatorName, "create_role", req.Name, req.DatasourceID, datasource_auth.AuditResultSuccess, "")
	return role, nil
}

func (s *InternalAuthService) ListInternalRoles(datasourceID, keyword string, page, pageSize int) ([]datasource_auth.DatasourceInternalRole, int64, error) {
	query := database.DB.Model(&datasource_auth.DatasourceInternalRole{})
	if datasourceID != "" {
		query = query.Where("datasource_id = ?", datasourceID)
	}
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []datasource_auth.DatasourceInternalRole
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *InternalAuthService) GetInternalRole(id string) (*datasource_auth.DatasourceInternalRole, error) {
	var role datasource_auth.DatasourceInternalRole
	if err := database.DB.Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (s *InternalAuthService) UpdateInternalRole(id string, req struct {
	Name        string
	Description string
}, operatorID, operatorName string) error {
	role, err := s.GetInternalRole(id)
	if err != nil {
		return err
	}

	update := map[string]interface{}{"updated_at": time.Now()}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Description != "" {
		update["description"] = req.Description
	}

	if err := database.DB.Model(&role).Updates(update).Error; err != nil {
		return err
	}

	s.invalidateCache(role.DatasourceID)
	s.LogAudit(operatorID, operatorName, "update_role", role.Name, role.DatasourceID, datasource_auth.AuditResultSuccess, "")
	return nil
}

func (s *InternalAuthService) DeleteInternalRole(id string, operatorID, operatorName string) error {
	role, err := s.GetInternalRole(id)
	if err != nil {
		return err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&datasource_auth.DatasourceUserRoleRel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("principal_type = ? AND principal_id = ?", datasource_auth.PrincipalTypeRole, id).Delete(&datasource_auth.DatasourcePermissionRule{}).Error; err != nil {
			return err
		}
		return tx.Delete(&role).Error
	}); err != nil {
		return err
	}

	s.invalidateCache(role.DatasourceID)
	s.LogAudit(operatorID, operatorName, "delete_role", role.Name, role.DatasourceID, datasource_auth.AuditResultSuccess, "")
	return nil
}

func (s *InternalAuthService) GetRoleUserCount(roleID string) (int64, error) {
	var count int64
	err := database.DB.Model(&datasource_auth.DatasourceUserRoleRel{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

func (s *InternalAuthService) AssignRoles(userID string, roleIDs []string, operatorID, operatorName string) error {
	user, err := s.GetInternalUser(userID)
	if err != nil {
		return err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&datasource_auth.DatasourceUserRoleRel{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			rel := &datasource_auth.DatasourceUserRoleRel{
				ID:      s.generateID(),
				UserID:  userID,
				RoleID:  rid,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(rel).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.invalidateCache(user.DatasourceID)
	s.LogAudit(operatorID, operatorName, "assign_roles", user.Username, user.DatasourceID, datasource_auth.AuditResultSuccess, fmt.Sprintf("roles=%v", roleIDs))
	return nil
}

func (s *InternalAuthService) GetUserRoles(userID string) ([]datasource_auth.DatasourceInternalRole, error) {
	var rels []datasource_auth.DatasourceUserRoleRel
	if err := database.DB.Where("user_id = ?", userID).Find(&rels).Error; err != nil {
		return nil, err
	}

	var roleIDs []string
	for _, rel := range rels {
		roleIDs = append(roleIDs, rel.RoleID)
	}

	if len(roleIDs) == 0 {
		return nil, nil
	}

	var roles []datasource_auth.DatasourceInternalRole
	err := database.DB.Where("id IN ?", roleIDs).Find(&roles).Error
	return roles, err
}

func (s *InternalAuthService) GrantPermission(req struct {
	DatasourceID  string
	PrincipalType string
	PrincipalID   string
	PrivilegeType string
	ObjectLevel   string
	DatabaseName  string
	TableName     string
	Columns       []string
}, operatorID, operatorName string) error {
	if req.ObjectLevel == datasource_auth.ObjectLevelColumn && len(req.Columns) == 0 {
		return fmt.Errorf("列级权限必须指定列名")
	}

	columnsJSON, _ := json.Marshal(req.Columns)

	rule := &datasource_auth.DatasourcePermissionRule{
		ID:             s.generateID(),
		DatasourceID:   req.DatasourceID,
		PrincipalType:  req.PrincipalType,
		PrincipalID:    req.PrincipalID,
		PrivilegeType:  req.PrivilegeType,
		ObjectLevel:    req.ObjectLevel,
		DatabaseName:   req.DatabaseName,
		Table:          req.TableName,
		Columns:        string(columnsJSON),
		Enabled:        true,
		CreatedAt:      time.Now(),
	}

	if err := database.DB.Create(rule).Error; err != nil {
		return err
	}

	var principalName string
	if req.PrincipalType == datasource_auth.PrincipalTypeUser {
		user, _ := s.GetInternalUser(req.PrincipalID)
		if user != nil {
			principalName = user.Username
		}
	} else {
		role, _ := s.GetInternalRole(req.PrincipalID)
		if role != nil {
			principalName = role.Name
		}
	}

	s.invalidateCache(req.DatasourceID)
	s.LogAudit(operatorID, operatorName, "grant_permission", principalName, req.DatasourceID, datasource_auth.AuditResultSuccess, fmt.Sprintf("priv=%s level=%s db=%s table=%s", req.PrivilegeType, req.ObjectLevel, req.DatabaseName, req.TableName))
	return nil
}

func (s *InternalAuthService) BatchGrantPermission(req struct {
	DatasourceID  string
	PrincipalType string
	PrincipalID   string
	PrivilegeType string
	Rules         []struct {
		ObjectLevel  string
		DatabaseName string
		TableName    string
		Columns      []string
	}
}, operatorID, operatorName string) error {
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, r := range req.Rules {
			columnsJSON, _ := json.Marshal(r.Columns)
			rule := &datasource_auth.DatasourcePermissionRule{
				ID:             s.generateID(),
				DatasourceID:   req.DatasourceID,
				PrincipalType:  req.PrincipalType,
				PrincipalID:    req.PrincipalID,
				PrivilegeType:  req.PrivilegeType,
				ObjectLevel:    r.ObjectLevel,
				DatabaseName:   r.DatabaseName,
				Table:          r.TableName,
				Columns:        string(columnsJSON),
				Enabled:        true,
				CreatedAt:      time.Now(),
			}
			if err := tx.Create(rule).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.invalidateCache(req.DatasourceID)

	var principalName string
	if req.PrincipalType == datasource_auth.PrincipalTypeUser {
		user, _ := s.GetInternalUser(req.PrincipalID)
		if user != nil {
			principalName = user.Username
		}
	} else {
		role, _ := s.GetInternalRole(req.PrincipalID)
		if role != nil {
			principalName = role.Name
		}
	}

	s.LogAudit(operatorID, operatorName, "batch_grant_permission", principalName, req.DatasourceID, datasource_auth.AuditResultSuccess, fmt.Sprintf("priv=%s rules=%d", req.PrivilegeType, len(req.Rules)))
	return nil
}

func (s *InternalAuthService) RevokePermission(ruleID string, operatorID, operatorName string) error {
	var rule datasource_auth.DatasourcePermissionRule
	if err := database.DB.Where("id = ?", ruleID).First(&rule).Error; err != nil {
		return err
	}

	if err := database.DB.Delete(&rule).Error; err != nil {
		return err
	}

	s.invalidateCache(rule.DatasourceID)
	s.LogAudit(operatorID, operatorName, "revoke_permission", ruleID, rule.DatasourceID, datasource_auth.AuditResultSuccess, "")
	return nil
}

func (s *InternalAuthService) BatchRevokePermission(ruleIDs []string, operatorID, operatorName string) error {
	var rules []datasource_auth.DatasourcePermissionRule
	if err := database.DB.Where("id IN ?", ruleIDs).Find(&rules).Error; err != nil {
		return err
	}

	if len(rules) == 0 {
		return nil
	}

	datasourceID := rules[0].DatasourceID

	if err := database.DB.Where("id IN ?", ruleIDs).Delete(&datasource_auth.DatasourcePermissionRule{}).Error; err != nil {
		return err
	}

	s.invalidateCache(datasourceID)
	s.LogAudit(operatorID, operatorName, "batch_revoke_permission", fmt.Sprintf("%d rules", len(ruleIDs)), datasourceID, datasource_auth.AuditResultSuccess, "")
	return nil
}

func (s *InternalAuthService) ListPermissionRules(datasourceID, principalType, principalID, privilegeType, objectLevel, privilegeCategory string, page, pageSize int) ([]datasource_auth.DatasourcePermissionRule, int64, error) {
	query := database.DB.Model(&datasource_auth.DatasourcePermissionRule{}).Where("enabled = ?", true)
	if datasourceID != "" {
		query = query.Where("datasource_id = ?", datasourceID)
	}
	if principalType != "" {
		query = query.Where("principal_type = ?", principalType)
	}
	if principalID != "" {
		query = query.Where("principal_id = ?", principalID)
	}
	if privilegeType != "" {
		query = query.Where("privilege_type = ?", privilegeType)
	}
	if objectLevel != "" {
		query = query.Where("object_level = ?", objectLevel)
	}
	if privilegeCategory != "" {
		query = query.Where("privilege_category = ?", privilegeCategory)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []datasource_auth.DatasourcePermissionRule
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	for i := range list {
		var principalName string
		if list[i].PrincipalType == datasource_auth.PrincipalTypeUser {
			user, _ := s.GetInternalUser(list[i].PrincipalID)
			if user != nil {
				principalName = user.Username
			}
		} else {
			role, _ := s.GetInternalRole(list[i].PrincipalID)
			if role != nil {
				principalName = role.Name
			}
		}
		list[i].PrincipalID = principalName
	}

	return list, total, nil
}

func (s *InternalAuthService) GetUserEffectivePermissions(datasourceID, userID string) ([]datasource_auth.DatasourcePermissionRule, error) {
	cacheKey := fmt.Sprintf("perm_%s_%s", datasourceID, userID)
	if cached, ok := s.cache.Load(cacheKey); ok {
		return cached.([]datasource_auth.DatasourcePermissionRule), nil
	}

	var rules []datasource_auth.DatasourcePermissionRule

	userRules, err := s.getUserDirectPermissions(userID)
	if err != nil {
		return nil, err
	}
	rules = append(rules, userRules...)

	roleRules, err := s.getUserRolePermissions(userID)
	if err != nil {
		return nil, err
	}
	rules = append(rules, roleRules...)

	dsRules, err := s.getDatasourceLevelPermissions(datasourceID, userID)
	if err != nil {
		return nil, err
	}
	rules = append(rules, dsRules...)

	s.cache.Store(cacheKey, rules)
	return rules, nil
}

func (s *InternalAuthService) getUserDirectPermissions(userID string) ([]datasource_auth.DatasourcePermissionRule, error) {
	var rules []datasource_auth.DatasourcePermissionRule
	err := database.DB.Where("principal_type = ? AND principal_id = ? AND enabled = ?",
		datasource_auth.PrincipalTypeUser, userID, true).Find(&rules).Error
	return rules, err
}

func (s *InternalAuthService) getUserRolePermissions(userID string) ([]datasource_auth.DatasourcePermissionRule, error) {
	var rels []datasource_auth.DatasourceUserRoleRel
	if err := database.DB.Where("user_id = ?", userID).Find(&rels).Error; err != nil {
		return nil, err
	}

	var roleIDs []string
	for _, rel := range rels {
		roleIDs = append(roleIDs, rel.RoleID)
	}

	if len(roleIDs) == 0 {
		return nil, nil
	}

	var rules []datasource_auth.DatasourcePermissionRule
	err := database.DB.Where("principal_type = ? AND principal_id IN ? AND enabled = ?",
		datasource_auth.PrincipalTypeRole, roleIDs, true).Find(&rules).Error
	return rules, err
}

func (s *InternalAuthService) getDatasourceLevelPermissions(datasourceID, userID string) ([]datasource_auth.DatasourcePermissionRule, error) {
	return nil, nil
}

func (s *InternalAuthService) CheckSQLPermission(datasourceID, userID, sqlText string) (bool, string) {
	rules, err := s.GetUserEffectivePermissions(datasourceID, userID)
	if err != nil {
		return false, "获取权限规则失败: " + err.Error()
	}

	if len(rules) == 0 {
		return false, "用户无任何权限"
	}

	sqlLower := strings.ToLower(sqlText)

	sqlTrimmed := strings.TrimSpace(sqlLower)
	sqlWords := strings.Fields(sqlTrimmed)

	if len(sqlWords) == 0 {
		return false, "无效的SQL语句"
	}

	firstWord := sqlWords[0]

	ddlKeywords := []string{"create", "alter", "drop", "truncate", "rename", "grant", "revoke"}
	dmlKeywords := []string{"insert", "update", "delete", "merge", "replace", "call"}
	selectKeywords := []string{"select", "show", "describe", "explain", "desc"}
	adminKeywords := []string{"flush", "shutdown", "reload", "kill", "set", "reset"}

	isDDL := containsAny(firstWord, ddlKeywords)
	isDML := containsAny(firstWord, dmlKeywords)
	isSelect := containsAny(firstWord, selectKeywords)
	isAdmin := containsAny(firstWord, adminKeywords)

	targetDB, targetTable := extractDBAndTable(sqlText)

	for _, rule := range rules {
		if rule.PrivilegeCategory == datasource_auth.PrivilegeCategorySystem {
			if isAdmin {
				return true, ""
			}
			continue
		}

		ruleDB := rule.DatabaseName
		ruleTable := rule.Table

		if !matchesDB(targetDB, ruleDB) {
			continue
		}

		if !matchesTable(targetTable, ruleTable) {
			continue
		}

		switch rule.PrivilegeType {
		case datasource_auth.PrivilegeTypeDDL:
			if isDDL || isDML || isSelect {
				return true, ""
			}
		case datasource_auth.PrivilegeTypeDML:
			if isDML || isSelect {
				return true, ""
			}
		case datasource_auth.PrivilegeTypeReadonly:
			if isSelect {
				return true, ""
			}
		}
	}

	if isDDL {
		return false, "无DDL权限"
	}
	if isDML {
		return false, "无DML权限"
	}
	if isSelect {
		return false, "无查询权限"
	}
	if isAdmin {
		return false, "无管理权限"
	}

	return false, "无权限执行此操作"
}

func containsAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if s == kw {
			return true
		}
	}
	return false
}

func extractDBAndTable(sqlText string) (string, string) {
	sqlLower := strings.ToLower(strings.TrimSpace(sqlText))

	var dbName, tableName string

	if strings.HasPrefix(sqlLower, "use ") {
		parts := strings.Fields(sqlLower)
		if len(parts) >= 2 {
			dbName = strings.Trim(parts[1], "`'\";")
		}
		return dbName, ""
	}

	ddlPatterns := []string{"create table ", "alter table ", "drop table ", "truncate table ", "rename table ", "create database ", "drop database "}
	for _, pattern := range ddlPatterns {
		if idx := strings.Index(sqlLower, pattern); idx != -1 {
			rest := sqlLower[idx+len(pattern):]
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				fullName := strings.Trim(parts[0], "`'\";")
				if strings.Contains(fullName, ".") {
					segments := strings.SplitN(fullName, ".", 2)
					dbName = segments[0]
					tableName = segments[1]
				} else if strings.HasPrefix(pattern, "create database") || strings.HasPrefix(pattern, "drop database") {
					dbName = fullName
				} else {
					tableName = fullName
				}
			}
			return dbName, tableName
		}
	}

	dmlPatterns := []string{"insert into ", "update ", "delete from ", "merge into ", "replace into "}
	for _, pattern := range dmlPatterns {
		if idx := strings.Index(sqlLower, pattern); idx != -1 {
			rest := sqlLower[idx+len(pattern):]
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				fullName := strings.Trim(parts[0], "`'\";")
				if strings.Contains(fullName, ".") {
					segments := strings.SplitN(fullName, ".", 2)
					dbName = segments[0]
					tableName = segments[1]
				} else {
					tableName = fullName
				}
			}
			return dbName, tableName
		}
	}

	if strings.HasPrefix(sqlLower, "select ") {
		fromIdx := strings.Index(sqlLower, " from ")
		if fromIdx != -1 {
			rest := sqlLower[fromIdx+6:]
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				fullName := strings.Trim(parts[0], "`'\";")
				if strings.Contains(fullName, ".") {
					segments := strings.SplitN(fullName, ".", 2)
					dbName = segments[0]
					tableName = segments[1]
				} else {
					tableName = fullName
				}
			}
		}
	}

	return dbName, tableName
}

func matchesDB(targetDB, ruleDB string) bool {
	if ruleDB == "*" || ruleDB == "" {
		return true
	}
	if targetDB == "" {
		return true
	}
	return targetDB == ruleDB
}

func matchesTable(targetTable, ruleTable string) bool {
	if ruleTable == "*" || ruleTable == "" {
		return true
	}
	if targetTable == "" {
		return true
	}
	return targetTable == ruleTable
}

func (s *InternalAuthService) invalidateCache(datasourceID string) {
	s.cache.Range(func(key, value interface{}) bool {
		if strings.HasPrefix(key.(string), "perm_"+datasourceID) {
			s.cache.Delete(key)
		}
		return true
	})
}

func (s *InternalAuthService) LogAudit(operatorID, operatorName, operType, operObject, datasourceID, result, detail string) {
	audit := &datasource_auth.DatasourceAuthAudit{
		ID:           s.generateID(),
		Operator:     operatorName,
		OperatorID:   operatorID,
		OperType:     operType,
		OperObject:   operObject,
		DatasourceID: datasourceID,
		Result:       result,
		Detail:       detail,
		OperTime:     time.Now(),
	}
	go func() {
		database.DB.Create(audit)
	}()
}

func (s *InternalAuthService) ListAuditLogs(datasourceID, operator, operType, result string, page, pageSize int) ([]datasource_auth.DatasourceAuthAudit, int64, error) {
	query := database.DB.Model(&datasource_auth.DatasourceAuthAudit{})
	if datasourceID != "" {
		query = query.Where("datasource_id = ?", datasourceID)
	}
	if operator != "" {
		query = query.Where("operator LIKE ?", "%"+operator+"%")
	}
	if operType != "" {
		query = query.Where("oper_type = ?", operType)
	}
	if result != "" {
		query = query.Where("result = ?", result)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []datasource_auth.DatasourceAuthAudit
	if err := query.Order("oper_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *InternalAuthService) SyncDBUsers(datasourceID string) error {
	ds, err := GetDatasourceByID(datasourceID)
	if err != nil {
		return err
	}

	password, err := dbtype.DecryptPassword(ds.Password)
	if err != nil {
		return fmt.Errorf("decrypt password failed: %w", err)
	}

	params := &dbtype.ConnectionParams{
		Type:       ds.DBType,
		Host:       ds.Host,
		Port:       ds.Port,
		Username:   ds.Username,
		Password:   password,
		Charset:    ds.Charset,
		Timezone:   ds.Timezone,
		SSLMode:    ds.SSLMode,
		SSLCAFile:  ds.SSLCAFile,
	}

	var sqlDB *sql.DB
	switch ds.DBType {
	case model.DBTypeMySQL, model.DBTypeTiDB:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("mysql", dsn)
	case model.DBTypeSQLite:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("sqlite", dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", ds.DBType)
	}
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(1)
	if err = sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	adapter := adapter.NewPermissionAdapter(ds.DBType)
	dbUsers, err := adapter.ListUsers(sqlDB)
	if err != nil {
		return err
	}

	builtInUsers := map[string]bool{"root": true, "admin": true, "mysql": true, "postgres": true}

	for _, u := range dbUsers {
		username := ""
		if v, ok := u["username"]; ok {
			username = fmt.Sprintf("%v", v)
		}
		host := "%"
		if v, ok := u["host"]; ok {
			host = fmt.Sprintf("%v", v)
		}
		if host == "" {
			host = "%"
		}

		var exist datasource_auth.DatasourceInternalUser
		if err := database.DB.Where("datasource_id = ? AND username = ? AND host = ?",
			datasourceID, username, host).First(&exist).Error; err == gorm.ErrRecordNotFound {

			isBuiltIn := builtInUsers[strings.ToLower(username)]

			user := &datasource_auth.DatasourceInternalUser{
				ID:           s.generateID(),
				DatasourceID: datasourceID,
				Username:     username,
				Host:         host,
				IsBuiltIn:    isBuiltIn,
				Status:       datasource_auth.UserStatusActive,
				CreatedAt:    time.Now(),
				CreatedBy:    "system",
				UpdatedAt:    time.Now(),
			}
			database.DB.Create(user)
		}
	}

	dbPrivileges, err := adapter.ListUserPrivileges(sqlDB)
	if err != nil {
		return err
	}

	for _, priv := range dbPrivileges {
		username := ""
		if v, ok := priv["username"]; ok {
			username = fmt.Sprintf("%v", v)
		}
		host := "%"
		if v, ok := priv["host"]; ok {
			host = fmt.Sprintf("%v", v)
		}
		if host == "" {
			host = "%"
		}
		databaseName := ""
		if v, ok := priv["database_name"]; ok {
			databaseName = fmt.Sprintf("%v", v)
		}
		tableName := ""
		if v, ok := priv["table_name"]; ok {
			tableName = fmt.Sprintf("%v", v)
		}
		privType := ""
		if v, ok := priv["privilege_type"]; ok {
			privType = fmt.Sprintf("%v", v)
		}

		if username == "" || databaseName == "" {
			continue
		}

		var user datasource_auth.DatasourceInternalUser
		if err := database.DB.Where("datasource_id = ? AND username = ? AND host = ?",
			datasourceID, username, host).First(&user).Error; err != nil {
			continue
		}

		objectLevel := datasource_auth.ObjectLevelDatabase
		if tableName != "" {
			objectLevel = datasource_auth.ObjectLevelTable
		}

		var existRule datasource_auth.DatasourcePermissionRule
		if err := database.DB.Where("datasource_id = ? AND principal_type = ? AND principal_id = ? AND privilege_type = ? AND object_level = ? AND database_name = ? AND table_name = ?",
			datasourceID, datasource_auth.PrincipalTypeUser, user.ID, privType, objectLevel, databaseName, tableName).First(&existRule).Error; err == gorm.ErrRecordNotFound {

			rule := &datasource_auth.DatasourcePermissionRule{
				ID:             s.generateID(),
				DatasourceID:   datasourceID,
				PrincipalType:  datasource_auth.PrincipalTypeUser,
				PrincipalID:    user.ID,
				PrivilegeType:  privType,
				ObjectLevel:    objectLevel,
				DatabaseName:   databaseName,
				Table:          tableName,
				Enabled:        true,
				CreatedAt:      time.Now(),
			}
			database.DB.Create(rule)
		}
	}

	s.invalidateCache(datasourceID)
	return nil
}

func (s *InternalAuthService) GetUserGrants(datasourceID, username, host string) (string, error) {
	ds, err := GetDatasourceByID(datasourceID)
	if err != nil {
		return "", err
	}

	password, err := dbtype.DecryptPassword(ds.Password)
	if err != nil {
		return "", fmt.Errorf("decrypt password failed: %w", err)
	}

	params := &dbtype.ConnectionParams{
		Type:       ds.DBType,
		Host:       ds.Host,
		Port:       ds.Port,
		Username:   ds.Username,
		Password:   password,
		Charset:    ds.Charset,
		Timezone:   ds.Timezone,
		SSLMode:    ds.SSLMode,
		SSLCAFile:  ds.SSLCAFile,
	}

	var sqlDB *sql.DB
	switch ds.DBType {
	case model.DBTypeMySQL, model.DBTypeTiDB:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("mysql", dsn)
	case model.DBTypeSQLite:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("sqlite", dsn)
	default:
		return "", fmt.Errorf("unsupported database type: %s", ds.DBType)
	}
	if err != nil {
		return "", fmt.Errorf("connect failed: %w", err)
	}
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(1)
	if err = sqlDB.Ping(); err != nil {
		return "", fmt.Errorf("ping failed: %w", err)
	}

	var grants []string
	switch ds.DBType {
	case model.DBTypeMySQL, model.DBTypeTiDB:
		stmt := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", username, host)
		rows, err := sqlDB.Query(stmt)
		if err != nil {
			return "", fmt.Errorf("show grants failed: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var grant string
			if err := rows.Scan(&grant); err != nil {
				return "", err
			}
			grants = append(grants, grant)
		}
	case model.DBTypeSQLite:
		grants = append(grants, "SQLite 不支持 SHOW GRANTS 命令")
		grants = append(grants, fmt.Sprintf("用户: %s 的权限需通过数据库文件权限控制", username))
	}

	return strings.Join(grants, "\n"), nil
}

type ObjectPermissionRule struct {
	ObjectType    string   `json:"objectType"`
	PrivilegeType string   `json:"privilegeType"`
	DatabaseName  string   `json:"databaseName"`
	TableName     string   `json:"tableName"`
	Columns       []string `json:"columns"`
}

type SaveUserPermissionRequest struct {
	DatasourceID       string                  `json:"datasourceId"`
	UserID             string                  `json:"userId"`
	ObjectPermissions  []ObjectPermissionRule  `json:"objectPermissions"`
	SystemPermissions  []string                `json:"systemPermissions"`
}

func (s *InternalAuthService) SaveUserPermissions(req SaveUserPermissionRequest, operatorID, operatorName string) error {
	user, err := s.GetInternalUser(req.UserID)
	if err != nil {
		return err
	}

	ds, err := GetDatasourceByID(req.DatasourceID)
	if err != nil {
		return err
	}

	var sqlDB *sql.DB
	var permAdapter adapter.PermissionAdapter
	var needExecuteDB bool

	if ds.DBType != model.DBTypeSQLite {
		needExecuteDB = true
		password := ds.Password
		if ds.Password != "" {
			if decrypted, err := dbtype.DecryptPassword(ds.Password); err == nil {
				password = decrypted
			}
		}
		if password == "" {
			return fmt.Errorf("datasource password is empty")
		}

		params := &dbtype.ConnectionParams{
			Type:       ds.DBType,
			Host:       ds.Host,
			Port:       ds.Port,
			Username:   ds.Username,
			Password:   password,
			Charset:    ds.Charset,
			Timezone:   ds.Timezone,
			SSLMode:    ds.SSLMode,
			SSLCAFile:  ds.SSLCAFile,
		}

		switch ds.DBType {
		case model.DBTypeMySQL, model.DBTypeTiDB:
			dsn, _, _ := dbtype.BuildDSN(params)
			sqlDB, err = sql.Open("mysql", dsn)
		default:
			return fmt.Errorf("unsupported database type: %s", ds.DBType)
		}
		if err != nil {
			return fmt.Errorf("connect failed: %w", err)
		}
		defer sqlDB.Close()

		sqlDB.SetMaxOpenConns(5)
		sqlDB.SetMaxIdleConns(1)
		if err = sqlDB.Ping(); err != nil {
			return fmt.Errorf("ping failed: %w", err)
		}

		permAdapter = adapter.NewPermissionAdapter(ds.DBType)
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingRules []datasource_auth.DatasourcePermissionRule
	if err := tx.Where("datasource_id = ? AND principal_type = ? AND principal_id = ?",
		req.DatasourceID, datasource_auth.PrincipalTypeUser, req.UserID).Find(&existingRules).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, rule := range existingRules {
		if needExecuteDB && sqlDB != nil && permAdapter != nil {
			if rule.PrivilegeCategory == datasource_auth.PrivilegeCategoryObject {
				var columns []string
				if rule.Columns != "" {
					json.Unmarshal([]byte(rule.Columns), &columns)
				}
				objType := adapter.ObjectType(rule.ObjectLevel)
				permAdapter.RevokeObject(sqlDB, user.Username, user.Host,
					s.getPrivilegeNames(rule.PrivilegeType), objType, rule.DatabaseName, rule.Table, columns)
			} else if rule.PrivilegeCategory == datasource_auth.PrivilegeCategorySystem {
				var sysPrivs []string
				if rule.SystemPrivileges != "" {
					json.Unmarshal([]byte(rule.SystemPrivileges), &sysPrivs)
				}
				if len(sysPrivs) > 0 {
					permAdapter.RevokeSystemPrivileges(sqlDB, user.Username, user.Host, sysPrivs)
				}
			}
		}
	}

	if err := tx.Where("datasource_id = ? AND principal_type = ? AND principal_id = ?",
		req.DatasourceID, datasource_auth.PrincipalTypeUser, req.UserID).Delete(&datasource_auth.DatasourcePermissionRule{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, objPerm := range req.ObjectPermissions {
		objectLevel := objPerm.ObjectType
		if objectLevel == "" {
			objectLevel = datasource_auth.ObjectLevelTable
		}

		columns := ""
		if len(objPerm.Columns) > 0 {
			cols, _ := json.Marshal(objPerm.Columns)
			columns = string(cols)
		}

		if needExecuteDB && sqlDB != nil && permAdapter != nil {
			objType := adapter.ObjectType(objectLevel)
			if err := permAdapter.GrantObject(sqlDB, user.Username, user.Host,
				s.getPrivilegeNames(objPerm.PrivilegeType), objType, objPerm.DatabaseName, objPerm.TableName, objPerm.Columns); err != nil {
				tx.Rollback()
				return err
			}
		}

		rule := &datasource_auth.DatasourcePermissionRule{
			ID:                s.generateID(),
			DatasourceID:      req.DatasourceID,
			PrincipalType:     datasource_auth.PrincipalTypeUser,
			PrincipalID:       req.UserID,
			PrivilegeType:     objPerm.PrivilegeType,
			ObjectLevel:       objectLevel,
			DatabaseName:      objPerm.DatabaseName,
			Table:             objPerm.TableName,
			Columns:           columns,
			PrivilegeCategory: datasource_auth.PrivilegeCategoryObject,
			Enabled:           true,
			CreatedAt:         time.Now(),
		}
		if err := tx.Create(rule).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(req.SystemPermissions) > 0 {
		if needExecuteDB && sqlDB != nil && permAdapter != nil {
			if err := permAdapter.GrantSystemPrivileges(sqlDB, user.Username, user.Host, req.SystemPermissions); err != nil {
				tx.Rollback()
				return err
			}
		}

		sysPrivs, _ := json.Marshal(req.SystemPermissions)
		rule := &datasource_auth.DatasourcePermissionRule{
			ID:                s.generateID(),
			DatasourceID:      req.DatasourceID,
			PrincipalType:     datasource_auth.PrincipalTypeUser,
			PrincipalID:       req.UserID,
			PrivilegeCategory: datasource_auth.PrivilegeCategorySystem,
			SystemPrivileges:  string(sysPrivs),
			Enabled:           true,
			CreatedAt:         time.Now(),
		}
		if err := tx.Create(rule).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if needExecuteDB && sqlDB != nil && permAdapter != nil {
		permAdapter.FlushPrivileges(sqlDB)
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	s.invalidateCache(req.DatasourceID)

	s.LogAudit(operatorID, operatorName, "grant_permission",
		fmt.Sprintf("用户: %s", user.Username), req.DatasourceID, "success",
		fmt.Sprintf("对象权限规则: %d 条, 系统权限: %d 项", len(req.ObjectPermissions), len(req.SystemPermissions)))

	return nil
}

func (s *InternalAuthService) getPrivilegeNames(privilegeType string) []string {
	switch privilegeType {
	case "ddl":
		return []string{"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP", "INDEX", "TRIGGER", "CREATE ROUTINE", "ALTER ROUTINE"}
	case "dml":
		return []string{"SELECT", "INSERT", "UPDATE", "DELETE", "EXECUTE"}
	case "readonly":
		return []string{"SELECT", "SHOW VIEW"}
	default:
		return []string{"SELECT"}
	}
}

func (s *InternalAuthService) GetRoleGrants(datasourceID, roleName string) (string, error) {
	ds, err := GetDatasourceByID(datasourceID)
	if err != nil {
		return "", err
	}

	password, err := dbtype.DecryptPassword(ds.Password)
	if err != nil {
		return "", fmt.Errorf("decrypt password failed: %w", err)
	}

	params := &dbtype.ConnectionParams{
		Type:       ds.DBType,
		Host:       ds.Host,
		Port:       ds.Port,
		Username:   ds.Username,
		Password:   password,
		Charset:    ds.Charset,
		Timezone:   ds.Timezone,
		SSLMode:    ds.SSLMode,
		SSLCAFile:  ds.SSLCAFile,
	}

	var sqlDB *sql.DB
	switch ds.DBType {
	case model.DBTypeMySQL, model.DBTypeTiDB:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("mysql", dsn)
	case model.DBTypeSQLite:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("sqlite", dsn)
	default:
		return "", fmt.Errorf("unsupported database type: %s", ds.DBType)
	}
	if err != nil {
		return "", fmt.Errorf("connect failed: %w", err)
	}
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(1)
	if err = sqlDB.Ping(); err != nil {
		return "", fmt.Errorf("ping failed: %w", err)
	}

	var grants []string
	switch ds.DBType {
	case model.DBTypeMySQL, model.DBTypeTiDB:
		stmt := fmt.Sprintf("SHOW GRANTS FOR '%s'", roleName)
		rows, err := sqlDB.Query(stmt)
		if err != nil {
			return "", fmt.Errorf("show grants failed: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var grant string
			if err := rows.Scan(&grant); err != nil {
				return "", err
			}
			grants = append(grants, grant)
		}
	case model.DBTypeSQLite:
		grants = append(grants, "SQLite 不支持角色和 SHOW GRANTS 命令")
	}

	return strings.Join(grants, "\n"), nil
}

func (s *InternalAuthService) GetUserEffectiveGrants(datasourceID, username, host string) ([]map[string]interface{}, error) {
	ds, err := GetDatasourceByID(datasourceID)
	if err != nil {
		return nil, err
	}

	password := ds.Password
	if ds.Password != "" {
		if decrypted, err := dbtype.DecryptPassword(ds.Password); err == nil {
			password = decrypted
		}
	}
	if password == "" {
		return nil, fmt.Errorf("datasource password is empty")
	}

	params := &dbtype.ConnectionParams{
		Type:       ds.DBType,
		Host:       ds.Host,
		Port:       ds.Port,
		Username:   ds.Username,
		Password:   password,
		Charset:    ds.Charset,
		Timezone:   ds.Timezone,
		SSLMode:    ds.SSLMode,
		SSLCAFile:  ds.SSLCAFile,
	}

	var sqlDB *sql.DB
	switch ds.DBType {
	case model.DBTypeMySQL, model.DBTypeTiDB:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("mysql", dsn)
	case model.DBTypeSQLite:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("sqlite", dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", ds.DBType)
	}
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(1)
	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	permAdapter := adapter.NewPermissionAdapter(ds.DBType)
	return permAdapter.GetUserEffectiveGrants(sqlDB, username, host)
}

func GetDatasourceByID(id string) (*model.Datasource, error) {
	var ds model.Datasource
	if err := database.DB.Where("datasource_id = ?", id).First(&ds).Error; err != nil {
		return nil, err
	}
	return &ds, nil
}

func (s *InternalAuthService) GrantSystemPrivileges(req struct {
	DatasourceID  string
	PrincipalType string
	PrincipalID   string
	SystemPrivileges []string
}, operatorID, operatorName string) error {
	var principalName, username, host string
	
	if req.PrincipalType == datasource_auth.PrincipalTypeUser {
		user, err := s.GetInternalUser(req.PrincipalID)
		if err != nil {
			return err
		}
		principalName = user.Username
		username = user.Username
		host = user.Host
	} else {
		role, err := s.GetInternalRole(req.PrincipalID)
		if err != nil {
			return err
		}
		principalName = role.Name
		username = role.Name
		host = "%"
	}

	ds, err := GetDatasourceByID(req.DatasourceID)
	if err != nil {
		return err
	}

	if ds.DBType == model.DBTypeSQLite {
		return fmt.Errorf("SQLite不支持系统权限管理")
	}

	params := &dbtype.ConnectionParams{
		Type:       ds.DBType,
		Host:       ds.Host,
		Port:       ds.Port,
		Username:   ds.Username,
		Password:   ds.Password,
		Charset:    ds.Charset,
		Timezone:   ds.Timezone,
		SSLMode:    ds.SSLMode,
		SSLCAFile:  ds.SSLCAFile,
	}

	var sqlDB *sql.DB
	switch ds.DBType {
	case model.DBTypeMySQL, model.DBTypeTiDB:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("mysql", dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", ds.DBType)
	}
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	defer sqlDB.Close()

	permAdapter := adapter.NewPermissionAdapter(ds.DBType)
	if err := permAdapter.GrantSystemPrivileges(sqlDB, username, host, req.SystemPrivileges); err != nil {
		return err
	}

	systemPrivilegesJSON, _ := json.Marshal(req.SystemPrivileges)

	rule := &datasource_auth.DatasourcePermissionRule{
		ID:              s.generateID(),
		DatasourceID:    req.DatasourceID,
		PrincipalType:   req.PrincipalType,
		PrincipalID:     req.PrincipalID,
		PrivilegeType:   "system",
		ObjectLevel:     "global",
		Enabled:         true,
		PrivilegeCategory: datasource_auth.PrivilegeCategorySystem,
		SystemPrivileges:  string(systemPrivilegesJSON),
		CreatedAt:       time.Now(),
	}

	if err := database.DB.Create(rule).Error; err != nil {
		return err
	}

	s.invalidateCache(req.DatasourceID)
	s.LogAudit(operatorID, operatorName, "grant_system_privileges", principalName, req.DatasourceID, datasource_auth.AuditResultSuccess, fmt.Sprintf("privs=%v", req.SystemPrivileges))
	return nil
}

func (s *InternalAuthService) RevokeSystemPrivileges(ruleID string, operatorID, operatorName string) error {
	var rule datasource_auth.DatasourcePermissionRule
	if err := database.DB.Where("id = ?", ruleID).First(&rule).Error; err != nil {
		return err
	}

	if rule.PrivilegeCategory != datasource_auth.PrivilegeCategorySystem {
		return fmt.Errorf("非系统权限规则")
	}

	ds, err := GetDatasourceByID(rule.DatasourceID)
	if err != nil {
		return err
	}

	var username, host string
	if rule.PrincipalType == datasource_auth.PrincipalTypeUser {
		user, err := s.GetInternalUser(rule.PrincipalID)
		if err != nil {
			return err
		}
		username = user.Username
		host = user.Host
	} else {
		role, err := s.GetInternalRole(rule.PrincipalID)
		if err != nil {
			return err
		}
		username = role.Name
		host = "%"
	}

	if ds.DBType != model.DBTypeSQLite {
		params := &dbtype.ConnectionParams{
			Type:       ds.DBType,
			Host:       ds.Host,
			Port:       ds.Port,
			Username:   ds.Username,
			Password:   ds.Password,
			Charset:    ds.Charset,
			Timezone:   ds.Timezone,
			SSLMode:    ds.SSLMode,
			SSLCAFile:  ds.SSLCAFile,
		}

		var sqlDB *sql.DB
		switch ds.DBType {
		case model.DBTypeMySQL, model.DBTypeTiDB:
			dsn, _, _ := dbtype.BuildDSN(params)
			sqlDB, err = sql.Open("mysql", dsn)
		default:
			return fmt.Errorf("unsupported database type: %s", ds.DBType)
		}
		if err != nil {
			return fmt.Errorf("connect failed: %w", err)
		}
		defer sqlDB.Close()

		var sysPrivs []string
		if rule.SystemPrivileges != "" {
			json.Unmarshal([]byte(rule.SystemPrivileges), &sysPrivs)
		}

		permAdapter := adapter.NewPermissionAdapter(ds.DBType)
		permAdapter.RevokeSystemPrivileges(sqlDB, username, host, sysPrivs)
	}

	if err := database.DB.Delete(&rule).Error; err != nil {
		return err
	}

	s.invalidateCache(rule.DatasourceID)
	s.LogAudit(operatorID, operatorName, "revoke_system_privileges", ruleID, rule.DatasourceID, datasource_auth.AuditResultSuccess, "")
	return nil
}

func (s *InternalAuthService) GetSystemPrivilegesList(datasourceID string) ([]adapter.SystemPrivilege, error) {
	ds, err := GetDatasourceByID(datasourceID)
	if err != nil {
		return nil, err
	}

	permAdapter := adapter.NewPermissionAdapter(ds.DBType)
	return permAdapter.GetSystemPrivileges(), nil
}

func (s *InternalAuthService) GetUserSystemPrivileges(datasourceID, userID string) ([]string, error) {
	var rules []datasource_auth.DatasourcePermissionRule
	err := database.DB.Where("datasource_id = ? AND principal_type = ? AND principal_id = ? AND privilege_category = ? AND enabled = ?",
		datasourceID, datasource_auth.PrincipalTypeUser, userID, datasource_auth.PrivilegeCategorySystem, true).Find(&rules).Error
	if err != nil {
		return nil, err
	}

	var privs []string
	for _, rule := range rules {
		var sysPrivs []string
		if rule.SystemPrivileges != "" {
			json.Unmarshal([]byte(rule.SystemPrivileges), &sysPrivs)
			privs = append(privs, sysPrivs...)
		}
	}

	return privs, nil
}

func (s *InternalAuthService) GetUserPermissionDetail(datasourceID, userID string) (map[string]interface{}, error) {
	user, err := s.GetInternalUser(userID)
	if err != nil {
		return nil, err
	}

	var objectRules []datasource_auth.DatasourcePermissionRule
	if err := database.DB.Where("datasource_id = ? AND principal_type = ? AND principal_id = ? AND privilege_category = ? AND enabled = ?",
		datasourceID, datasource_auth.PrincipalTypeUser, userID, datasource_auth.PrivilegeCategoryObject, true).Find(&objectRules).Error; err != nil {
		return nil, err
	}

	objectPermissions := make([]map[string]interface{}, len(objectRules))
	for i, rule := range objectRules {
		objectPermissions[i] = map[string]interface{}{
			"id":             rule.ID,
			"objectType":     rule.ObjectLevel,
			"privilegeType":  rule.PrivilegeType,
			"databaseName":   rule.DatabaseName,
			"tableName":      rule.Table,
			"columns":        rule.Columns,
			"createdAt":      rule.CreatedAt,
		}
	}

	systemPrivs, err := s.GetUserSystemPrivileges(datasourceID, userID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"user":              user,
		"objectPermissions": objectPermissions,
		"systemPermissions": systemPrivs,
	}, nil
}

func (s *InternalAuthService) GrantObjectPermission(req struct {
	DatasourceID  string
	PrincipalType string
	PrincipalID   string
	ObjectType    string
	DatabaseName  string
	ObjectName    string
	Columns       []string
	Privileges    []string
}, operatorID, operatorName string) error {
	var principalName, username, host string
	
	if req.PrincipalType == datasource_auth.PrincipalTypeUser {
		user, err := s.GetInternalUser(req.PrincipalID)
		if err != nil {
			return err
		}
		principalName = user.Username
		username = user.Username
		host = user.Host
	} else {
		role, err := s.GetInternalRole(req.PrincipalID)
		if err != nil {
			return err
		}
		principalName = role.Name
		username = role.Name
		host = "%"
	}

	ds, err := GetDatasourceByID(req.DatasourceID)
	if err != nil {
		return err
	}

	if ds.DBType != model.DBTypeSQLite {
		params := &dbtype.ConnectionParams{
			Type:       ds.DBType,
			Host:       ds.Host,
			Port:       ds.Port,
			Username:   ds.Username,
			Password:   ds.Password,
			Charset:    ds.Charset,
			Timezone:   ds.Timezone,
			SSLMode:    ds.SSLMode,
			SSLCAFile:  ds.SSLCAFile,
		}

		var sqlDB *sql.DB
		switch ds.DBType {
		case model.DBTypeMySQL, model.DBTypeTiDB:
			dsn, _, _ := dbtype.BuildDSN(params)
			sqlDB, err = sql.Open("mysql", dsn)
		default:
			return fmt.Errorf("unsupported database type: %s", ds.DBType)
		}
		if err != nil {
			return fmt.Errorf("connect failed: %w", err)
		}
		defer sqlDB.Close()

		permAdapter := adapter.NewPermissionAdapter(ds.DBType)
		objType := adapter.ObjectType(req.ObjectType)
		if err := permAdapter.GrantObject(sqlDB, username, host, req.Privileges, objType, req.DatabaseName, req.ObjectName, req.Columns); err != nil {
			return err
		}
	}

	columnsJSON, _ := json.Marshal(req.Columns)
	privilegesJSON, _ := json.Marshal(req.Privileges)

	rule := &datasource_auth.DatasourcePermissionRule{
		ID:             s.generateID(),
		DatasourceID:   req.DatasourceID,
		PrincipalType:  req.PrincipalType,
		PrincipalID:    req.PrincipalID,
		PrivilegeType:  string(privilegesJSON),
		ObjectLevel:    req.ObjectType,
		DatabaseName:   req.DatabaseName,
		Table:          req.ObjectName,
		Columns:        string(columnsJSON),
		Enabled:        true,
		PrivilegeCategory: datasource_auth.PrivilegeCategoryObject,
		CreatedAt:      time.Now(),
	}

	if err := database.DB.Create(rule).Error; err != nil {
		return err
	}

	s.invalidateCache(req.DatasourceID)
	s.LogAudit(operatorID, operatorName, "grant_object_permission", principalName, req.DatasourceID, datasource_auth.AuditResultSuccess, fmt.Sprintf("objType=%s db=%s obj=%s privs=%v", req.ObjectType, req.DatabaseName, req.ObjectName, req.Privileges))
	return nil
}

func (s *InternalAuthService) RevokeObjectPermission(ruleID string, operatorID, operatorName string) error {
	var rule datasource_auth.DatasourcePermissionRule
	if err := database.DB.Where("id = ?", ruleID).First(&rule).Error; err != nil {
		return err
	}

	if rule.PrivilegeCategory != datasource_auth.PrivilegeCategoryObject {
		return fmt.Errorf("非对象权限规则")
	}

	ds, err := GetDatasourceByID(rule.DatasourceID)
	if err != nil {
		return err
	}

	var username, host string
	if rule.PrincipalType == datasource_auth.PrincipalTypeUser {
		user, err := s.GetInternalUser(rule.PrincipalID)
		if err != nil {
			return err
		}
		username = user.Username
		host = user.Host
	} else {
		role, err := s.GetInternalRole(rule.PrincipalID)
		if err != nil {
			return err
		}
		username = role.Name
		host = "%"
	}

	if ds.DBType != model.DBTypeSQLite {
		params := &dbtype.ConnectionParams{
			Type:       ds.DBType,
			Host:       ds.Host,
			Port:       ds.Port,
			Username:   ds.Username,
			Password:   ds.Password,
			Charset:    ds.Charset,
			Timezone:   ds.Timezone,
			SSLMode:    ds.SSLMode,
			SSLCAFile:  ds.SSLCAFile,
		}

		var sqlDB *sql.DB
		switch ds.DBType {
		case model.DBTypeMySQL, model.DBTypeTiDB:
			dsn, _, _ := dbtype.BuildDSN(params)
			sqlDB, err = sql.Open("mysql", dsn)
		default:
			return fmt.Errorf("unsupported database type: %s", ds.DBType)
		}
		if err != nil {
			return fmt.Errorf("connect failed: %w", err)
		}
		defer sqlDB.Close()

		var privs []string
		if rule.PrivilegeType != "" {
			json.Unmarshal([]byte(rule.PrivilegeType), &privs)
		}

		var cols []string
		if rule.Columns != "" {
			json.Unmarshal([]byte(rule.Columns), &cols)
		}

		permAdapter := adapter.NewPermissionAdapter(ds.DBType)
		objType := adapter.ObjectType(rule.ObjectLevel)
		permAdapter.RevokeObject(sqlDB, username, host, privs, objType, rule.DatabaseName, rule.Table, cols)
	}

	if err := database.DB.Delete(&rule).Error; err != nil {
		return err
	}

	s.invalidateCache(rule.DatasourceID)
	s.LogAudit(operatorID, operatorName, "revoke_object_permission", ruleID, rule.DatasourceID, datasource_auth.AuditResultSuccess, "")
	return nil
}

func (s *InternalAuthService) GetObjectPrivileges(datasourceID, objectType string) []string {
	ds, err := GetDatasourceByID(datasourceID)
	if err != nil {
		return []string{}
	}

	permAdapter := adapter.NewPermissionAdapter(ds.DBType)
	objType := adapter.ObjectType(objectType)
	return permAdapter.GetObjectPrivileges(objType)
}

func (s *InternalAuthService) ListDatabases(datasourceID string) ([]string, error) {
	ds, err := GetDatasourceByID(datasourceID)
	if err != nil {
		return nil, err
	}

	if ds.DBType == model.DBTypeSQLite {
		return []string{"main"}, nil
	}

	params := &dbtype.ConnectionParams{
		Type:       ds.DBType,
		Host:       ds.Host,
		Port:       ds.Port,
		Username:   ds.Username,
		Password:   ds.Password,
		Charset:    ds.Charset,
		Timezone:   ds.Timezone,
		SSLMode:    ds.SSLMode,
		SSLCAFile:  ds.SSLCAFile,
		Database:   ds.DefaultDB,
	}

	conn, err := dbtype.Connect(datasourceID, params)
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}

	permAdapter := adapter.NewPermissionAdapter(ds.DBType)
	return permAdapter.ListDatabases(conn.DB)
}

func (s *InternalAuthService) ListObjects(datasourceID, dbName, objectType string) ([]string, error) {
	ds, err := GetDatasourceByID(datasourceID)
	if err != nil {
		return nil, err
	}

	if ds.DBType == model.DBTypeSQLite {
		params := &dbtype.ConnectionParams{
			Type:       ds.DBType,
			Host:       ds.Host,
			Port:       ds.Port,
			Username:   ds.Username,
			Password:   ds.Password,
			Charset:    ds.Charset,
			Timezone:   ds.Timezone,
			SSLMode:    ds.SSLMode,
			SSLCAFile:  ds.SSLCAFile,
		}
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err := sql.Open("sqlite", dsn)
		if err != nil {
			return nil, err
		}
		defer sqlDB.Close()

		permAdapter := adapter.NewPermissionAdapter(ds.DBType)
		objType := adapter.ObjectType(objectType)
		switch objType {
		case adapter.ObjectTypeTable:
			return permAdapter.ListTables(sqlDB, dbName)
		case adapter.ObjectTypeView:
			return permAdapter.ListViews(sqlDB, dbName)
		case adapter.ObjectTypeColumn:
			return []string{}, nil
		default:
			return []string{}, nil
		}
	}

	params := &dbtype.ConnectionParams{
		Type:       ds.DBType,
		Host:       ds.Host,
		Port:       ds.Port,
		Username:   ds.Username,
		Password:   ds.Password,
		Charset:    ds.Charset,
		Timezone:   ds.Timezone,
		SSLMode:    ds.SSLMode,
		SSLCAFile:  ds.SSLCAFile,
	}

	var sqlDB *sql.DB
	switch ds.DBType {
	case model.DBTypeMySQL, model.DBTypeTiDB:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("mysql", dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", ds.DBType)
	}
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	defer sqlDB.Close()

	permAdapter := adapter.NewPermissionAdapter(ds.DBType)
	objType := adapter.ObjectType(objectType)
	switch objType {
	case adapter.ObjectTypeTable:
		return permAdapter.ListTables(sqlDB, dbName)
	case adapter.ObjectTypeView:
		return permAdapter.ListViews(sqlDB, dbName)
	case adapter.ObjectTypeProcedure:
		return permAdapter.ListProcedures(sqlDB, dbName)
	case adapter.ObjectTypeFunction:
		return permAdapter.ListFunctions(sqlDB, dbName)
	case adapter.ObjectTypeTrigger:
		return permAdapter.ListTriggers(sqlDB, dbName)
	case adapter.ObjectTypeEvent:
		return permAdapter.ListEvents(sqlDB, dbName)
	default:
		return []string{}, nil
	}
}

func (s *InternalAuthService) ListColumns(datasourceID, dbName, tableName string) ([]string, error) {
	ds, err := GetDatasourceByID(datasourceID)
	if err != nil {
		return nil, err
	}

	if ds.DBType == model.DBTypeSQLite {
		params := &dbtype.ConnectionParams{
			Type:       ds.DBType,
			Host:       ds.Host,
			Port:       ds.Port,
			Username:   ds.Username,
			Password:   ds.Password,
			Charset:    ds.Charset,
			Timezone:   ds.Timezone,
			SSLMode:    ds.SSLMode,
			SSLCAFile:  ds.SSLCAFile,
		}
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err := sql.Open("sqlite", dsn)
		if err != nil {
			return nil, err
		}
		defer sqlDB.Close()

		permAdapter := adapter.NewPermissionAdapter(ds.DBType)
		return permAdapter.ListColumns(sqlDB, dbName, tableName)
	}

	params := &dbtype.ConnectionParams{
		Type:       ds.DBType,
		Host:       ds.Host,
		Port:       ds.Port,
		Username:   ds.Username,
		Password:   ds.Password,
		Charset:    ds.Charset,
		Timezone:   ds.Timezone,
		SSLMode:    ds.SSLMode,
		SSLCAFile:  ds.SSLCAFile,
	}

	var sqlDB *sql.DB
	switch ds.DBType {
	case model.DBTypeMySQL, model.DBTypeTiDB:
		dsn, _, _ := dbtype.BuildDSN(params)
		sqlDB, err = sql.Open("mysql", dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", ds.DBType)
	}
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	defer sqlDB.Close()

	permAdapter := adapter.NewPermissionAdapter(ds.DBType)
	return permAdapter.ListColumns(sqlDB, dbName, tableName)
}