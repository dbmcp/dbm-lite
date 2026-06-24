/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"time"

	"dbm-lite/internal/database"
	"dbm-lite/internal/model"

	"github.com/google/uuid"
)

// ====== 权限申请与审批 ======

const (
	privApplyPending  = "pending"
	privApplyApproved = "approved"
	privApplyRejected = "rejected"
	privApplyExpired  = "expired"
)

type PrivApplyService struct{}

func NewPrivApplyService() *PrivApplyService { return &PrivApplyService{} }

func (s *PrivApplyService) List(page, pageSize int, status, applicantID, keyword string) ([]model.QueryPrivApply, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.QueryPrivApply
	var total int64
	query := database.DB.Model(&model.QueryPrivApply{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if applicantID != "" {
		query = query.Where("applicant_id = ?", applicantID)
	}
	if keyword != "" {
		query = query.Where("database_name LIKE ? OR table_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (s *PrivApplyService) Get(id string) (*model.QueryPrivApply, error) {
	var a model.QueryPrivApply
	err := database.DB.Where("apply_id = ?", id).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Submit 提交权限申请
func (s *PrivApplyService) Submit(a *model.QueryPrivApply, userID, userName string, validDays, rowLimit int) error {
	a.ApplyID = uuid.New().String()
	a.ApplicantID = userID
	a.ApplicantName = userName
	a.ValidDays = validDays
	a.RowLimit = rowLimit
	if a.PrivType == "" {
		a.PrivType = "table"
	}
	a.Status = privApplyPending
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	return database.DB.Create(a).Error
}

// Approve 审批通过，事务化生成有效权限记录
func (s *PrivApplyService) Approve(applyID, approverID, approverName, remark string) (*model.QueryPrivApply, error) {
	tx := database.DB.Begin()
	var a model.QueryPrivApply
	if err := tx.Where("apply_id = ? AND status = ?", applyID, privApplyPending).First(&a).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	a.Status = privApplyApproved
	a.ApproverID = approverID
	a.ApproverName = approverName
	a.ApprovalRemark = remark
	a.UpdatedAt = time.Now()
	if err := tx.Model(&a).Updates(map[string]interface{}{
		"status":          a.Status,
		"approver_id":     a.ApproverID,
		"approver_name":   a.ApproverName,
		"approval_remark": a.ApprovalRemark,
		"updated_at":      a.UpdatedAt,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	expireAt := time.Now().AddDate(0, 0, a.ValidDays)
	priv := &model.QueryPrivilege{
		PrivID:        uuid.New().String(),
		UserID:        a.ApplicantID,
		DatasourceID:  a.DatasourceID,
		DatabaseName:  a.DatabaseName,
		TblName:       a.TblName,
		PrivType:      a.PrivType,
		OperationType: a.OperationType,
		Columns:       a.Columns,
		RowLimit:      a.RowLimit,
		ApplyID:       a.ApplyID,
		ExpireAt:      expireAt,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsExpired:     false,
	}
	if priv.OperationType == "" {
		priv.OperationType = "select"
	}
	if err := tx.Create(priv).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// Reject 驳回申请
func (s *PrivApplyService) Reject(applyID, approverID, approverName, remark string) (*model.QueryPrivApply, error) {
	var a model.QueryPrivApply
	err := database.DB.Where("apply_id = ? AND status = ?", applyID, privApplyPending).First(&a).Error
	if err != nil {
		return nil, err
	}
	a.Status = privApplyRejected
	a.ApproverID = approverID
	a.ApproverName = approverName
	a.ApprovalRemark = remark
	a.UpdatedAt = time.Now()
	err = database.DB.Model(&a).Updates(map[string]interface{}{
		"status":          a.Status,
		"approver_id":     a.ApproverID,
		"approver_name":   a.ApproverName,
		"approval_remark": a.ApprovalRemark,
		"updated_at":      a.UpdatedAt,
	}).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Revoke 主动回收权限
func (s *PrivApplyService) Revoke(privID string) error {
	return database.DB.Model(&model.QueryPrivilege{}).
		Where("priv_id = ?", privID).
		Updates(map[string]interface{}{"is_expired": true, "updated_at": time.Now()}).Error
}

// RevokeByIDs 批量回收权限
func (s *PrivApplyService) RevokeByIDs(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return database.DB.Model(&model.QueryPrivilege{}).
		Where("priv_id IN ?", ids).
		Updates(map[string]interface{}{"is_expired": true, "updated_at": time.Now()}).Error
}

// Grant 直接授权（不走申请工单流程）
func (s *PrivApplyService) Grant(userID, userName, datasourceID, dbName, tableName, privType, operationType, columnsStr string, rowLimit, validDays int) (*model.QueryPrivilege, error) {
	if validDays <= 0 {
		validDays = 365 * 100 // 默认长期有效
	}
	if rowLimit < 0 {
		rowLimit = 0
	}
	if operationType == "" {
		operationType = "select"
	}
	if privType == "" {
		privType = "table"
	}
	priv := &model.QueryPrivilege{
		PrivID:        uuid.New().String(),
		UserID:        userID,
		DatasourceID:  datasourceID,
		DatabaseName:  dbName,
		TblName:       tableName,
		PrivType:      privType,
		OperationType: operationType,
		Columns:       columnsStr,
		RowLimit:      rowLimit,
		ApplyID:       "direct:" + userName,
		ExpireAt:      time.Now().AddDate(0, 0, validDays),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsExpired:     false,
	}
	if err := database.DB.Create(priv).Error; err != nil {
		return nil, err
	}
	logPrivAudit(userID, userName, "priv.grant", "", priv.PrivID,
		"datasource:"+datasourceID+" db:"+dbName+" table:"+tableName, "", "granted")
	return priv, nil
}

// ListAllEffectivePrivileges 列出所有有效权限（带筛选）
func (s *PrivApplyService) ListAllEffectivePrivileges(page, pageSize int, userID, datasourceID, dbName, tableName string, onlyExpired bool) ([]model.QueryPrivilege, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.QueryPrivilege
	var total int64
	query := database.DB.Model(&model.QueryPrivilege{})
	if onlyExpired {
		query = query.Where("is_expired = ? OR expire_at <= ?", true, time.Now())
	} else {
		query = query.Where("is_expired = ? AND expire_at > ?", false, time.Now())
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if datasourceID != "" {
		query = query.Where("datasource_id = ?", datasourceID)
	}
	if dbName != "" {
		query = query.Where("database_name LIKE ?", "%"+dbName+"%")
	}
	if tableName != "" {
		query = query.Where("table_name LIKE ?", "%"+tableName+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// CleanupExpired 定时清理过期权限
func (s *PrivApplyService) CleanupExpired() (int64, error) {
	tx := database.DB.Begin()
	now := time.Now()
	res1 := tx.Model(&model.QueryPrivilege{}).
		Where("is_expired = ? AND expire_at <= ?", false, now).
		Updates(map[string]interface{}{"is_expired": true, "updated_at": now})
	if res1.Error != nil {
		tx.Rollback()
		return 0, res1.Error
	}
	updated := res1.RowsAffected
	res2 := tx.Model(&model.QueryPrivApply{}).
		Where("status = ?", privApplyApproved).
		Where("apply_id IN (SELECT apply_id FROM query_privileges WHERE is_expired = ?)", true).
		Updates(map[string]interface{}{"status": privApplyExpired, "updated_at": now})
	if res2.Error != nil {
		tx.Rollback()
		return 0, res2.Error
	}
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	return updated, nil
}

// GetUserEffectivePrivileges 读取用户在指定数据源下所有有效权限
func (s *PrivApplyService) GetUserEffectivePrivileges(userID, datasourceID string) ([]model.QueryPrivilege, error) {
	var list []model.QueryPrivilege
	err := database.DB.Where("user_id = ? AND datasource_id = ? AND is_expired = ?", userID, datasourceID, false).
		Where("expire_at > ?", time.Now()).
		Find(&list).Error
	return list, err
}

// ====== 敏感列管理 ======

type SensitiveColumnService struct{}

func NewSensitiveColumnService() *SensitiveColumnService { return &SensitiveColumnService{} }

func (s *SensitiveColumnService) List(datasourceID string) ([]model.SensitiveColumn, error) {
	var list []model.SensitiveColumn
	query := database.DB.Model(&model.SensitiveColumn{})
	if datasourceID != "" {
		query = query.Where("datasource_id = ?", datasourceID)
	}
	err := query.Order("database_name, table_name, column_name").Find(&list).Error
	return list, err
}

func (s *SensitiveColumnService) Create(sc *model.SensitiveColumn) error {
	sc.CreatedAt = time.Now()
	sc.UpdatedAt = time.Now()
	return database.DB.Create(sc).Error
}

func (s *SensitiveColumnService) Delete(id int64) error {
	return database.DB.Where("id = ?", id).Delete(&model.SensitiveColumn{}).Error
}

// logPrivAudit 记录权限审计日志
func logPrivAudit(operatorID, operator, operType, applyID, targetID, detail, before, after string) {
	audit := model.PrivAuditLog{
		LogID:    uuid.NewString(),
		OperatorID: operatorID,
		Operator:   operator,
		OperType:   operType,
		ApplyID:    applyID,
		TargetID:   targetID,
		Detail:     detail,
		Before:     before,
		After:      after,
		CreatedAt:  time.Now(),
	}
	_ = database.DB.Create(&audit).Error
}
