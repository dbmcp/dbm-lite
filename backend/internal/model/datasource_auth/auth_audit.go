package datasource_auth

import "time"

type DatasourceAuthAudit struct {
	ID          string    `gorm:"column:id;primaryKey;size:64" json:"id"`
	Operator    string    `gorm:"column:operator;size:64" json:"operator"`
	OperatorID  string    `gorm:"column:operator_id;size:64" json:"operatorId"`
	OperType    string    `gorm:"column:oper_type;size:64" json:"operType"`
	OperObject  string    `gorm:"column:oper_object;size:256" json:"operObject"`
	DatasourceID string   `gorm:"column:datasource_id;size:64" json:"datasourceId"`
	ClientIP    string    `gorm:"column:client_ip;size:64" json:"clientIP"`
	Result      string    `gorm:"column:result;size:32" json:"result"`
	Detail      string    `gorm:"column:detail;type:text" json:"detail"`
	OperTime    time.Time `gorm:"column:oper_time" json:"operTime"`
}

func (DatasourceAuthAudit) TableName() string {
	return "datasource_auth_audits"
}

const (
	AuditResultSuccess = "success"
	AuditResultFailed  = "failed"
)