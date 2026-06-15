/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"errors"
	"strings"
	"time"

	"dbm-lite/config"
	"dbm-lite/internal/database"
	"dbm-lite/internal/dbtype"
	"dbm-lite/internal/model"
	"dbm-lite/pkg/crypto"

	"github.com/google/uuid"
)

var ErrDatasourceNotFound = errors.New("数据源不存在")

type DatasourceService struct{}

func NewDatasourceService() *DatasourceService { return &DatasourceService{} }

// ==================== 矩阵视图 ====================

type EnvGroup struct {
	Env            string                  `json:"env"`
	EnvName        string                  `json:"envName"`
	DatasourceList []*DatasourceMatrixItem `json:"datasourceList"`
}

type DatasourceMatrixItem struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	DBType         string   `json:"dbType"`
	Type           string   `json:"type"`
	DatasourceType string   `json:"datasourceType"`
	Env            string   `json:"env"`
	Status         string   `json:"status"`
	ConnectStatus  string   `json:"connectStatus"`
	Description    string   `json:"description"`
	Tags           []string `json:"tags"`
	CreateTime     string   `json:"createTime"`
	IconType       string   `json:"iconType"`
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	ColorLabel     string   `json:"colorLabel"`
}

func (s *DatasourceService) GetMatrix() ([]*EnvGroup, error) {
	var list []model.Datasource
	err := database.DB.Order("updated_at DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}

	envNames := map[string]string{
		"dev":     "开发环境",
		"test":    "测试环境",
		"staging": "预发环境",
		"prod":    "生产环境",
	}
	envList := []string{"dev", "test", "staging", "prod"}
	result := make([]*EnvGroup, 0, 4)

	for _, env := range envList {
		group := &EnvGroup{
			Env:            env,
			EnvName:        envNames[env],
			DatasourceList: []*DatasourceMatrixItem{},
		}

		for _, ds := range list {
			if ds.Env == env {
				iconType := "self-hosted"
				if strings.ToLower(ds.DBType) == "sqlite" {
					iconType = "local"
				}

				tags := []string{}
				if ds.Tags != "" {
					tags = strings.Split(ds.Tags, ",")
				}

				item := &DatasourceMatrixItem{
					ID:             ds.DatasourceID,
					Name:           ds.Name,
					DBType:         ds.DBType,
					Type:           ds.DBType,
					DatasourceType: ds.DBType,
					Env:            ds.Env,
					Status:         ds.Status,
					ConnectStatus:  ds.ConnStatus,
					Description:    ds.Remark,
					Tags:           tags,
					CreateTime:     ds.CreatedAt.Format("2006-01-02 15:04:05"),
					IconType:       iconType,
					Host:           ds.Host,
					Port:           ds.Port,
					ColorLabel:     ds.ColorLabel,
				}
				group.DatasourceList = append(group.DatasourceList, item)
			}
		}

		if len(group.DatasourceList) > 6 {
			group.DatasourceList = group.DatasourceList[:6]
		}
		result = append(result, group)
	}
	return result, nil
}

// ==================== 密码脱敏 ====================

func maskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + "***" + password[len(password)-2:]
}

// ==================== V2 列表响应模型 ====================

type DatasourceListItem struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	DatasourceType string   `json:"datasourceType"`
	Type           string   `json:"type"`
	DBType         string   `json:"dbType"`
	Env            string   `json:"env"`
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	DatabaseName   string   `json:"databaseName"`
	Description    string   `json:"description"`
	Status         string   `json:"status"`
	ConnectStatus  string   `json:"connectStatus"`
	CreateTime     string   `json:"createTime"`
	UpdateTime     string   `json:"updateTime"`
	Tags           []string `json:"tags"`
	ColorLabel     string   `json:"colorLabel"`
	ConnLatencyMs int64    `json:"connLatencyMs"`
	FilePath       string   `json:"filePath"`
	OpenMode       string   `json:"openMode"`
}

type DatasourceDetail struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	DatasourceType string   `json:"datasourceType"`
	Type           string   `json:"type"`
	DBType         string   `json:"dbType"`
	Env            string   `json:"env"`
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	DatabaseName   string   `json:"databaseName"`
	Description    string   `json:"description"`
	Status         string   `json:"status"`
	ConnectStatus  string   `json:"connectStatus"`
	CreateTime     string   `json:"createTime"`
	UpdateTime     string   `json:"updateTime"`
	Tags           []string `json:"tags"`
	OwnerId        string   `json:"ownerId"`
	OrgId          string   `json:"orgId"`
	ColorLabel     string   `json:"colorLabel"`
	FilePath       string   `json:"filePath"`
	OpenMode       string   `json:"openMode"`
}

