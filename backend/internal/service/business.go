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
	var bizCount int64
	database.DB.Model(&model.Business{}).Where("project_id = ?", id).Count(&bizCount)
	if bizCount > 0 {
		return errors.New("请先删除关联业务后再删除项目")
	}
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

// ==================== 项目成员管理 ====================

func (s *ProjectService) ListProjectMembers(projectId string) ([]map[string]interface{}, error) {
	var rows []map[string]interface{}
	err := database.DB.Table("project_members").
		Select("project_members.project_id as project_id, project_members.user_id as user_id, project_members.role as role, project_members.joined_at as joined_at, users.username as username, users.display_name as display_name").
		Joins("LEFT JOIN users ON users.user_id = project_members.user_id").
		Where("project_members.project_id = ?", projectId).
		Order("project_members.joined_at DESC").
		Scan(&rows).Error
	return rows, err
}

func (s *ProjectService) AssignProjectMembers(projectId string, userIds []string, role string) error {
	if role == "" {
		role = model.ProjectRoleViewer
	}
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("project_id = ?", projectId).Delete(&model.ProjectMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, uid := range userIds {
		if uid == "" {
			continue
		}
		pm := &model.ProjectMember{ProjectID: projectId, UserID: uid, Role: role, JoinedAt: time.Now()}
		if err := tx.Create(pm).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (s *ProjectService) AddProjectMember(projectId, userId, role string) error {
	if role == "" {
		role = model.ProjectRoleViewer
	}
	pm := &model.ProjectMember{ProjectID: projectId, UserID: userId, Role: role, JoinedAt: time.Now()}
	return database.DB.Create(pm).Error
}

func (s *ProjectService) RemoveProjectMember(projectId, userId string) error {
	return database.DB.Where("project_id = ? AND user_id = ?", projectId, userId).Delete(&model.ProjectMember{}).Error
}

// ==================== 业务成员管理 ====================

func (s *BusinessService) ListBusinessMembers(businessId string) ([]map[string]interface{}, error) {
	var rows []map[string]interface{}
	err := database.DB.Table("business_members").
		Select("business_members.business_id as business_id, business_members.user_id as user_id, business_members.role as role, business_members.joined_at as joined_at, users.username as username, users.display_name as display_name").
		Joins("LEFT JOIN users ON users.user_id = business_members.user_id").
		Where("business_members.business_id = ?", businessId).
		Order("business_members.joined_at DESC").
		Scan(&rows).Error
	return rows, err
}

func (s *BusinessService) AssignBusinessMembers(businessId string, userIds []string, role string) error {
	if role == "" {
		role = model.ProjectRoleViewer
	}
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("business_id = ?", businessId).Delete(&model.BusinessMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, uid := range userIds {
		if uid == "" {
			continue
		}
		bm := &model.BusinessMember{BusinessID: businessId, UserID: uid, Role: role, JoinedAt: time.Now()}
		if err := tx.Create(bm).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (s *BusinessService) AddBusinessMember(businessId, userId, role string) error {
	if role == "" {
		role = model.ProjectRoleViewer
	}
	bm := &model.BusinessMember{BusinessID: businessId, UserID: userId, Role: role, JoinedAt: time.Now()}
	return database.DB.Create(bm).Error
}

func (s *BusinessService) RemoveBusinessMember(businessId, userId string) error {
	return database.DB.Where("business_id = ? AND user_id = ?", businessId, userId).Delete(&model.BusinessMember{}).Error
}

// ==================== 项目/业务 概览统计 ====================

func (s *ProjectService) Overview(projectId string) (map[string]interface{}, error) {
	var bizCount int64
	database.DB.Model(&model.Business{}).Where("project_id = ?", projectId).Count(&bizCount)

	var byEnv []map[string]interface{}
	database.DB.Model(&model.Business{}).
		Select("env as env, COUNT(*) as total").
		Where("project_id = ?", projectId).
		Group("env").Scan(&byEnv)

	var memberCount int64
	database.DB.Model(&model.ProjectMember{}).Where("project_id = ?", projectId).Count(&memberCount)

	var serverCount int64
	database.DB.Model(&model.Server{}).Where("project_id = ?", projectId).Count(&serverCount)

	var dsCount int64
	database.DB.Model(&model.Datasource{}).Where("project_id = ?", projectId).Count(&dsCount)

	return map[string]interface{}{
		"projectId":    projectId,
		"businesses":   bizCount,
		"byEnv":        byEnv,
		"members":      memberCount,
		"servers":      serverCount,
		"datasources":  dsCount,
	}, nil
}

func (s *BusinessService) Overview(businessId string) (map[string]interface{}, error) {
	var serverCount int64
	database.DB.Model(&model.Server{}).Where("business_id = ?", businessId).Count(&serverCount)

	var dsCount int64
	database.DB.Model(&model.Datasource{}).Where("business_id = ?", businessId).Count(&dsCount)

	var memberCount int64
	database.DB.Model(&model.BusinessMember{}).Where("business_id = ?", businessId).Count(&memberCount)

	var biz *model.Business
	database.DB.Model(&model.Business{}).Where("business_id = ?", businessId).First(&biz)

	return map[string]interface{}{
		"businessId":  businessId,
		"projectId":   biz.ProjectID,
		"env":         biz.Env,
		"name":        biz.Name,
		"servers":     serverCount,
		"datasources": dsCount,
		"members":     memberCount,
	}, nil
}
