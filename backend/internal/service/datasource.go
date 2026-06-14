/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"errors"
	"strings"
	"time"

	"dbm-lite/config"
	"dbm-lite/internal/database"
	"dbm-lite/internal/model"
	"dbm-lite/pkg/crypto"

	"github.com/google/uuid"
)

type DatasourceService struct{}

func NewDatasourceService() *DatasourceService { return &DatasourceService{} }

func (s *DatasourceService) Create(ds *model.Datasource, rawPassword string, createdBy, createdByName string) error {
	if ds == nil {
		return errors.New("数据源信息为空")
	}
	if strings.TrimSpace(ds.Name) == "" {
		return errors.New("名称不能为空")
	}
	if !model.IsSupportedDBType(ds.DBType) {
		return errors.New("不支持的数据库类型")
	}

	if rawPassword != "" {
		encPwd, err := crypto.EncryptAES(rawPassword, config.App.AESKey)
		if err != nil {
			return err
		}
		ds.Password = encPwd
	}

	ds.DatasourceID = uuid.New().String()
	ds.CreatedBy = createdBy
	ds.CreatedAt = time.Now()
	ds.UpdatedAt = time.Now()
	if ds.Status == "" {
		ds.Status = model.StatusActive
	}
	if ds.Env == "" {
		ds.Env = model.EnvDev
	}
	if ds.ColorLabel == "" {
		ds.ColorLabel = model.ColorLabelBlue
	}
	if ds.DBType == model.DBTypeSQLite && ds.OpenMode == "" {
		ds.OpenMode = "rw"
	}
	if ds.Charset == "" {
		ds.Charset = "utf8mb4"
	}
	if ds.Timezone == "" {
		ds.Timezone = "Local"
	}
	if ds.SSLMode == "" {
		ds.SSLMode = "false"
	}
	return database.DB.Create(ds).Error
}

