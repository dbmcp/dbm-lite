/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"errors"
	"time"

	"dbm-lite/internal/database"
	"dbm-lite/internal/model"

	"github.com/google/uuid"
)

// ==================== 项目管理 ====================

type ProjectService struct{}

func NewProjectService() *ProjectService { return &ProjectService{} }

func (s *ProjectService) CreateProject(p *model.Project) error {
	if p.Name == "" {
		return errors.New("项目名称不能为空")
	}
	p.ProjectID = uuid.New().String()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return database.DB.Create(p).Error
}

func (s *ProjectService) ListProjects(page, pageSize int, keyword string) ([]model.Project, int64, error) {
	var list []model.Project
	var total int64
	q := database.DB.Model(&model.Project{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *ProjectService) AllProjects() ([]model.Project, error) {
	var list []model.Project
	err := database.DB.Model(&model.Project{}).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (s *ProjectService) UpdateProject(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return database.DB.Model(&model.Project{}).Where("project_id = ?", id).Updates(updates).Error
}

func (s *ProjectService) DeleteProject(id string) error {
	// 检查是否有关联的业务
	var bizCount int64
	database.DB.Model(&model.Business{}).Where("project_id = ?", id).Count(&bizCount)
	if bizCount > 0 {
		return errors.New("请先删除关联业务后再删除项目")
	}
	// 检查是否有关联的服务器
	var srvCount int64
	database.DB.Model(&model.Server{}).Where("project_id = ?", id).Count(&srvCount)
	if srvCount > 0 {
		return errors.New("请先删除关联服务器后再删除项目")
	}
	return database.DB.Where("project_id = ?", id).Delete(&model.Project{}).Error
}

// ==================== 业务管理 ====================

type BusinessService struct{}

func NewBusinessService() *BusinessService { return &BusinessService{} }

func (s *BusinessService) CreateBusiness(b *model.Business) error {
	if b.Name == "" {
		return errors.New("业务名称不能为空")
	}
	b.BusinessID = uuid.New().String()
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()
	if b.Env == "" {
		b.Env = model.EnvDev
	}
	return database.DB.Create(b).Error
}

func (s *BusinessService) ListBusinesses(page, pageSize int, keyword, projectId string) ([]model.Business, int64, error) {
	var list []model.Business
	var total int64
	q := database.DB.Model(&model.Business{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if projectId != "" {
		q = q.Where("project_id = ?", projectId)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *BusinessService) AllBusinesses() ([]model.Business, error) {
	var list []model.Business
	err := database.DB.Model(&model.Business{}).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (s *BusinessService) UpdateBusiness(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return database.DB.Model(&model.Business{}).Where("business_id = ?", id).Updates(updates).Error
}

func (s *BusinessService) DeleteBusiness(id string) error {
	return database.DB.Where("business_id = ?", id).Delete(&model.Business{}).Error
}

// ==================== 服务器管理 ====================

type ServerService struct{}

func NewServerService() *ServerService { return &ServerService{} }

func (s *ServerService) Create(srv *model.Server, createdBy string) error {
	if srv.Name == "" {
		return errors.New("服务器名称不能为空")
	}
	srv.ServerID = uuid.New().String()
	srv.CreatedBy = createdBy
	srv.CreatedAt = time.Now()
	srv.UpdatedAt = time.Now()
	if srv.Status == "" {
		srv.Status = model.StatusActive
	}
	return database.DB.Create(srv).Error
}

func (s *ServerService) EnsureByHost(host string, createdBy string) (*model.Server, error) {
	if host == "" {
		return nil, nil
	}
	var existing model.Server
	err := database.DB.Where("host = ?", host).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	srv := &model.Server{
		Name:   "Server-" + host,
		Host:   host,
		Port:   22,
		OS:     "Linux",
		Status: model.StatusActive,
	}
	err = s.Create(srv, createdBy)
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func (s *ServerService) List(page, pageSize int, keyword string) ([]model.Server, int64, error) {
	var list []model.Server
	var total int64
	q := database.DB.Model(&model.Server{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR host LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *ServerService) ListByProject(page, pageSize int, keyword, projectId string) ([]model.Server, int64, error) {
	var list []model.Server
	var total int64
	q := database.DB.Model(&model.Server{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR host LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if projectId != "" {
		q = q.Where("project_id = ?", projectId)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *ServerService) All() ([]model.Server, error) {
	var list []model.Server
	err := database.DB.Model(&model.Server{}).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (s *ServerService) Update(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return database.DB.Model(&model.Server{}).Where("server_id = ?", id).Updates(updates).Error
}

func (s *ServerService) Delete(id string) error {
	return database.DB.Where("server_id = ?", id).Delete(&model.Server{}).Error
}
