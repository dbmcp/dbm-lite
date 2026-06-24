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
)

type DashboardService struct{}

func NewDashboardService() *DashboardService { return &DashboardService{} }

type CountPair struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type DailySeriesItem struct {
	Day   string `json:"day"`
	Count int64  `json:"count"`
}

type TopItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type DashboardStats struct {
	TotalUsers        int64             `json:"totalUsers"`
	TotalDatasources  int64             `json:"totalDatasources"`
	TotalSqlExec      int64             `json:"totalSqlExec"`
	TodaySqlExec      int64             `json:"todaySqlExec"`
	TotalAuditLogs    int64             `json:"totalAuditLogs"`
	TodayAuditLogs    int64             `json:"todayAuditLogs"`
	OnlineDatasources int64             `json:"onlineDatasources"`
	FailedDatasources int64             `json:"failedDatasources"`
	TotalServers      int64             `json:"totalServers"`
	TotalBusinesses   int64             `json:"totalBusinesses"`
	TotalProjects     int64             `json:"totalProjects"`

	DatasourceByType []CountPair        `json:"datasourceByType"`
	RecentSqlHistory []model.SQLHistory `json:"recentSqlHistory"`
	RecentAudit      []model.AuditLog   `json:"recentAudit"`
	SqlDailyTrend    []DailySeriesItem  `json:"sqlDailyTrend"`
	AuditDailyTrend  []DailySeriesItem  `json:"auditDailyTrend"`
	TopDatasources   []TopItem          `json:"topDatasources"`
	TopUsers         []TopItem          `json:"topUsers"`
	AvgLatencyMs     int64              `json:"avgLatencyMs"`
	FailRate         float64            `json:"failRate"`
	SystemNow        string             `json:"systemNow"`
}

func (s *DashboardService) Summary() (*DashboardStats, error) {
	stats := &DashboardStats{}
	now := time.Now()
	stats.SystemNow = now.Format("2006-01-02 15:04:05")
	db := database.DB

	db.Model(&model.User{}).Count(&stats.TotalUsers)
	db.Model(&model.Datasource{}).Count(&stats.TotalDatasources)
	db.Model(&model.SQLHistory{}).Count(&stats.TotalSqlExec)
	db.Model(&model.AuditLog{}).Count(&stats.TotalAuditLogs)
	db.Model(&model.Server{}).Count(&stats.TotalServers)
	db.Model(&model.Business{}).Count(&stats.TotalBusinesses)
	db.Model(&model.Project{}).Count(&stats.TotalProjects)

	startOfDay := now.Truncate(24 * time.Hour)
	db.Model(&model.SQLHistory{}).Where("created_at >= ?", startOfDay).Count(&stats.TodaySqlExec)
	db.Model(&model.AuditLog{}).Where("created_at >= ?", startOfDay).Count(&stats.TodayAuditLogs)

	db.Model(&model.Datasource{}).Where("conn_status IN ?", []string{"ok", "success"}).Count(&stats.OnlineDatasources)
	db.Model(&model.Datasource{}).Where("conn_status = ?", "fail").Count(&stats.FailedDatasources)

	var byType []struct {
		DBType string `gorm:"column:db_type"`
		Count  int64
	}
	db.Model(&model.Datasource{}).Select("db_type, COUNT(*) as count").Group("db_type").Order("count DESC").Scan(&byType)
	stats.DatasourceByType = make([]CountPair, 0, len(byType))
	for _, t := range byType {
		name := t.DBType
		if name == "" {
			name = "未分类"
		}
		stats.DatasourceByType = append(stats.DatasourceByType, CountPair{Name: name, Count: t.Count})
	}

	var history []model.SQLHistory
	db.Model(&model.SQLHistory{}).Order("created_at DESC").Limit(8).Find(&history)
	stats.RecentSqlHistory = history

	var audit []model.AuditLog
	db.Model(&model.AuditLog{}).Order("created_at DESC").Limit(8).Find(&audit)
	stats.RecentAudit = audit

	stats.SqlDailyTrend = buildDailyTrend("sql_history", "created_at", 14)
	stats.AuditDailyTrend = buildDailyTrend("audit_logs", "created_at", 14)

	var topDs []struct {
		DatasourceName string `gorm:"column:datasource_name"`
		Count          int64
	}
	db.Model(&model.SQLHistory{}).
		Select("datasource_name, COUNT(*) as count").
		Where("datasource_name <> ''").
		Group("datasource_name").
		Order("count DESC").
		Limit(6).
		Scan(&topDs)
	for _, t := range topDs {
		stats.TopDatasources = append(stats.TopDatasources, TopItem{Name: t.DatasourceName, Count: t.Count})
	}

	var topUsers []struct {
		Username string
		Count    int64
	}
	db.Model(&model.SQLHistory{}).
		Select("username, COUNT(*) as count").
		Where("username <> ''").
		Group("username").
		Order("count DESC").
		Limit(6).
		Scan(&topUsers)
	for _, t := range topUsers {
		stats.TopUsers = append(stats.TopUsers, TopItem{Name: t.Username, Count: t.Count})
	}

	var avg struct {
		Avg float64
	}
	db.Model(&model.SQLHistory{}).Select("COALESCE(AVG(duration_ms), 0) as avg").Scan(&avg)
	stats.AvgLatencyMs = int64(avg.Avg)

	if stats.TotalSqlExec > 0 {
		var failed int64
		db.Model(&model.SQLHistory{}).Where("status = ?", "failed").Count(&failed)
		stats.FailRate = float64(failed) / float64(stats.TotalSqlExec)
	}

	return stats, nil
}

// buildDailyTrend 按日期维度对近 N 天做统计，无数据日期补 0，统一返回 "YYYY-MM-DD"
func buildDailyTrend(tableName, timeCol string, days int) []DailySeriesItem {
	now := time.Now()
	start := now.Truncate(24 * time.Hour).AddDate(0, 0, -days+1)
	startStr := start.Format("2006-01-02 15:04:05")

	type Row struct {
		Day   string `gorm:"column:day"`
		Count int64  `gorm:"column:cnt"`
	}
	var rows []Row
	database.DB.Raw(
		"SELECT substr("+timeCol+",1,10) as day, COUNT(*) as cnt FROM "+tableName+
			" WHERE "+timeCol+" >= ? GROUP BY substr("+timeCol+",1,10) ORDER BY day ASC",
		startStr,
	).Scan(&rows)

	rowMap := make(map[string]int64, len(rows))
	for _, r := range rows {
		rowMap[r.Day] = r.Count
	}
	out := make([]DailySeriesItem, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		out = append(out, DailySeriesItem{Day: day, Count: rowMap[day]})
	}
	return out
}