func (s *DatasourceService) List(page, pageSize int, keyword, dbType, status, sortBy, businessId, env string) ([]model.Datasource, int64, error) {
	var list []model.Datasource
	var total int64
	q := database.DB.Model(&model.Datasource{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR host LIKE ? OR remark LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if dbType != "" {
		q = q.Where("db_type = ?", dbType)
	}
	if status != "" {
		if status == "connected" || status == "ok" {
			q = q.Where("conn_status = ?", model.ConnStatusOK)
		} else if status == "failed" || status == "fail" {
			q = q.Where("conn_status = ?", model.ConnStatusFail)
		} else if status == "untested" {
			q = q.Where("conn_status IS NULL OR conn_status = ''")
		} else {
			q = q.Where("status = ?", status)
		}
	}
	if businessId != "" {
		q = q.Where("business_id = ?", businessId)
	}
	if env != "" {
		q = q.Where("env = ?", env)
	}

	q.Count(&total)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	orderStr := "created_at DESC"
	switch strings.ToLower(sortBy) {
	case "name":
		orderStr = "name ASC"
	case "namedesc":
		orderStr = "name DESC"
	case "lasttest":
		orderStr = "last_conn_test_at DESC, created_at DESC"
	case "recent":
		orderStr = "updated_at DESC, created_at DESC"
	case "latency":
		orderStr = "conn_latency_ms ASC, created_at DESC"
	}

	err := q.Order(orderStr).Offset(offset).Limit(pageSize).Omit("password").Find(&list).Error
	return list, total, err
}

func (s *DatasourceService) GetById(id string) (*model.Datasource, error) {
	var ds model.Datasource
	if err := database.DB.Where("datasource_id = ?", id).First(&ds).Error; err != nil {
		return nil, err
	}
	if ds.Password != "" {
		plain, err := crypto.DecryptAES(ds.Password, config.App.AESKey)
		if err == nil {
			ds.Password = plain
		}
	}
	return &ds, nil
}

func (s *DatasourceService) GetByIdNoDecrypt(id string) (*model.Datasource, error) {
	var ds model.Datasource
	if err := database.DB.Where("datasource_id = ?", id).First(&ds).Error; err != nil {
		return nil, err
	}
	return &ds, nil
}

func (s *DatasourceService) Update(id string, updates map[string]interface{}, rawPassword string) error {
	if rawPassword != "" {
		encPwd, err := crypto.EncryptAES(rawPassword, config.App.AESKey)
		if err != nil {
			return err
		}
		updates["password"] = encPwd
	}
	updates["updated_at"] = time.Now()
	return database.DB.Model(&model.Datasource{}).Where("datasource_id = ?", id).Updates(updates).Error
}

func (s *DatasourceService) Delete(id string) error {
	return database.DB.Where("datasource_id = ?", id).Delete(&model.Datasource{}).Error
}

func (s *DatasourceService) Copy(id string, createdBy, createdByName string) (*model.Datasource, error) {
	ds, err := s.GetById(id)
	if err != nil {
		return nil, err
	}

	newDs := &model.Datasource{
		Name:       ds.Name + " 副本",
		DBType:     ds.DBType,
		Host:       ds.Host,
		Port:       ds.Port,
		Username:   ds.Username,
		DefaultDB:  ds.DefaultDB,
		FilePath:   ds.FilePath,
		OpenMode:   ds.OpenMode,
		Charset:    ds.Charset,
		Timezone:   ds.Timezone,
		SSLMode:    ds.SSLMode,
		SSLCAFile:  ds.SSLCAFile,
		ReadOnly:   ds.ReadOnly,
		ColorLabel: ds.ColorLabel,
		Tags:       ds.Tags,
		BusinessID: ds.BusinessID,
		ServerID:   ds.ServerID,
		ProjectID:  ds.ProjectID,
		Env:        ds.Env,
		Remark:     ds.Remark,
		Status:     ds.Status,
		Version:    "",
		ConnStatus: "",
	}

	rawPwd := ds.Password
	if err := s.Create(newDs, rawPwd, createdBy, createdByName); err != nil {
		return nil, err
	}
	return newDs, nil
}

func (s *DatasourceService) Stats() (map[string]interface{}, error) {
	var total int64
	database.DB.Model(&model.Datasource{}).Count(&total)

	typeResult := make(map[string]int)
	rows, err := database.DB.Model(&model.Datasource{}).Select("db_type, COUNT(*) as count").Group("db_type").Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t string
			var c int64
			rows.Scan(&t, &c)
			typeResult[strings.ToLower(t)] = int(c)
		}
	}

	envResult := make(map[string]int)
	rows2, err2 := database.DB.Model(&model.Datasource{}).Select("env, COUNT(*) as count").Group("env").Rows()
	if err2 == nil {
		defer rows2.Close()
		for rows2.Next() {
			var e string
			var c int64
			rows2.Scan(&e, &c)
			envResult[e] = int(c)
		}
	}

	var okCount, failCount int64
	database.DB.Model(&model.Datasource{}).Where("conn_status = ?", model.ConnStatusOK).Count(&okCount)
	database.DB.Model(&model.Datasource{}).Where("conn_status = ?", model.ConnStatusFail).Count(&failCount)

	return map[string]interface{}{
		"total":    total,
		"byType":   typeResult,
		"byEnv":    envResult,
		"ok":       okCount,
		"failed":   failCount,
		"untested": total - okCount - failCount,
	}, nil
}

func (s *DatasourceService) UpdateConnStatus(id, status string, latencyMs int64, version string) {
	now := time.Now()
	updates := map[string]interface{}{
		"conn_status":       status,
		"last_conn_test_at": &now,
		"updated_at":        now,
	}
	if latencyMs > 0 {
		updates["conn_latency_ms"] = latencyMs
	}
	if version != "" {
		updates["version"] = version
	}
	database.DB.Model(&model.Datasource{}).Where("datasource_id = ?", id).Updates(updates)
}

func (s *DatasourceService) UpdateConnStatusFail(id, message string) {
	now := time.Now()
	database.DB.Model(&model.Datasource{}).Where("datasource_id = ?", id).Updates(map[string]interface{}{
		"conn_status":       model.ConnStatusFail,
		"last_conn_test_at": &now,
		"updated_at":        now,
		"remark":            message,
	})
}

func (s *DatasourceService) ListByBusiness(businessId string) ([]model.Datasource, error) {
	var list []model.Datasource
	err := database.DB.Where("business_id = ?", businessId).Omit("password").Find(&list).Error
	return list, err
}
