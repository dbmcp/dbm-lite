/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package sqllint

import (
	"regexp"
	"strings"
)

type ReviewResult struct {
	IsHighRisk  bool     `json:"isHighRisk"`
	Reasons     []string `json:"reasons"`
	HasWhere    bool     `json:"hasWhere"`
	IsSelect    bool     `json:"isSelect"`
	IsDDL       bool     `json:"isDdl"`
	IsDML       bool     `json:"isDml"`
	Suggestions []string `json:"suggestions"`
}

var (
	deletePattern = regexp.MustCompile(`(?i)^\s*(DELETE|DROP|TRUNCATE|ALTER\s+TABLE|UPDATE)\s+`)
	// 返回结果集的语句统一视为“查询类”，包括 SHOW / DESC / DESCRIBE / EXPLAIN / ANALYZE (MySQL) / CHECK TABLE / WITH / PRAGMA (SQLite) 等
	selectPattern = regexp.MustCompile(`(?i)^\s*(SELECT|SHOW|DESCRIBE|DESC|EXPLAIN|ANALYZE|CHECK\s+TABLE|WITH|PRAGMA)\b`)
	dmlPattern    = regexp.MustCompile(`(?i)^\s*(INSERT|UPDATE|DELETE|REPLACE)\s+`)
	ddlPattern    = regexp.MustCompile(`(?i)^\s*(CREATE|ALTER|DROP|TRUNCATE|RENAME)\s+`)
	wherePattern  = regexp.MustCompile(`(?i)\bWHERE\b`)
	limitPattern  = regexp.MustCompile(`(?i)\bLIMIT\b`)
)

func Review(sql string) *ReviewResult {
	r := &ReviewResult{
		Reasons:     []string{},
		Suggestions: []string{},
	}

	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return r
	}

	r.IsSelect = selectPattern.MatchString(trimmed)
	r.IsDDL = ddlPattern.MatchString(trimmed)
	r.IsDML = dmlPattern.MatchString(trimmed)
	r.HasWhere = wherePattern.MatchString(trimmed)

	// 高危SQL检测
	if deletePattern.MatchString(trimmed) {
		if !r.HasWhere {
			r.IsHighRisk = true
			r.Reasons = append(r.Reasons, "检测到DELETE/UPDATE/DROP/TRUNCATE操作但没有WHERE条件，可能影响全表数据")
			r.Suggestions = append(r.Suggestions, "请添加WHERE条件限制影响范围，或在测试环境先验证")
		}
	}

	if strings.Contains(strings.ToUpper(trimmed), "DROP ") {
		r.IsHighRisk = true
		r.Reasons = append(r.Reasons, "检测到DROP操作，将永久删除表/数据库")
		r.Suggestions = append(r.Suggestions, "DROP操作不可恢复，请确认有备份且确实需要执行")
	}

	if strings.Contains(strings.ToUpper(trimmed), "TRUNCATE ") {
		r.IsHighRisk = true
		r.Reasons = append(r.Reasons, "检测到TRUNCATE操作，将清空全表且不可回滚")
		r.Suggestions = append(r.Suggestions, "TRUNCATE无法回滚，请确认有备份")
	}

	if r.IsSelect && !limitPattern.MatchString(trimmed) {
		r.Suggestions = append(r.Suggestions, "建议添加LIMIT限制返回行数，避免大查询影响性能")
	}

	return r
}
