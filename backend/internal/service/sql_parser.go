/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"regexp"
	"strings"
)

// ====== SQL 解析工具（轻量实现，兼容 MySQL/TiDB/SQLite） ======

type ParsedSQL struct {
	OperationType string          // select/insert/update/delete/create/drop/alter/other
	Tables        []TableRef      // 涉及的表
	Columns       []string        // 涉及的列（仅 SELECT）
}

type TableRef struct {
	DB    string
	Table string
}

// ParseSQL 轻量级 SQL 解析，提取表名/列名，仅用于权限校验
// 设计原则：解析失败绝不拦截执行，直接返回空的 Table 列表
func ParseSQL(sql string) *ParsedSQL {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return &ParsedSQL{OperationType: "unknown"}
	}
	// 去除行注释与块注释
	sql = stripComments(sql)
	upper := strings.ToUpper(sql)

	// 识别操作类型
	var opType string
	switch {
	case strings.HasPrefix(upper, "SELECT"):
		opType = "select"
	case strings.HasPrefix(upper, "INSERT"):
		opType = "insert"
	case strings.HasPrefix(upper, "UPDATE"):
		opType = "update"
	case strings.HasPrefix(upper, "DELETE"):
		opType = "delete"
	case strings.HasPrefix(upper, "CREATE"):
		opType = "create"
	case strings.HasPrefix(upper, "DROP"):
		opType = "drop"
	case strings.HasPrefix(upper, "ALTER"):
		opType = "alter"
	case strings.HasPrefix(upper, "SHOW"), strings.HasPrefix(upper, "DESC"), strings.HasPrefix(upper, "DESCRIBE"):
		opType = "describe"
	case strings.HasPrefix(upper, "SET"), strings.HasPrefix(upper, "USE"):
		opType = "session"
	default:
		opType = "other"
	}

	result := &ParsedSQL{OperationType: opType}

	// 表名提取（FROM/JOIN/UPDATE/INSERT INTO）
	// 通用正则：识别 `db`.`table` 或 db.table 或 `table` 或 table
	switch opType {
	case "select":
		result.Tables = extractSelectTables(sql)
		result.Columns = extractSelectColumns(sql)
	case "insert":
		result.Tables = extractInsertTables(sql)
	case "update":
		result.Tables = extractUpdateTables(sql)
	case "delete":
		result.Tables = extractDeleteTables(sql)
	}

	// 去重
	seen := make(map[string]struct{})
	unique := make([]TableRef, 0, len(result.Tables))
	for _, t := range result.Tables {
		key := strings.ToLower(t.DB + "." + t.Table)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, t)
	}
	result.Tables = unique
	return result
}

func stripComments(sql string) string {
	// 去除 /* ... */ 块注释
	re := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	out := re.ReplaceAllString(sql, "")
	// 去除 -- 行注释
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		if idx := strings.Index(l, "--"); idx >= 0 {
			lines[i] = l[:idx]
		}
		if idx := strings.Index(l, "#"); idx >= 0 {
			lines[i] = l[:idx]
		}
	}
	return strings.Join(lines, "\n")
}

func extractSelectTables(sql string) []TableRef {
	// 匹配 FROM/JOIN 后的表名（支持 db.table, `db`.`table`, `table`）
	re := regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+((?:` + "`" + `[^` + "`" + `]+` + "`" + `|[\w]+)(?:\.(?:` + "`" + `[^` + "`" + `]+` + "`" + `|[\w]+))?)`)
	matches := re.FindAllStringSubmatch(sql, -1)
	result := make([]TableRef, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		ref := parseTableRef(m[1])
		if ref.Table != "" {
			result = append(result, ref)
		}
	}
	return result
}

func extractInsertTables(sql string) []TableRef {
	re := regexp.MustCompile(`(?i)\bINTO\s+((?:` + "`" + `[^` + "`" + `]+` + "`" + `|[\w]+)(?:\.(?:` + "`" + `[^` + "`" + `]+` + "`" + `|[\w]+))?)`)
	matches := re.FindAllStringSubmatch(sql, -1)
	result := make([]TableRef, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		ref := parseTableRef(m[1])
		if ref.Table != "" {
			result = append(result, ref)
		}
	}
	return result
}

func extractUpdateTables(sql string) []TableRef {
	re := regexp.MustCompile(`(?i)\bUPDATE\s+((?:` + "`" + `[^` + "`" + `]+` + "`" + `|[\w]+)(?:\.(?:` + "`" + `[^` + "`" + `]+` + "`" + `|[\w]+))?)`)
	matches := re.FindAllStringSubmatch(sql, -1)
	result := make([]TableRef, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		ref := parseTableRef(m[1])
		if ref.Table != "" {
			result = append(result, ref)
		}
	}
	return result
}

func extractDeleteTables(sql string) []TableRef {
	// DELETE FROM db.table ...
	re := regexp.MustCompile(`(?i)\bFROM\s+((?:` + "`" + `[^` + "`" + `]+` + "`" + `|[\w]+)(?:\.(?:` + "`" + `[^` + "`" + `]+` + "`" + `|[\w]+))?)`)
	matches := re.FindAllStringSubmatch(sql, -1)
	result := make([]TableRef, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		ref := parseTableRef(m[1])
		if ref.Table != "" {
			result = append(result, ref)
		}
	}
	return result
}

func extractSelectColumns(sql string) []string {
	// 简化：尝试取 SELECT 和 FROM 之间的部分
	re := regexp.MustCompile(`(?i)SELECT\s+([\s\S]+?)\s+FROM\s+`)
	match := re.FindStringSubmatch(sql)
	if len(match) < 2 {
		return nil
	}
	colsSection := match[1]
	// 如果有 *，说明选择所有列，这里不做列级过滤的进一步展开
	if strings.Contains(colsSection, "*") {
		return nil
	}
	parts := strings.Split(colsSection, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		// 去除别名（AS xxx）
		if idx := strings.LastIndex(strings.ToUpper(p), " AS "); idx > 0 {
			p = strings.TrimSpace(p[:idx])
		}
		// 去除引号
		p = strings.Trim(p, "`\"[]")
		// 去除函数调用与表达式，仅保留简单列名
		if strings.Contains(p, "(") || strings.Contains(p, " ") || strings.Contains(p, ".") {
			// 取最后一部分
			idx := strings.LastIndex(p, ".")
			if idx >= 0 && idx < len(p)-1 {
				p = strings.TrimSpace(p[idx+1:])
				p = strings.Trim(p, "`\"[]")
			} else {
				continue
			}
		}
		if p != "" && p != "*" {
			out = append(out, p)
		}
	}
	return out
}

func parseTableRef(s string) TableRef {
	// 去除引号反引号方括号
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	// 去除别名
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return TableRef{}
	}
	ref := fields[0]
	parts := strings.Split(ref, ".")
	if len(parts) >= 2 {
		return TableRef{DB: parts[len(parts)-2], Table: parts[len(parts)-1]}
	}
	return TableRef{DB: "", Table: parts[0]}
}