type CreateDatasourceReq struct {
	Name           string   `json:"name"`
	DatasourceType string   `json:"datasourceType"`
	Type           string   `json:"type"`
	DBType         string   `json:"dbType"`
	Env            string   `json:"env"`
	Host           string   `json:"host"`
	Port           *int     `json:"port"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	DatabaseName   string   `json:"databaseName"`
	Description    string   `json:"description"`
	Tags           []string `json:"tags"`
	FilePath       string   `json:"filePath"`
	OpenMode       string   `json:"openMode"`
	ColorLabel     string   `json:"colorLabel"`
}

type UpdateDatasourceReq struct {
	Name         string   `json:"name"`
	Env          string   `json:"env"`
	Host         string   `json:"host"`
	Port         *int     `json:"port"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	DatabaseName string   `json:"databaseName"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	FilePath     string   `json:"filePath"`
	OpenMode     string   `json:"openMode"`
	ColorLabel   string   `json:"colorLabel"`
	Status       string   `json:"status"`
}

type TestConnectionReq struct {
	DBType       string `json:"dbType"`
	Host           string `json:"host"`
	Port           *int   `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	DatabaseName string `json:"databaseName"`
	FilePath       string `json:"filePath"`
}

type TestConnectionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Version string `json:"version"`
	Cost    int64  `json:"cost"`
}

// ==================== V2 业务方法 ====================

