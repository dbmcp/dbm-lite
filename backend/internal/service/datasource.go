/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dbm-lite/internal/database"
	"dbm-lite/internal/dbtype"
	"dbm-lite/internal/model"

	"gorm.io/gorm"
)

// getUsernameMap 批量通过 user_id 查询 username，避免 N+1
func getUsernameMap(userIds []string) map[string]string {
	result := map[string]string{}
	if len(userIds) == 0 {
		return result
	}
	var users []model.User
	if err := database.DB.Select("user_id, username, display_name").Where("user_id IN ?", userIds).Find(&users).Error; err != nil {
		return result
	}
	for _, u := range users {
		if u.DisplayName != "" {
			result[u.UserID] = u.DisplayName
		} else {
			result[u.UserID] = u.Username
		}
	}
	return result
}

// getUsername 单个查询用户名
func getUsername(userId string) string {
	if userId == "" {
		return ""
	}
	var u model.User
	err := database.DB.Select("username, display_name").Where("user_id = ?", userId).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ""
		}
		return ""
	}
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Username
}

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

// GetById 根据 ID 查询数据源（包含解密后的密码）
func (s *DatasourceService) GetById(id string) (*model.Datasource, error) {
	var ds model.Datasource
	if err := database.DB.Where("datasource_id = ?", id).First(&ds).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDatasourceNotFound
		}
		return nil, err
	}
	if ds.Password != "" {
		if plain, err := dbtype.DecryptPassword(ds.Password); err == nil {
			ds.Password = plain
		}
	}
	return &ds, nil
}

// GetByIdNoDecrypt 根据 ID 查询数据源（不解密密码），用于列表/详情页返回
func (s *DatasourceService) GetByIdNoDecrypt(id string) (*model.Datasource, error) {
	var ds model.Datasource
	if err := database.DB.Where("datasource_id = ?", id).First(&ds).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDatasourceNotFound
		}
		return nil, err
	}
	ds.Password = ""
	return &ds, nil
}

// List 分页查询数据源，支持关键字 / 类型 / 状态过滤，可排序
func (s *DatasourceService) List(page, pageSize int, keyword, dbType, status, sortBy string, extra ...string) ([]*model.Datasource, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	query := database.DB.Model(&model.Datasource{})
	if keyword != "" {
		k := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR host LIKE ? OR remark LIKE ?", k, k, k)
	}
	if dbType != "" {
		query = query.Where("db_type = ?", strings.ToLower(dbType))
	}
	if status != "" {
		query = query.Where("conn_status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.Datasource
	orderExpr := "updated_at DESC"
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "name":
		orderExpr = "name ASC"
	case "recent", "lasttest":
		orderExpr = "last_conn_test_at DESC"
	}
	if err := query.Order(orderExpr).
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	for _, d := range list {
		d.Password = ""
	}
	return list, total, nil
}

// Create 创建数据源。密码会被 AES 加密后再写入。
func (s *DatasourceService) Create(ds *model.Datasource, rawPassword, userId, username string) error {
	if ds.DatasourceID == "" {
		ds.DatasourceID = "ds" + time.Now().Format("20060102150405")
	}
	if rawPassword != "" {
		encrypted, err := dbtype.EncryptPassword(rawPassword)
		if err != nil {
			return err
		}
		ds.Password = encrypted
	}
	now := time.Now()
	if ds.CreatedAt.IsZero() {
		ds.CreatedAt = model.DateTime(now)
	}
	ds.UpdatedAt = model.DateTime(now)
	ds.OwnerID = userId
	return database.DB.Create(ds).Error
}

// Update 更新数据源，支持按字段 map 增量更新；提供 rawPassword 时会同步加密更新密码
// 使用字段白名单避免前端传来的非数据库字段（如 autoCreateServer、createdByName 等）导致 SQL 错误
func (s *DatasourceService) Update(id string, updates map[string]interface{}, rawPassword string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("数据源 id 不能为空")
	}
	// 字段名映射：前端 JSON 字段 -> GORM 数据库列名
	fieldMap := map[string]string{
		"name":            "name",
		"dbType":          "db_type",
		"host":            "host",
		"port":            "port",
		"username":        "username",
		"defaultDatabase": "default_db",
		"filePath":        "file_path",
		"openMode":        "open_mode",
		"charset":         "charset",
		"timezone":        "timezone",
		"sslMode":         "ssl_mode",
		"sslCaFile":       "ssl_ca_file",
		"colorLabel":      "color_label",
		"tags":            "tags",
		"businessId":      "business_id",
		"serverId":        "server_id",
		"projectId":       "project_id",
		"env":             "env",
		"remark":          "remark",
		"timeout":         "timeout",
		"connStatus":      "conn_status",
		"connLatencyMs":   "conn_latency_ms",
		"status":          "status",
		"readOnly":        "read_only",
		"version":         "version",
		"createdBy":       "created_by",
		"ownerId":         "owner_id",
		"orgId":           "org_id",
		"datasourceType":  "datasource_type",
		"type":            "type",
	}

	cleaned := map[string]interface{}{}
	for k, v := range updates {
		// 处理 sslMode 特殊类型（前端传布尔值，数据库存字符串）
		if k == "sslMode" {
			switch val := v.(type) {
			case bool:
				if val {
					cleaned["ssl_mode"] = "true"
				} else {
					cleaned["ssl_mode"] = "false"
				}
			case string:
				cleaned["ssl_mode"] = val
			default:
				cleaned["ssl_mode"] = "false"
			}
			continue
		}
		// readOnly 前端传 bool，db 存 tinyint
		if k == "readOnly" {
			if b, ok := v.(bool); ok {
				cleaned["read_only"] = b
			} else if s2, ok := v.(string); ok {
				cleaned["read_only"] = s2 == "true" || s2 == "1"
			}
			continue
		}
		// port/timeout/connLatencyMs 数值类型容错
		if k == "port" || k == "timeout" || k == "connLatencyMs" {
			switch val := v.(type) {
			case float64:
				cleaned[fieldMap[k]] = int64(val)
			case int:
				cleaned[fieldMap[k]] = int64(val)
			case int64:
				cleaned[fieldMap[k]] = val
			case string:
				if n, err := strconv.ParseInt(val, 10, 64); err == nil {
					cleaned[fieldMap[k]] = n
				}
			}
			continue
		}
		// 只更新白名单中的字段
		if colName, ok := fieldMap[k]; ok {
			// 字符串字段去 nil/空指针问题
			if v == nil {
				cleaned[colName] = ""
			} else {
				cleaned[colName] = v
			}
		}
	}
	cleaned["updated_at"] = time.Now()
	if rawPassword != "" {
		encrypted, err := dbtype.EncryptPassword(rawPassword)
		if err != nil {
			return err
		}
		cleaned["password"] = encrypted
	}
	return database.DB.Model(&model.Datasource{}).
		Where("datasource_id = ?", id).
		Updates(cleaned).Error
}

