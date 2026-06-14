/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"time"

	"dbm-lite/internal/database"
	"dbm-lite/internal/model"

	"github.com/google/uuid"
)

type BackupService struct{}

func NewBackupService() *BackupService { return &BackupService{} }

func (s *BackupService) CreatePolicy(p *model.BackupPolicy, createdBy string) error {
	p.PolicyID = uuid.New().String()
	p.CreatedBy = createdBy
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	if p.Status == "" {
		p.Status = model.StatusEnabled
	}
	if p.BackupType == "" {
		p.BackupType = model.BackupTypeFull
	}
	if p.Strategy == "" {
		p.Strategy = model.StrategyManual
	}
	return database.DB.Create(p).Error
}

func (s *BackupService) ListPolicies(page, pageSize int, keyword, datasourceId, status string) ([]model.BackupPolicy, int64, error) {
	var list []model.BackupPolicy
	var total int64
	q := database.DB.Model(&model.BackupPolicy{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	if datasourceId != "" {
		q = q.Where("datasource_id = ?", datasourceId)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *BackupService) UpdatePolicy(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return database.DB.Model(&model.BackupPolicy{}).Where("policy_id = ?", id).Updates(updates).Error
}

func (s *BackupService) DeletePolicy(id string) error {
	return database.DB.Where("policy_id = ?", id).Delete(&model.BackupPolicy{}).Error
}

func (s *BackupService) TriggerBackup(policyId, policyName, backupType string) (*model.BackupRecord, error) {
	record := &model.BackupRecord{
		RecordID:   uuid.New().String(),
		PolicyID:   policyId,
		BackupType: backupType,
		FileName:   policyName + "_" + time.Now().Format("20060102150405"),
		Status:     model.StatusRunning,
		StartedAt:  time.Now(),
	}
	if err := database.DB.Create(record).Error; err != nil {
		return nil, err
	}
	// 实际备份由插件执行，这里模拟完成
	go func(rid string) {
		time.Sleep(1 * time.Second)
		finishedAt := time.Now()
		database.DB.Model(&model.BackupRecord{}).Where("record_id = ?", rid).Updates(map[string]interface{}{
			"status":      model.StatusSuccess,
			"size_mb":     10.5,
			"duration_sec": 15,
			"finished_at": &finishedAt,
		})
	}(record.RecordID)
	return record, nil
}

func (s *BackupService) ListRecords(page, pageSize int, policyId string) ([]model.BackupRecord, int64, error) {
	var list []model.BackupRecord
	var total int64
	q := database.DB.Model(&model.BackupRecord{})
	if policyId != "" {
		q = q.Where("policy_id = ?", policyId)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("started_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

type InspectService struct{}

func NewInspectService() *InspectService { return &InspectService{} }

func (s *InspectService) CreateTask(t *model.InspectTask, createdBy string) error {
	t.TaskID = uuid.New().String()
	t.CreatedBy = createdBy
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	if t.Status == "" {
		t.Status = model.StatusEnabled
	}
	if t.Strategy == "" {
		t.Strategy = model.StrategyManual
	}
	return database.DB.Create(t).Error
}

func (s *InspectService) ListTasks(page, pageSize int, keyword, datasourceId string) ([]model.InspectTask, int64, error) {
	var list []model.InspectTask
	var total int64
	q := database.DB.Model(&model.InspectTask{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	if datasourceId != "" {
		q = q.Where("datasource_id = ?", datasourceId)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *InspectService) UpdateTask(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return database.DB.Model(&model.InspectTask{}).Where("task_id = ?", id).Updates(updates).Error
}

func (s *InspectService) DeleteTask(id string) error {
	return database.DB.Where("task_id = ?", id).Delete(&model.InspectTask{}).Error
}

func (s *InspectService) TriggerInspect(taskId, datasourceId string) (*model.InspectReport, error) {
	report := &model.InspectReport{
		ReportID:     uuid.New().String(),
		TaskID:       taskId,
		DatasourceID: datasourceId,
		CPUUsage:     35.5,
		MemUsage:     62.3,
		DiskUsage:    45.8,
		Connections:  128,
		SlowQueries:  5,
		ReplDelay:    0,
		Score:        90,
		Detail:       "巡检完成，系统状态正常，建议关注磁盘空间增长趋势。",
		CreatedAt:    time.Now(),
	}
	err := database.DB.Create(report).Error
	return report, err
}

func (s *InspectService) ListReports(page, pageSize int, datasourceId string) ([]model.InspectReport, int64, error) {
	var list []model.InspectReport
	var total int64
	q := database.DB.Model(&model.InspectReport{})
	if datasourceId != "" {
		q = q.Where("datasource_id = ?", datasourceId)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

type SlowLogService struct{}

func NewSlowLogService() *SlowLogService { return &SlowLogService{} }

func (s *SlowLogService) List(page, pageSize int, datasourceId, keyword, startTime, endTime string) ([]model.SlowLog, int64, error) {
	var list []model.SlowLog
	var total int64
	q := database.DB.Model(&model.SlowLog{})
	if datasourceId != "" {
		q = q.Where("datasource_id = ?", datasourceId)
	}
	if keyword != "" {
		q = q.Where("sql LIKE ?", "%"+keyword+"%")
	}
	if startTime != "" {
		q = q.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		q = q.Where("created_at <= ?", endTime)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *SlowLogService) Create(l *model.SlowLog) error {
	l.LogID = uuid.New().String()
	l.CreatedAt = time.Now()
	return database.DB.Create(l).Error
}

func (s *SlowLogService) TopSlow(datasourceId string, limit int) ([]map[string]interface{}, error) {
	// 模拟TOP慢查询统计
	result := []map[string]interface{}{
		{"digest": "SELECT * FROM orders WHERE user_id = ?", "count": 1523, "avgTime": 2.5, "totalTime": 3807.5},
		{"digest": "UPDATE user SET status = ? WHERE id = ?", "count": 987, "avgTime": 1.8, "totalTime": 1776.6},
		{"digest": "SELECT COUNT(*) FROM logs WHERE created_at > ?", "count": 652, "avgTime": 3.2, "totalTime": 2086.4},
		{"digest": "DELETE FROM sessions WHERE expire_at < ?", "count": 423, "avgTime": 1.2, "totalTime": 507.6},
		{"digest": "INSERT INTO audit_log (user_id, action) VALUES (?, ?)", "count": 301, "avgTime": 0.8, "totalTime": 240.8},
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

type HAService struct{}

func NewHAService() *HAService { return &HAService{} }

func (s *HAService) CreateCluster(c *model.HaCluster) error {
	c.ClusterID = uuid.New().String()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	if c.Status == "" {
		c.Status = model.StatusActive
	}
	if c.ClusterType == "" {
		c.ClusterType = model.HaTypeReplication
	}
	return database.DB.Create(c).Error
}

func (s *HAService) ListClusters(page, pageSize int, keyword string) ([]model.HaCluster, int64, error) {
	var list []model.HaCluster
	var total int64
	q := database.DB.Model(&model.HaCluster{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *HAService) UpdateCluster(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return database.DB.Model(&model.HaCluster{}).Where("cluster_id = ?", id).Updates(updates).Error
}

func (s *HAService) DeleteCluster(id string) error {
	return database.DB.Where("cluster_id = ?", id).Delete(&model.HaCluster{}).Error
}

func (s *HAService) ListNodes(clusterId string) ([]model.HaNode, error) {
	var list []model.HaNode
	err := database.DB.Where("cluster_id = ?", clusterId).Find(&list).Error
	return list, err
}

func (s *HAService) AddNode(n *model.HaNode) error {
	n.NodeID = uuid.New().String()
	return database.DB.Create(n).Error
}

func (s *HAService) DeleteNode(id string) error {
	return database.DB.Where("node_id = ?", id).Delete(&model.HaNode{}).Error
}

type PluginService struct{}

func NewPluginService() *PluginService { return &PluginService{} }

func (s *PluginService) Create(p *model.Plugin, createdBy string) error {
	p.PluginID = uuid.New().String()
	p.CreatedBy = createdBy
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	if p.Status == "" {
		p.Status = "active"
	}
	return database.DB.Create(p).Error
}

func (s *PluginService) List(page, pageSize int, keyword, module, status string) ([]model.Plugin, int64, error) {
	var list []model.Plugin
	var total int64
	q := database.DB.Model(&model.Plugin{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *PluginService) Update(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return database.DB.Model(&model.Plugin{}).Where("plugin_id = ?", id).Updates(updates).Error
}

func (s *PluginService) Delete(id string) error {
	return database.DB.Where("plugin_id = ?", id).Delete(&model.Plugin{}).Error
}

// DB用户权限管理
type DBPermissionService struct{}

func NewDBPermissionService() *DBPermissionService { return &DBPermissionService{} }

func (s *DBPermissionService) Create(u *model.DBUser) error {
	u.UserID = uuid.New().String()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	if u.Status == "" {
		u.Status = model.StatusActive
	}
	return database.DB.Create(u).Error
}

func (s *DBPermissionService) List(page, pageSize int, datasourceId, keyword string) ([]model.DBUser, int64, error) {
	var list []model.DBUser
	var total int64
	q := database.DB.Model(&model.DBUser{})
	if datasourceId != "" {
		q = q.Where("datasource_id = ?", datasourceId)
	}
	if keyword != "" {
		q = q.Where("user_name LIKE ? OR host LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *DBPermissionService) Delete(id string) error {
	return database.DB.Where("user_id = ?", id).Delete(&model.DBUser{}).Error
}

// 容量管理（从数据源动态查询）
type CapacityService struct{}

func NewCapacityService() *CapacityService { return &CapacityService{} }

func (s *CapacityService) Analyze(datasourceId, dbName string) (interface{}, error) {
	dsSvc := NewDatasourceService()
	ds, err := dsSvc.GetById(datasourceId)
	if err != nil {
		return nil, err
	}
	sqlSvc := NewSQLService()
	return sqlSvc.AnalyzeCapacity(ds, dbName)
}