func (s *DatasourceService) ListDatasource(keyword, dbType string, current, pageSize int) ([]*DatasourceListItem, int64, error) {
	var list []model.Datasource
	q := database.DB.Model(&model.Datasource{})

	if keyword != "" {
		q = q.Where("name LIKE ? OR host LIKE ? OR remark LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if dbType != "" {
		q = q.Where("db_type = ?", dbType)
	}

	var total int64
	q.Count(&total)

	if pageSize > 100 {
		pageSize = 100
	}
	if current < 1 {
		current = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (current - 1) * pageSize
	err := q.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	resultList := make([]*DatasourceListItem, 0, len(list))
	for _, ds := range list {
		tags := []string{}
		if ds.Tags != "" {
			tags = strings.Split(ds.Tags, ",")
		}

		resultList = append(resultList, &DatasourceListItem{
			ID:             ds.DatasourceID,
			Name:           ds.Name,
			DatasourceType: ds.DBType,
			Type:           ds.DBType,
			DBType:         ds.DBType,
			Env:            ds.Env,
			Host:           ds.Host,
			Port:           ds.Port,
			Username:       ds.Username,
			Password:       maskPassword(ds.Password),
			DatabaseName:   ds.DefaultDB,
			Description:    ds.Remark,
			Status:         ds.Status,
			ConnectStatus:  ds.ConnStatus,
			CreateTime:     ds.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdateTime:     ds.UpdatedAt.Format("2006-01-02 15:04:05"),
			Tags:           tags,
			ColorLabel:     ds.ColorLabel,
			ConnLatencyMs:  ds.ConnLatencyMs,
			FilePath:       ds.FilePath,
			OpenMode:       ds.OpenMode,
		})
	}

	return resultList, total, nil
}

func (s *DatasourceService) GetDatasourceInfo(id string) (*DatasourceDetail, error) {
	var ds model.Datasource
	err := database.DB.Where("datasource_id = ?", id).First(&ds).Error
	if err != nil {
		return nil, ErrDatasourceNotFound
	}

	tags := []string{}
	if ds.Tags != "" {
		tags = strings.Split(ds.Tags, ",")
	}

	return &DatasourceDetail{
		ID:             ds.DatasourceID,
		Name:           ds.Name,
		DatasourceType: ds.DBType,
		Type:           ds.DBType,
		DBType:         ds.DBType,
		Env:            ds.Env,
		Host:           ds.Host,
		Port:           ds.Port,
		Username:       ds.Username,
		Password:       maskPassword(ds.Password),
		DatabaseName:   ds.DefaultDB,
		Description:    ds.Remark,
		Status:         ds.Status,
		ConnectStatus:  ds.ConnStatus,
		CreateTime:     ds.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdateTime:     ds.UpdatedAt.Format("2006-01-02 15:04:05"),
		Tags:           tags,
		OwnerId:        ds.CreatedBy,
		OrgId:          "",
		ColorLabel:     ds.ColorLabel,
		FilePath:       ds.FilePath,
		OpenMode:       ds.OpenMode,
	}, nil
}

func (s *DatasourceService) CreateDatasource(req *CreateDatasourceReq, userId string) (*DatasourceDetail, error) {
	if req.Name == "" {
		return nil, errors.New("名称不能为空")
	}
	if len(req.Name) > 128 {
		return nil, errors.New("名称长度不能超过128个字符")
	}

	dbType := strings.ToLower(req.DBType)
	if dbType == "" {
		dbType = strings.ToLower(req.Type)
	}
	if dbType == "" {
		dbType = strings.ToLower(req.DatasourceType)
	}
	if !model.IsSupportedDBType(dbType) {
		return nil, errors.New("不支持的数据库类型")
	}
	if req.Env == "" {
		return nil, errors.New("环境不能为空")
	}

	var count int64
	database.DB.Model(&model.Datasource{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		return nil, errors.New("数据源名称已存在")
	}

	ds := &model.Datasource{
		Name:       req.Name,
		DBType:     dbType,
		Env:        req.Env,
		Host:       req.Host,
		Username:   req.Username,
		DefaultDB:  req.DatabaseName,
		Remark:     req.Description,
		Status:     model.StatusActive,
		ConnStatus: "unknown",
		FilePath:   req.FilePath,
		OpenMode:   strings.ToLower(req.OpenMode),
		ColorLabel: req.ColorLabel,
	}

	if ds.ColorLabel == "" {
		ds.ColorLabel = "blue"
	}
	if dbType == "sqlite" && ds.OpenMode == "" {
		ds.OpenMode = "rw"
	}

	if req.Port != nil && *req.Port > 0 {
		ds.Port = *req.Port
	} else {
		switch dbType {
		case "mysql":
			ds.Port = 3306
		case "tidb":
			ds.Port = 4000
		}
	}

	if req.Tags != nil {
		ds.Tags = strings.Join(req.Tags, ",")
	}

	ds.DatasourceID = uuid.New().String()
	ds.CreatedAt = time.Now()
	ds.UpdatedAt = time.Now()
	ds.CreatedBy = userId

	if req.Password != "" {
		encPwd, err := crypto.EncryptAES(req.Password, config.App.AESKey)
		if err != nil {
			return nil, err
		}
		ds.Password = encPwd
	}

	err := database.DB.Create(ds).Error
	if err != nil {
		return nil, err
	}

	return s.GetDatasourceInfo(ds.DatasourceID)
}

func (s *DatasourceService) UpdateDatasource(id string, req *UpdateDatasourceReq) (*DatasourceDetail, error) {
	var ds model.Datasource
	err := database.DB.Where("datasource_id = ?", id).First(&ds).Error
	if err != nil {
		return nil, ErrDatasourceNotFound
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Env != "" {
		updates["env"] = req.Env
	}
	if req.Host != "" {
		updates["host"] = req.Host
	}
	if req.Port != nil && *req.Port > 0 {
		updates["port"] = *req.Port
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		encPwd, err := crypto.EncryptAES(req.Password, config.App.AESKey)
		if err != nil {
			return nil, err
		}
		updates["password"] = encPwd
	}
	if req.DatabaseName != "" {
		updates["default_db"] = req.DatabaseName
	}
	if req.Description != "" {
		updates["remark"] = req.Description
	}
	if req.Tags != nil {
		updates["tags"] = strings.Join(req.Tags, ",")
	}
	if req.FilePath != "" {
		updates["file_path"] = req.FilePath
	}
	if req.OpenMode != "" {
		updates["open_mode"] = strings.ToLower(req.OpenMode)
	}
	if req.ColorLabel != "" {
		updates["color_label"] = req.ColorLabel
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	err = database.DB.Model(&model.Datasource{}).Where("datasource_id = ?", id).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	return s.GetDatasourceInfo(id)
}

func (s *DatasourceService) DeleteDatasource(id string) error {
	var ds model.Datasource
	err := database.DB.Where("datasource_id = ?", id).First(&ds).Error
	if err != nil {
		return ErrDatasourceNotFound
	}

	database.DB.Where("datasource_id = ?", id).Delete(&model.SQLHistory{})
	return database.DB.Where("datasource_id = ?", id).Delete(&model.Datasource{}).Error
}

func (s *DatasourceService) TestConnection(req *TestConnectionReq) (*TestConnectionResult, error) {
	port := 0
	if req.Port != nil {
		port = *req.Port
	}

	params := &dbtype.ConnectionParams{
		Type:     req.DBType,
		Host:     req.Host,
		Port:     port,
		Username: req.Username,
		Password: req.Password,
		Database: req.DatabaseName,
		FilePath: req.FilePath,
	}

	result := dbtype.TestConnect(params)

	msg := result.Message
	if len(msg) > 200 {
		msg = msg[:200]
	}

	return &TestConnectionResult{
		Success: result.Success,
		Message: msg,
		Version: result.Version,
		Cost:    result.LatencyMs,
	}, nil
}

func (s *DatasourceService) TestConnectionInternal(ds *model.Datasource) {
	plainPwd := ds.Password
	if ds.Password != "" {
		decrypted, err := crypto.DecryptAES(ds.Password, config.App.AESKey)
		if err == nil {
			plainPwd = decrypted
		}
	}

	params := &dbtype.ConnectionParams{
		Type:     ds.DBType,
		Host:     ds.Host,
		Port:     ds.Port,
		Username: ds.Username,
		Password: plainPwd,
		Database: ds.DefaultDB,
		FilePath: ds.FilePath,
	}

	result := dbtype.TestConnect(params)

	status := "fail"
	if result.Success {
		status = "ok"
	}

	database.DB.Model(&model.Datasource{}).Where("datasource_id = ?", ds.DatasourceID).Updates(map[string]interface{}{
		"conn_status":       status,
		"version":           result.Version,
		"conn_latency_ms":  result.LatencyMs,
		"last_conn_test_at": time.Now(),
		"updated_at":        time.Now(),
	})
}

func (s *DatasourceService) ListRecentlyDatasource(limit int) ([]*DatasourceListItem, error) {
	if limit <= 0 {
		limit = 8
	}
	if limit > 100 {
		limit = 100
	}

	var list []model.Datasource
	err := database.DB.Order("updated_at DESC").Limit(limit).Find(&list).Error
	if err != nil {
		return nil, err
	}

	result := make([]*DatasourceListItem, 0, len(list))
	for _, ds := range list {
		tags := []string{}
		if ds.Tags != "" {
			tags = strings.Split(ds.Tags, ",")
		}

		result = append(result, &DatasourceListItem{
			ID:             ds.DatasourceID,
			Name:           ds.Name,
			DatasourceType: ds.DBType,
			Type:           ds.DBType,
			DBType:         ds.DBType,
			Env:            ds.Env,
			Host:           ds.Host,
			Port:           ds.Port,
			Username:       ds.Username,
			Password:       maskPassword(ds.Password),
			DatabaseName:   ds.DefaultDB,
			Description:    ds.Remark,
			Status:         ds.Status,
			ConnectStatus:  ds.ConnStatus,
			CreateTime:     ds.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdateTime:     ds.UpdatedAt.Format("2006-01-02 15:04:05"),
			Tags:           tags,
			ColorLabel:     ds.ColorLabel,
			ConnLatencyMs:  ds.ConnLatencyMs,
			FilePath:       ds.FilePath,
			OpenMode:       ds.OpenMode,
		})
	}

	return result, nil
}

// ==================== V1 兼容方法 ====================

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
		ds.Env = "dev"
	}
	if ds.ColorLabel == "" {
		ds.ColorLabel = "blue"
	}
	if strings.ToLower(ds.DBType) == "sqlite" && ds.OpenMode == "" {
		ds.OpenMode = "rw"
	}
	return database.DB.Create(ds).Error
}

func (s *DatasourceService) List(page, pageSize int, keyword, dbType, status, sortBy, businessId, env string) ([]model.Datasource, int64, error) {
	var list []model.Datasource
	var total int64
	q := database.DB.Model(&model.Datasource{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR host LIKE ? OR remark LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if dbType != "" {
		q = q.Where("db_type = ?", dbType)
	}
	if status != "" {
		if status == "connected" || status == "ok" {
			q = q.Where("conn_status = ?", "ok")
		} else if status == "failed" || status == "fail" {
			q = q.Where("conn_status = ?", "fail")
		} else if status == "untested" {
			q = q.Where("conn_status IS NULL OR conn_status = '' OR conn_status = 'unknown'")
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
		ColorLabel: ds.ColorLabel,
		Tags:       ds.Tags,
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
	database.DB.Model(&model.Datasource{}).Where("conn_status = ?", "ok").Count(&okCount)
	database.DB.Model(&model.Datasource{}).Where("conn_status = ?", "fail").Count(&failCount)

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
		"conn_status":       "fail",
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