// Delete 删除数据源
func (s *DatasourceService) Delete(id string) error {
	return database.DB.Where("datasource_id = ?", id).Delete(&model.Datasource{}).Error
}

// Copy 复制数据源（包括密码），新名称默认追加 "-copy"
func (s *DatasourceService) Copy(id, userId, username string) (*model.Datasource, error) {
	src, err := s.GetById(id)
	if err != nil {
		return nil, err
	}
	encrypted := ""
	if src.Password != "" {
		if enc, encErr := dbtype.EncryptPassword(src.Password); encErr != nil {
			return nil, fmt.Errorf("password encrypt failed: %w", encErr)
		} else {
			encrypted = enc
		}
	}
	now := time.Now()
	newDs := &model.Datasource{
		DatasourceID:  "ds" + time.Now().Format("20060102150405"),
		Name:          src.Name + "-copy",
		DBType:        src.DBType,
		Host:          src.Host,
		Port:          src.Port,
		Username:      src.Username,
		Password:      encrypted,
		DefaultDB:     src.DefaultDB,
		FilePath:      src.FilePath,
		OpenMode:      src.OpenMode,
		Charset:       src.Charset,
		Timezone:      src.Timezone,
		SSLMode:       src.SSLMode,
		SSLCAFile:     src.SSLCAFile,
		ReadOnly:      src.ReadOnly,
		ColorLabel:    src.ColorLabel,
		Tags:          src.Tags,
		BusinessID:    src.BusinessID,
		ServerID:      src.ServerID,
		ProjectID:     src.ProjectID,
		Env:           src.Env,
		Remark:        src.Remark,
		Status:        src.Status,
		Timeout:       src.Timeout,
		OwnerID:       userId,
		CreatedAt:     model.DateTime(now),
		UpdatedAt:     model.DateTime(now),
		ConnStatus:    "",
		ConnLatencyMs: 0,
	}
	if err := database.DB.Create(newDs).Error; err != nil {
		return nil, err
	}
	newDs.Password = ""
	return newDs, nil
}

// UpdateConnStatus 更新数据源连接状态、延迟、版本
func (s *DatasourceService) UpdateConnStatus(id, connStatus string, latencyMs int64, version string) error {
	updates := map[string]interface{}{
		"conn_status":       connStatus,
		"conn_latency_ms":   latencyMs,
		"last_conn_test_at": time.Now(),
		"updated_at":        time.Now(),
	}
	if version != "" {
		updates["version"] = version
	}
	return database.DB.Model(&model.Datasource{}).
		Where("datasource_id = ?", id).
		Updates(updates).Error
}

// Stats 返回数据源统计（总数 / 成功 / 失败 / 未测试）
func (s *DatasourceService) Stats() (map[string]interface{}, error) {
	var total int64
	if err := database.DB.Model(&model.Datasource{}).Count(&total).Error; err != nil {
		return nil, err
	}
	var ok, fail int64
	_ = database.DB.Model(&model.Datasource{}).Where("conn_status = ?", model.ConnStatusOK).Count(&ok).Error
	_ = database.DB.Model(&model.Datasource{}).Where("conn_status = ?", model.ConnStatusFail).Count(&fail).Error
	return map[string]interface{}{
		"total":    total,
		"success":  ok,
		"fail":     fail,
		"untested": total - ok - fail,
	}, nil
}
