/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type DateTime time.Time

func (dt DateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(dt)
	if t.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + t.Format("2006-01-02 15:04:05") + `"`), nil
}

func (dt *DateTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" {
		*dt = DateTime{}
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		t2, err2 := time.Parse(time.RFC3339, s)
		if err2 != nil {
			return err
		}
		*dt = DateTime(t2)
		return nil
	}
	*dt = DateTime(t)
	return nil
}

func (dt DateTime) String() string {
	t := time.Time(dt)
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func (dt DateTime) IsZero() bool {
	return time.Time(dt).IsZero()
}

func (dt DateTime) Format(layout string) string {
	t := time.Time(dt)
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

func (dt *DateTime) Scan(value interface{}) error {
	if value == nil {
		*dt = DateTime{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*dt = DateTime(v)
	case string:
		return dt.UnmarshalJSON([]byte(`"` + v + `"`))
	case []byte:
		return dt.UnmarshalJSON([]byte(`"` + string(v) + `"`))
	default:
		return fmt.Errorf("unsupported scan type for DateTime: %T", value)
	}
	return nil
}

func (dt DateTime) Value() (driver.Value, error) {
	t := time.Time(dt)
	if t.IsZero() {
		return nil, nil
	}
	return t, nil
}

type Datasource struct {
	DatasourceID   string     `gorm:"column:datasource_id;primaryKey;size:64" json:"datasourceId"`
	Name           string     `gorm:"column:name;size:128;not null;index" json:"name"`
	DBType         string     `gorm:"column:db_type;size:32;not null;index" json:"dbType"`
	Type           string     `gorm:"column:type;size:32" json:"type"`
	DatasourceType string     `gorm:"column:datasource_type;size:32;default:'rdbms'" json:"datasourceType"`
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
	CreatedByName  string     `gorm:"-" json:"createdByName"`
	CreatedAt      DateTime   `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      DateTime   `gorm:"column:updated_at" json:"updatedAt"`
	LastConnTestAt *DateTime  `gorm:"column:last_conn_test_at" json:"lastConnTestAt,omitempty"`
	ConnStatus     string     `gorm:"column:conn_status;size:32" json:"connStatus"`
	ConnLatencyMs  int64      `gorm:"column:conn_latency_ms;default:0" json:"connLatencyMs"`
	Status         string     `gorm:"column:status;size:32;default:'active'" json:"status"`
	Timeout        int        `gorm:"column:timeout;default:0" json:"timeout,omitempty"`
	OwnerID        string     `gorm:"column:owner_id;size:64" json:"ownerId"`
	OrgID          string     `gorm:"column:org_id;size:64" json:"orgId"`
	LastUseTime    *DateTime  `gorm:"column:last_use_time" json:"lastUseTime,omitempty"`
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
	ConnStatusOK        = "ok"
	ConnStatusFail      = "fail"
	ConnStatusNone      = ""
	ConnStatusSuccess   = "success"
	ConnStatusUnknown   = "unknown"
	ConnStatusConnecting = "connecting"
)

func (d *Datasource) IsConnectionOK() bool {
	return d.ConnStatus == ConnStatusOK
}

func (d *Datasource) HasPassword() bool {
	return d.Password != ""
}
