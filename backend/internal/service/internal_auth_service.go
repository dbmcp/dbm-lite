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

func (s *InternalAuthService) ListPermissionRules(datasourceID, principalType, principalID, privilegeType, objectLevel string, page, pageSize int) ([]datasource_auth.DatasourcePermissionRule, int64, error) {
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

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []datasource_auth.DatasourcePermissionRule
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
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
	hasDML := false
	hasDDL := false

	for _, rule := range rules {
		if rule.PrivilegeType == datasource_auth.PrivilegeTypeDDL {
			hasDDL = true
			hasDML = true
		} else if rule.PrivilegeType == datasource_auth.PrivilegeTypeDML {
			hasDML = true
		}
	}

	if strings.Contains(sqlLower, "insert") || strings.Contains(sqlLower, "update") || strings.Contains(sqlLower, "delete") {
		if !hasDML {
			return false, "无DML权限"
		}
	}

	if strings.Contains(sqlLower, "create") || strings.Contains(sqlLower, "alter") ||
		strings.Contains(sqlLower, "drop") || strings.Contains(sqlLower, "truncate") {
		if !hasDDL {
			return false, "无DDL权限"
		}
	}

	return true, ""
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

func (s *InternalAuthService) ListAuditLogs(datasourceID, operator, operType string, page, pageSize int) ([]datasource_auth.DatasourceAuthAudit, int64, error) {
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

func GetDatasourceByID(id string) (*model.Datasource, error) {
	var ds model.Datasource
	err := database.DB.Where("datasource_id = ?", id).First(&ds).Error
	return &ds, err
}