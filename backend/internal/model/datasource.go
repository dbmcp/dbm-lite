/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import (
	"time"
)

type Datasource struct {
	DatasourceID   string     `gorm:"column:datasource_id;primaryKey;size:64" json:"datasourceId"`
	Name           string     `gorm:"column:name;size:128;not null;index" json:"name"`
	DBType         string     `gorm:"column:db_type;size:32;not null;index" json:"dbType"`
	Host           string     `gorm:"column:host;size:255;index" json:"host"`
	Port           int        `gorm:"column:port" json:"port"`
	Username       string     `gorm:"column:username;size:128" json:"username"`
	Password       string     `gorm:"column:password;size:512" json:"-"`
	DefaultDB      string     `gorm:"column:default_db;size:128" json:"defaultDatabase"`
	FilePath       string     `gorm:"column:file_path;size:512" json:"filePath,omitempty"`
	OpenMode       string     `gorm:"column:open_mode;size:16;default:'rw'" json:"openMode,omitempty"`
	Charset        string     `gorm:"column:charset;size:32;default:'utf8mb4'" json:"charset"`
	Timezone       string     `gorm:"column:timezone;size:64;default:'Local'" json:"timezone"`
	SSLMode        string     `gorm:"column:ssl_mode;size:32;default:'false'" json:"sslMode"`
	SSLCAFile      string     `gorm:"column:ssl_ca_file;size:512" json:"sslCaFile,omitempty"`
	ReadOnly       bool       `gorm:"column:read_only;default:false" json:"readOnly"`
	ColorLabel     string     `gorm:"column:color_label;size:16;default:'blue'" json:"colorLabel"`
	Version        string     `gorm:"column:version;size:64" json:"version"`
	Tags           string     `gorm:"column:tags;size:256" json:"tags"`
	BusinessID     string     `gorm:"column:business_id;size:64;index" json:"businessId"`
	ServerID       string     `gorm:"column:server_id;size:64;index" json:"serverId"`
	ProjectID      string     `gorm:"column:project_id;size:64;index" json:"projectId"`
	Env            string     `gorm:"column:env;size:32" json:"env"`
	Remark         string     `gorm:"column:remark;size:512" json:"remark"`
	CreatedBy      string     `gorm:"column:created_by;size:64" json:"createdBy"`
	CreatedAt      time.Time  `gorm:"column:created_at;index" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	LastConnTestAt *time.Time `gorm:"column:last_conn_test_at" json:"lastConnTestAt,omitempty"`
	ConnStatus     string     `gorm:"column:conn_status;size:32" json:"connStatus"`
	ConnLatencyMs  int64      `gorm:"column:conn_latency_ms;default:0" json:"connLatencyMs"`
	Status         string     `gorm:"column:status;size:32;default:'active'" json:"status"`
}

func (Datasource) TableName() string { return "datasources" }

const (
	DBTypeMySQL  = "mysql"
	DBTypeTiDB   = "tidb"
	DBTypeSQLite = "sqlite"
)

func IsSupportedDBType(t string) bool {
	switch t {
	case DBTypeMySQL, DBTypeTiDB, DBTypeSQLite:
		return true
	default:
		return false
	}
}

const (
	ColorLabelBlue   = "blue"
	ColorLabelGreen  = "green"
	ColorLabelRed    = "red"
	ColorLabelYellow = "yellow"
	ColorLabelPurple = "purple"
	ColorLabelOrange = "orange"
	ColorLabelGray   = "gray"
)

func ValidColorLabels() []string {
	return []string{ColorLabelBlue, ColorLabelGreen, ColorLabelRed, ColorLabelYellow, ColorLabelPurple, ColorLabelOrange, ColorLabelGray}
}

const (
	ConnStatusOK   = "ok"
	ConnStatusFail = "fail"
	ConnStatusNone = ""
)

func (d *Datasource) IsConnectionOK() bool {
	return d.ConnStatus == ConnStatusOK
}

func (d *Datasource) HasPassword() bool {
	return d.Password != ""
}
