package adapter

import (
	"database/sql"
	"fmt"
	"strings"

	"dbm-lite/internal/model"
)

type ObjectType string

const (
	ObjectTypeDatabase  ObjectType = "database"
	ObjectTypeTable     ObjectType = "table"
	ObjectTypeColumn    ObjectType = "column"
	ObjectTypeView      ObjectType = "view"
	ObjectTypeProcedure ObjectType = "procedure"
	ObjectTypeFunction  ObjectType = "function"
	ObjectTypeTrigger   ObjectType = "trigger"
	ObjectTypeEvent     ObjectType = "event"
)

type SystemPrivilege struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type PermissionAdapter interface {
	CreateUser(conn *sql.DB, username, host, password string) error
	AlterUser(conn *sql.DB, username, host, password string) error
	DropUser(conn *sql.DB, username, host string) error
	RenameUser(conn *sql.DB, oldUsername, oldHost, newUsername, newHost string) error
	EnableUser(conn *sql.DB, username, host string, enable bool) error
	SetPasswordExpire(conn *sql.DB, username, host string, expireTime string) error
	SetAccountLock(conn *sql.DB, username, host string, lockDays int) error

	Grant(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error
	Revoke(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error
	GrantObject(conn *sql.DB, username, host string, privs []string, objectType ObjectType, dbName, objName string, columns []string) error
	RevokeObject(conn *sql.DB, username, host string, privs []string, objectType ObjectType, dbName, objName string, columns []string) error

	GrantSystemPrivileges(conn *sql.DB, username, host string, privs []string) error
	RevokeSystemPrivileges(conn *sql.DB, username, host string, privs []string) error
	GetGrantedSystemPrivileges(conn *sql.DB, username, host string) ([]string, error)

	ListUsers(conn *sql.DB) ([]map[string]interface{}, error)
	ListUserPrivileges(conn *sql.DB) ([]map[string]interface{}, error)
	ListRoles(conn *sql.DB) ([]map[string]interface{}, error)
	ListDatabases(conn *sql.DB) ([]string, error)
	ListTables(conn *sql.DB, dbName string) ([]string, error)
	ListColumns(conn *sql.DB, dbName, tableName string) ([]string, error)
	ListViews(conn *sql.DB, dbName string) ([]string, error)
	ListProcedures(conn *sql.DB, dbName string) ([]string, error)
	ListFunctions(conn *sql.DB, dbName string) ([]string, error)
	ListTriggers(conn *sql.DB, dbName string) ([]string, error)
	ListEvents(conn *sql.DB, dbName string) ([]string, error)

	GetUserEffectiveGrants(conn *sql.DB, username, host string) ([]map[string]interface{}, error)

	FlushPrivileges(conn *sql.DB) error

	SupportColumnLevel() bool
	SupportDynamicPrivileges() bool
	SupportRoles() bool
	GetDynamicPrivileges() []SystemPrivilege
	GetSystemPrivileges() []SystemPrivilege
	GetObjectPrivileges(objectType ObjectType) []string
}

func NewPermissionAdapter(dbType string) PermissionAdapter {
	switch dbType {
	case model.DBTypeMySQL:
		return &MySQLPermissionAdapter{}
	case model.DBTypeTiDB:
		return &TiDBPermissionAdapter{}
	case model.DBTypeSQLite:
		return &SQLitePermissionAdapter{}
	case "postgresql":
		return &PostgreSQLPermissionAdapter{}
	default:
		return &MySQLPermissionAdapter{}
	}
}

type BasePermissionAdapter struct{}

func (a *BasePermissionAdapter) getPrivileges(privType string) string {
	switch privType {
	case "readonly":
		return "SELECT"
	case "dml":
		return "SELECT,INSERT,UPDATE,DELETE"
	case "ddl":
		return "SELECT,INSERT,UPDATE,DELETE,CREATE,ALTER,DROP,TRUNCATE"
	default:
		return "SELECT"
	}
}

func (a *BasePermissionAdapter) FlushPrivileges(conn *sql.DB) error {
	return nil
}

func (a *BasePermissionAdapter) SupportColumnLevel() bool {
	return false
}

func (a *BasePermissionAdapter) SupportDynamicPrivileges() bool {
	return false
}

func (a *BasePermissionAdapter) SupportRoles() bool {
	return false
}

func (a *BasePermissionAdapter) GetDynamicPrivileges() []SystemPrivilege {
	return nil
}

func (a *BasePermissionAdapter) GetSystemPrivileges() []SystemPrivilege {
	return nil
}

func (a *BasePermissionAdapter) GetObjectPrivileges(objectType ObjectType) []string {
	switch objectType {
	case ObjectTypeDatabase, ObjectTypeTable, ObjectTypeView:
		return []string{"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP", "INDEX", "REFERENCES", "TRIGGER"}
	case ObjectTypeProcedure, ObjectTypeFunction:
		return []string{"EXECUTE", "ALTER ROUTINE", "CREATE ROUTINE"}
	case ObjectTypeTrigger:
		return []string{"TRIGGER"}
	case ObjectTypeEvent:
		return []string{"EVENT"}
	default:
		return []string{"SELECT", "INSERT", "UPDATE", "DELETE"}
	}
}

func (a *BasePermissionAdapter) ListUserPrivileges(conn *sql.DB) ([]map[string]interface{}, error) {
	return nil, nil
}

func (a *BasePermissionAdapter) RenameUser(conn *sql.DB, oldUsername, oldHost, newUsername, newHost string) error {
	return nil
}

func (a *BasePermissionAdapter) SetPasswordExpire(conn *sql.DB, username, host string, expireTime string) error {
	return nil
}

func (a *BasePermissionAdapter) SetAccountLock(conn *sql.DB, username, host string, lockDays int) error {
	return nil
}

func (a *BasePermissionAdapter) GrantObject(conn *sql.DB, username, host string, privs []string, objectType ObjectType, dbName, objName string, columns []string) error {
	return nil
}

func (a *BasePermissionAdapter) RevokeObject(conn *sql.DB, username, host string, privs []string, objectType ObjectType, dbName, objName string, columns []string) error {
	return nil
}

func (a *BasePermissionAdapter) GrantSystemPrivileges(conn *sql.DB, username, host string, privs []string) error {
	return nil
}

func (a *BasePermissionAdapter) RevokeSystemPrivileges(conn *sql.DB, username, host string, privs []string) error {
	return nil
}

func (a *BasePermissionAdapter) GetGrantedSystemPrivileges(conn *sql.DB, username, host string) ([]string, error) {
	return nil, nil
}

func (a *BasePermissionAdapter) ListDatabases(conn *sql.DB) ([]string, error) {
	return nil, nil
}

func (a *BasePermissionAdapter) ListTables(conn *sql.DB, dbName string) ([]string, error) {
	return nil, nil
}

func (a *BasePermissionAdapter) ListColumns(conn *sql.DB, dbName, tableName string) ([]string, error) {
	return nil, nil
}

func (a *BasePermissionAdapter) ListViews(conn *sql.DB, dbName string) ([]string, error) {
	return nil, nil
}

func (a *BasePermissionAdapter) ListProcedures(conn *sql.DB, dbName string) ([]string, error) {
	return nil, nil
}

func (a *BasePermissionAdapter) ListFunctions(conn *sql.DB, dbName string) ([]string, error) {
	return nil, nil
}

func (a *BasePermissionAdapter) ListTriggers(conn *sql.DB, dbName string) ([]string, error) {
	return nil, nil
}

func (a *BasePermissionAdapter) ListEvents(conn *sql.DB, dbName string) ([]string, error) {
	return nil, nil
}

func (a *BasePermissionAdapter) GetUserEffectiveGrants(conn *sql.DB, username, host string) ([]map[string]interface{}, error) {
	return nil, nil
}

type MySQLPermissionAdapter struct {
	BasePermissionAdapter
}

func (a *MySQLPermissionAdapter) CreateUser(conn *sql.DB, username, host, password string) error {
	stmt := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'", username, host, password)
	_, err := conn.Exec(stmt)
	return err
}

func (a *MySQLPermissionAdapter) AlterUser(conn *sql.DB, username, host, password string) error {
	stmt := fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED BY '%s'", username, host, password)
	_, err := conn.Exec(stmt)
	return err
}

func (a *MySQLPermissionAdapter) DropUser(conn *sql.DB, username, host string) error {
	stmt := fmt.Sprintf("DROP USER '%s'@'%s'", username, host)
	_, err := conn.Exec(stmt)
	return err
}

func (a *MySQLPermissionAdapter) RenameUser(conn *sql.DB, oldUsername, oldHost, newUsername, newHost string) error {
	stmt := fmt.Sprintf("RENAME USER '%s'@'%s' TO '%s'@'%s'", oldUsername, oldHost, newUsername, newHost)
	_, err := conn.Exec(stmt)
	return err
}

func (a *MySQLPermissionAdapter) EnableUser(conn *sql.DB, username, host string, enable bool) error {
	var stmt string
	if enable {
		stmt = fmt.Sprintf("ALTER USER '%s'@'%s' ACCOUNT UNLOCK", username, host)
	} else {
		stmt = fmt.Sprintf("ALTER USER '%s'@'%s' ACCOUNT LOCK", username, host)
	}
	_, err := conn.Exec(stmt)
	if err != nil {
		return err
	}
	return a.FlushPrivileges(conn)
}

func (a *MySQLPermissionAdapter) SetPasswordExpire(conn *sql.DB, username, host string, expireTime string) error {
	var stmt string
	if expireTime == "" {
		stmt = fmt.Sprintf("ALTER USER '%s'@'%s' PASSWORD EXPIRE NEVER", username, host)
	} else {
		stmt = fmt.Sprintf("ALTER USER '%s'@'%s' PASSWORD EXPIRE '%s'", username, host, expireTime)
	}
	_, err := conn.Exec(stmt)
	return err
}

func (a *MySQLPermissionAdapter) SetAccountLock(conn *sql.DB, username, host string, lockDays int) error {
	var stmt string
	if lockDays <= 0 {
		stmt = fmt.Sprintf("ALTER USER '%s'@'%s' ACCOUNT UNLOCK", username, host)
	} else {
		stmt = fmt.Sprintf("ALTER USER '%s'@'%s' ACCOUNT LOCK", username, host)
	}
	_, err := conn.Exec(stmt)
	return err
}

func (a *MySQLPermissionAdapter) Grant(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error {
	privs := a.getPrivileges(privType)
	if tableName == "" {
		stmt := fmt.Sprintf("GRANT %s ON %s.* TO '%s'@'%s'", privs, dbName, username, host)
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
		return a.FlushPrivileges(conn)
	}
	if len(columns) > 0 {
		cols := strings.Join(columns, ",")
		stmt := fmt.Sprintf("GRANT %s (%s) ON %s.%s TO '%s'@'%s'", privs, cols, dbName, tableName, username, host)
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
		return a.FlushPrivileges(conn)
	}
	stmt := fmt.Sprintf("GRANT %s ON %s.%s TO '%s'@'%s'", privs, dbName, tableName, username, host)
	_, err := conn.Exec(stmt)
	if err != nil {
		return err
	}
	return a.FlushPrivileges(conn)
}

func (a *MySQLPermissionAdapter) Revoke(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error {
	privs := a.getPrivileges(privType)
	if tableName == "" {
		stmt := fmt.Sprintf("REVOKE %s ON %s.* FROM '%s'@'%s'", privs, dbName, username, host)
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
		return a.FlushPrivileges(conn)
	}
	if len(columns) > 0 {
		cols := strings.Join(columns, ",")
		stmt := fmt.Sprintf("REVOKE %s (%s) ON %s.%s FROM '%s'@'%s'", privs, cols, dbName, tableName, username, host)
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
		return a.FlushPrivileges(conn)
	}
	stmt := fmt.Sprintf("REVOKE %s ON %s.%s FROM '%s'@'%s'", privs, dbName, tableName, username, host)
	_, err := conn.Exec(stmt)
	if err != nil {
		return err
	}
	return a.FlushPrivileges(conn)
}

func (a *MySQLPermissionAdapter) GrantObject(conn *sql.DB, username, host string, privs []string, objectType ObjectType, dbName, objName string, columns []string) error {
	privStr := strings.Join(privs, ",")
	var stmt string

	switch objectType {
	case ObjectTypeDatabase:
		stmt = fmt.Sprintf("GRANT %s ON %s.* TO '%s'@'%s'", privStr, dbName, username, host)
	case ObjectTypeTable, ObjectTypeView:
		if len(columns) > 0 {
			cols := strings.Join(columns, ",")
			stmt = fmt.Sprintf("GRANT %s (%s) ON %s.%s TO '%s'@'%s'", privStr, cols, dbName, objName, username, host)
		} else {
			stmt = fmt.Sprintf("GRANT %s ON %s.%s TO '%s'@'%s'", privStr, dbName, objName, username, host)
		}
	case ObjectTypeColumn:
		cols := strings.Join(columns, ",")
		stmt = fmt.Sprintf("GRANT %s (%s) ON %s.%s TO '%s'@'%s'", privStr, cols, dbName, objName, username, host)
	case ObjectTypeProcedure, ObjectTypeFunction:
		stmt = fmt.Sprintf("GRANT %s ON %s.%s TO '%s'@'%s'", privStr, dbName, objName, username, host)
	case ObjectTypeTrigger:
		stmt = fmt.Sprintf("GRANT TRIGGER ON %s.%s TO '%s'@'%s'", dbName, objName, username, host)
	case ObjectTypeEvent:
		stmt = fmt.Sprintf("GRANT EVENT ON %s.* TO '%s'@'%s'", dbName, username, host)
	default:
		stmt = fmt.Sprintf("GRANT %s ON %s.%s TO '%s'@'%s'", privStr, dbName, objName, username, host)
	}

	_, err := conn.Exec(stmt)
	if err != nil {
		return err
	}
	return a.FlushPrivileges(conn)
}

func (a *MySQLPermissionAdapter) RevokeObject(conn *sql.DB, username, host string, privs []string, objectType ObjectType, dbName, objName string, columns []string) error {
	privStr := strings.Join(privs, ",")
	var stmt string

	switch objectType {
	case ObjectTypeDatabase:
		stmt = fmt.Sprintf("REVOKE %s ON %s.* FROM '%s'@'%s'", privStr, dbName, username, host)
	case ObjectTypeTable, ObjectTypeView:
		if len(columns) > 0 {
			cols := strings.Join(columns, ",")
			stmt = fmt.Sprintf("REVOKE %s (%s) ON %s.%s FROM '%s'@'%s'", privStr, cols, dbName, objName, username, host)
		} else {
			stmt = fmt.Sprintf("REVOKE %s ON %s.%s FROM '%s'@'%s'", privStr, dbName, objName, username, host)
		}
	case ObjectTypeColumn:
		cols := strings.Join(columns, ",")
		stmt = fmt.Sprintf("REVOKE %s (%s) ON %s.%s FROM '%s'@'%s'", privStr, cols, dbName, objName, username, host)
	case ObjectTypeProcedure, ObjectTypeFunction:
		stmt = fmt.Sprintf("REVOKE %s ON %s.%s FROM '%s'@'%s'", privStr, dbName, objName, username, host)
	case ObjectTypeTrigger:
		stmt = fmt.Sprintf("REVOKE TRIGGER ON %s.%s FROM '%s'@'%s'", dbName, objName, username, host)
	case ObjectTypeEvent:
		stmt = fmt.Sprintf("REVOKE EVENT ON %s.* FROM '%s'@'%s'", dbName, username, host)
	default:
		stmt = fmt.Sprintf("REVOKE %s ON %s.%s FROM '%s'@'%s'", privStr, dbName, objName, username, host)
	}

	_, err := conn.Exec(stmt)
	if err != nil {
		return err
	}
	return a.FlushPrivileges(conn)
}

func (a *MySQLPermissionAdapter) GrantSystemPrivileges(conn *sql.DB, username, host string, privs []string) error {
	privStr := strings.Join(privs, ",")
	stmt := fmt.Sprintf("GRANT %s ON *.* TO '%s'@'%s'", privStr, username, host)
	_, err := conn.Exec(stmt)
	if err != nil {
		return err
	}
	return a.FlushPrivileges(conn)
}

func (a *MySQLPermissionAdapter) RevokeSystemPrivileges(conn *sql.DB, username, host string, privs []string) error {
	privStr := strings.Join(privs, ",")
	stmt := fmt.Sprintf("REVOKE %s ON *.* FROM '%s'@'%s'", privStr, username, host)
	_, err := conn.Exec(stmt)
	if err != nil {
		return err
	}
	return a.FlushPrivileges(conn)
}

func (a *MySQLPermissionAdapter) GetGrantedSystemPrivileges(conn *sql.DB, username, host string) ([]string, error) {
	stmt := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", username, host)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []string
	for rows.Next() {
		var grant string
		if err := rows.Scan(&grant); err != nil {
			return nil, err
		}
		grants = append(grants, grant)
	}
	return grants, nil
}

func (a *MySQLPermissionAdapter) ListUsers(conn *sql.DB) ([]map[string]interface{}, error) {
	rows, err := conn.Query("SELECT User, Host, Account_locked, Password_expired FROM mysql.user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []map[string]interface{}
	for rows.Next() {
		var user, host, locked, expired string
		if err := rows.Scan(&user, &host, &locked, &expired); err != nil {
			return nil, err
		}
		users = append(users, map[string]interface{}{
			"username":       user,
			"host":           host,
			"account_locked": locked == "Y",
			"password_expired": expired == "Y",
		})
	}
	return users, nil
}

func (a *MySQLPermissionAdapter) ListRoles(conn *sql.DB) ([]map[string]interface{}, error) {
	rows, err := conn.Query("SELECT Role_name, Role_host FROM mysql.role_edges")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []map[string]interface{}
	for rows.Next() {
		var name, host string
		if err := rows.Scan(&name, &host); err != nil {
			return nil, err
		}
		roles = append(roles, map[string]interface{}{"name": name, "host": host})
	}
	return roles, nil
}

func (a *MySQLPermissionAdapter) FlushPrivileges(conn *sql.DB) error {
	_, err := conn.Exec("FLUSH PRIVILEGES")
	return err
}

func (a *MySQLPermissionAdapter) SupportColumnLevel() bool {
	return true
}

func (a *MySQLPermissionAdapter) SupportRoles() bool {
	return true
}

func (a *MySQLPermissionAdapter) GetSystemPrivileges() []SystemPrivilege {
	return []SystemPrivilege{
		{"PROCESS", "查看进程", "管理"},
		{"RELOAD", "执行FLUSH操作", "管理"},
		{"CREATE USER", "创建用户", "管理"},
		{"DROP USER", "删除用户", "管理"},
		{"FILE", "文件导入导出", "管理"},
		{"SUPER", "高级管理权限", "管理"},
		{"SHUTDOWN", "关闭服务器", "管理"},
		{"REPLICATION SLAVE", "复制从库", "复制"},
		{"REPLICATION CLIENT", "复制客户端", "复制"},
		{"SHOW DATABASES", "查看数据库列表", "查询"},
		{"SELECT", "查询权限", "查询"},
		{"LOCK TABLES", "锁定表", "数据操作"},
		{"EXECUTE", "执行存储过程", "数据操作"},
		{"ALTER ROUTINE", "修改存储过程", "数据操作"},
		{"CREATE ROUTINE", "创建存储过程", "数据操作"},
		{"INDEX", "索引权限", "数据操作"},
		{"TRIGGER", "触发器权限", "数据操作"},
		{"EVENT", "事件调度权限", "数据操作"},
		{"CREATE TABLESPACE", "创建表空间", "DDL"},
		{"ALTER TABLESPACE", "修改表空间", "DDL"},
		{"DROP TABLESPACE", "删除表空间", "DDL"},
		{"CREATE TEMPORARY TABLES", "创建临时表", "DDL"},
		{"CREATE VIEW", "创建视图", "DDL"},
		{"SHOW VIEW", "查看视图", "DDL"},
		{"CREATE ROLE", "创建角色", "角色管理"},
		{"DROP ROLE", "删除角色", "角色管理"},
		{"GRANT OPTION", "授权选项", "角色管理"},
	}
}

func (a *MySQLPermissionAdapter) ListUserPrivileges(conn *sql.DB) ([]map[string]interface{}, error) {
	var privileges []map[string]interface{}

	rows, err := conn.Query(`
		SELECT u.User, u.Host, d.Db, '' as Table_name, 
			CASE 
				WHEN d.Select_priv = 'Y' AND d.Insert_priv = 'Y' AND d.Update_priv = 'Y' AND d.Delete_priv = 'Y' AND d.Create_priv = 'Y' AND d.Alter_priv = 'Y' AND d.Drop_priv = 'Y' THEN 'ddl'
				WHEN d.Select_priv = 'Y' AND d.Insert_priv = 'Y' AND d.Update_priv = 'Y' AND d.Delete_priv = 'Y' THEN 'dml'
				WHEN d.Select_priv = 'Y' THEN 'readonly'
				ELSE 'readonly'
			END as privilege_type
		FROM mysql.db d
		JOIN mysql.user u ON d.User = u.User AND d.Host = u.Host
		WHERE d.Db != '' AND d.Db NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user, host, dbName, tableName, privType string
		if err := rows.Scan(&user, &host, &dbName, &tableName, &privType); err != nil {
			return nil, err
		}
		privileges = append(privileges, map[string]interface{}{
			"username":       user,
			"host":           host,
			"database_name":  dbName,
			"table_name":     tableName,
			"privilege_type": privType,
		})
	}

	rows, err = conn.Query(`
		SELECT tp.User, tp.Host, tp.Db, tp.Table_name,
			CASE 
				WHEN tp.Select_priv = 'Y' AND tp.Insert_priv = 'Y' AND tp.Update_priv = 'Y' AND tp.Delete_priv = 'Y' AND tp.Create_priv = 'Y' AND tp.Alter_priv = 'Y' AND tp.Drop_priv = 'Y' THEN 'ddl'
				WHEN tp.Select_priv = 'Y' AND tp.Insert_priv = 'Y' AND tp.Update_priv = 'Y' AND tp.Delete_priv = 'Y' THEN 'dml'
				WHEN tp.Select_priv = 'Y' THEN 'readonly'
				ELSE 'readonly'
			END as privilege_type
		FROM mysql.tables_priv tp
		WHERE tp.Db != '' AND tp.Db NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
	`)
	if err != nil {
		return privileges, nil
	}
	defer rows.Close()

	for rows.Next() {
		var user, host, dbName, tableName, privType string
		if err := rows.Scan(&user, &host, &dbName, &tableName, &privType); err != nil {
			return privileges, nil
		}
		privileges = append(privileges, map[string]interface{}{
			"username":       user,
			"host":           host,
			"database_name":  dbName,
			"table_name":     tableName,
			"privilege_type": privType,
		})
	}

	return privileges, nil
}

func (a *MySQLPermissionAdapter) ListDatabases(conn *sql.DB) ([]string, error) {
	rows, err := conn.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbs []string
	for rows.Next() {
		var db string
		if err := rows.Scan(&db); err != nil {
			return nil, err
		}
		if db != "information_schema" && db != "performance_schema" && db != "sys" && db != "mysql" {
			dbs = append(dbs, db)
		}
	}
	return dbs, nil
}

func (a *MySQLPermissionAdapter) ListTables(conn *sql.DB, dbName string) ([]string, error) {
	stmt := fmt.Sprintf("SHOW TABLES FROM `%s`", dbName)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (a *MySQLPermissionAdapter) ListColumns(conn *sql.DB, dbName, tableName string) ([]string, error) {
	stmt := fmt.Sprintf("SHOW COLUMNS FROM `%s`.`%s`", dbName, tableName)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var col, colType, nullable, key, defaultValue, extra string
		if err := rows.Scan(&col, &colType, &nullable, &key, &defaultValue, &extra); err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	return cols, nil
}

func (a *MySQLPermissionAdapter) ListViews(conn *sql.DB, dbName string) ([]string, error) {
	stmt := fmt.Sprintf("SELECT TABLE_NAME FROM information_schema.VIEWS WHERE TABLE_SCHEMA = '%s'", dbName)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var views []string
	for rows.Next() {
		var view string
		if err := rows.Scan(&view); err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (a *MySQLPermissionAdapter) ListProcedures(conn *sql.DB, dbName string) ([]string, error) {
	stmt := fmt.Sprintf("SELECT ROUTINE_NAME FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA = '%s' AND ROUTINE_TYPE = 'PROCEDURE'", dbName)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var procs []string
	for rows.Next() {
		var proc string
		if err := rows.Scan(&proc); err != nil {
			return nil, err
		}
		procs = append(procs, proc)
	}
	return procs, nil
}

func (a *MySQLPermissionAdapter) ListFunctions(conn *sql.DB, dbName string) ([]string, error) {
	stmt := fmt.Sprintf("SELECT ROUTINE_NAME FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA = '%s' AND ROUTINE_TYPE = 'FUNCTION'", dbName)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var funcs []string
	for rows.Next() {
		var fn string
		if err := rows.Scan(&fn); err != nil {
			return nil, err
		}
		funcs = append(funcs, fn)
	}
	return funcs, nil
}

func (a *MySQLPermissionAdapter) ListTriggers(conn *sql.DB, dbName string) ([]string, error) {
	stmt := fmt.Sprintf("SELECT TRIGGER_NAME FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA = '%s'", dbName)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var triggers []string
	for rows.Next() {
		var trig string
		if err := rows.Scan(&trig); err != nil {
			return nil, err
		}
		triggers = append(triggers, trig)
	}
	return triggers, nil
}

func (a *MySQLPermissionAdapter) ListEvents(conn *sql.DB, dbName string) ([]string, error) {
	stmt := fmt.Sprintf("SELECT EVENT_NAME FROM information_schema.EVENTS WHERE EVENT_SCHEMA = '%s'", dbName)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []string
	for rows.Next() {
		var event string
		if err := rows.Scan(&event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (a *MySQLPermissionAdapter) GetUserEffectiveGrants(conn *sql.DB, username, host string) ([]map[string]interface{}, error) {
	stmt := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", username, host)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []map[string]interface{}
	for rows.Next() {
		var grant string
		if err := rows.Scan(&grant); err != nil {
			return nil, err
		}
		grants = append(grants, map[string]interface{}{"grant": grant})
	}
	return grants, nil
}

type TiDBPermissionAdapter struct {
	MySQLPermissionAdapter
}

func (a *TiDBPermissionAdapter) SupportDynamicPrivileges() bool {
	return true
}

func (a *TiDBPermissionAdapter) GetDynamicPrivileges() []SystemPrivilege {
	return []SystemPrivilege{
		{"BACKUP_ADMIN", "备份管理", "备份恢复"},
		{"RESTORE_ADMIN", "恢复管理", "备份恢复"},
		{"ROLE_ADMIN", "角色管理", "管理"},
		{"CONNECTION_ADMIN", "连接管理", "管理"},
		{"PLACEMENT_ADMIN", "Placement管理", "管理"},
		{"DASHBOARD_CLIENT", "Dashboard访问", "监控"},
		{"RESTRICTED_USER_ADMIN", "受限用户管理", "管理"},
		{"RESTRICTED_ROLE_ADMIN", "受限角色管理", "管理"},
		{"RESTRICTED_TABLE_ADMIN", "受限表管理", "管理"},
		{"RESTRICTED_GRANT_ADMIN", "受限授权管理", "管理"},
		{"SHARDING_DDL_ADMIN", "分库分表DDL", "管理"},
		{"CONTROL_GROUP_ADMIN", "控制组管理", "管理"},
		{"ENCRYPTION_KEY_ADMIN", "加密密钥管理", "安全"},
		{"SESSION_VARIABLES_ADMIN", "会话变量管理", "管理"},
		{"SET_USER_ID", "设置用户ID", "管理"},
		{"CONFIG_ADMIN", "配置管理", "管理"},
		{"RESOURCE_GROUP_ADMIN", "资源组管理", "管理"},
		{"BINDING_ADMIN", "SQL绑定管理", "SQL优化"},
		{"DUMP_ADMIN", "数据导出管理", "数据操作"},
		{"CHECKER_ADMIN", "校验管理", "管理"},
		{"TELEMETRY_ADMIN", "遥测管理", "监控"},
		{"PERSIST_CONFIG_ADMIN", "持久化配置管理", "管理"},
		{"CLUSTER_ADMIN", "集群管理", "管理"},
		{"REPLICA_READ_ADMIN", "副本读取管理", "复制"},
		{"AUDIT_ADMIN", "审计管理", "安全"},
	}
}

func (a *TiDBPermissionAdapter) GetSystemPrivileges() []SystemPrivilege {
	mysqlPrivs := a.MySQLPermissionAdapter.GetSystemPrivileges()
	dynamicPrivs := a.GetDynamicPrivileges()
	return append(mysqlPrivs, dynamicPrivs...)
}

type SQLitePermissionAdapter struct {
	BasePermissionAdapter
}

func (a *SQLitePermissionAdapter) CreateUser(conn *sql.DB, username, host, password string) error {
	return nil
}

func (a *SQLitePermissionAdapter) AlterUser(conn *sql.DB, username, host, password string) error {
	return nil
}

func (a *SQLitePermissionAdapter) DropUser(conn *sql.DB, username, host string) error {
	return nil
}

func (a *SQLitePermissionAdapter) EnableUser(conn *sql.DB, username, host string, enable bool) error {
	return nil
}

func (a *SQLitePermissionAdapter) Grant(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error {
	return nil
}

func (a *SQLitePermissionAdapter) Revoke(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error {
	return nil
}

func (a *SQLitePermissionAdapter) ListUsers(conn *sql.DB) ([]map[string]interface{}, error) {
	return []map[string]interface{}{{"username": "admin", "host": ""}}, nil
}

func (a *SQLitePermissionAdapter) ListRoles(conn *sql.DB) ([]map[string]interface{}, error) {
	return nil, nil
}

func (a *SQLitePermissionAdapter) ListUserPrivileges(conn *sql.DB) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"username":       "admin",
			"host":           "",
			"database_name":  "main",
			"table_name":     "",
			"privilege_type": "ddl",
		},
	}, nil
}

func (a *SQLitePermissionAdapter) ListDatabases(conn *sql.DB) ([]string, error) {
	return []string{"main"}, nil
}

func (a *SQLitePermissionAdapter) ListTables(conn *sql.DB, dbName string) ([]string, error) {
	rows, err := conn.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (a *SQLitePermissionAdapter) ListColumns(conn *sql.DB, dbName, tableName string) ([]string, error) {
	stmt := fmt.Sprintf("PRAGMA table_info(`%s`)", tableName)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var cid, name, colType, notNull, dfltValue, pk string
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		cols = append(cols, name)
	}
	return cols, nil
}

func (a *SQLitePermissionAdapter) ListViews(conn *sql.DB, dbName string) ([]string, error) {
	rows, err := conn.Query("SELECT name FROM sqlite_master WHERE type='view'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var views []string
	for rows.Next() {
		var view string
		if err := rows.Scan(&view); err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (a *SQLitePermissionAdapter) SupportColumnLevel() bool {
	return false
}

func (a *SQLitePermissionAdapter) GetUserEffectiveGrants(conn *sql.DB, username, host string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"grant": fmt.Sprintf("-- SQLite 用户: %s@%s 的权限信息", username, host)},
		{"grant": "-- SQLite 通过文件系统权限控制访问"},
		{"grant": "-- 平台内部权限规则已在 dbm-lite 中管理"},
	}, nil
}

type PostgreSQLPermissionAdapter struct {
	BasePermissionAdapter
}

func (a *PostgreSQLPermissionAdapter) CreateUser(conn *sql.DB, username, host, password string) error {
	stmt := fmt.Sprintf("CREATE ROLE %s WITH LOGIN PASSWORD '%s'", username, password)
	if host != "%" && host != "" {
		stmt += fmt.Sprintf(" FROM '%s'", host)
	}
	_, err := conn.Exec(stmt)
	return err
}

func (a *PostgreSQLPermissionAdapter) AlterUser(conn *sql.DB, username, host, password string) error {
	stmt := fmt.Sprintf("ALTER ROLE %s WITH PASSWORD '%s'", username, password)
	_, err := conn.Exec(stmt)
	return err
}

func (a *PostgreSQLPermissionAdapter) DropUser(conn *sql.DB, username, host string) error {
	stmt := fmt.Sprintf("DROP ROLE %s", username)
	_, err := conn.Exec(stmt)
	return err
}

func (a *PostgreSQLPermissionAdapter) EnableUser(conn *sql.DB, username, host string, enable bool) error {
	var stmt string
	if enable {
		stmt = fmt.Sprintf("ALTER ROLE %s WITH LOGIN", username)
	} else {
		stmt = fmt.Sprintf("ALTER ROLE %s WITH NOLOGIN", username)
	}
	_, err := conn.Exec(stmt)
	return err
}

func (a *PostgreSQLPermissionAdapter) Grant(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error {
	privs := a.getPrivileges(privType)
	if tableName == "" {
		stmt := fmt.Sprintf("GRANT %s ON ALL TABLES IN SCHEMA %s TO %s", privs, dbName, username)
		_, err := conn.Exec(stmt)
		return err
	}
	if len(columns) > 0 {
		cols := strings.Join(columns, ",")
		stmt := fmt.Sprintf("GRANT %s (%s) ON %s.%s TO %s", privs, cols, dbName, tableName, username)
		_, err := conn.Exec(stmt)
		return err
	}
	stmt := fmt.Sprintf("GRANT %s ON %s.%s TO %s", privs, dbName, tableName, username)
	_, err := conn.Exec(stmt)
	return err
}

func (a *PostgreSQLPermissionAdapter) Revoke(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error {
	privs := a.getPrivileges(privType)
	if tableName == "" {
		stmt := fmt.Sprintf("REVOKE %s ON ALL TABLES IN SCHEMA %s FROM %s", privs, dbName, username)
		_, err := conn.Exec(stmt)
		return err
	}
	if len(columns) > 0 {
		cols := strings.Join(columns, ",")
		stmt := fmt.Sprintf("REVOKE %s (%s) ON %s.%s FROM %s", privs, cols, dbName, tableName, username)
		_, err := conn.Exec(stmt)
		return err
	}
	stmt := fmt.Sprintf("REVOKE %s ON %s.%s FROM %s", privs, dbName, tableName, username)
	_, err := conn.Exec(stmt)
	return err
}

func (a *PostgreSQLPermissionAdapter) ListUsers(conn *sql.DB) ([]map[string]interface{}, error) {
	rows, err := conn.Query("SELECT usename FROM pg_user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []map[string]interface{}
	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err != nil {
			return nil, err
		}
		users = append(users, map[string]interface{}{"username": user, "host": ""})
	}
	return users, nil
}

func (a *PostgreSQLPermissionAdapter) ListRoles(conn *sql.DB) ([]map[string]interface{}, error) {
	rows, err := conn.Query("SELECT rolname FROM pg_roles WHERE rolcanlogin = false")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []map[string]interface{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, map[string]interface{}{"name": name})
	}
	return roles, nil
}

func (a *PostgreSQLPermissionAdapter) SupportColumnLevel() bool {
	return true
}

func (a *PostgreSQLPermissionAdapter) GetUserEffectiveGrants(conn *sql.DB, username, host string) ([]map[string]interface{}, error) {
	stmt := fmt.Sprintf(`SELECT usename, pg_has_role(usename, 'SUPERUSER', 'MEMBER') as is_superuser FROM pg_user WHERE usename = '%s'`, username)
	rows, err := conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []map[string]interface{}
	for rows.Next() {
		var uname string
		var isSuper bool
		if err := rows.Scan(&uname, &isSuper); err != nil {
			return nil, err
		}
		if isSuper {
			grants = append(grants, map[string]interface{}{"grant": fmt.Sprintf("GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %s", uname)})
			grants = append(grants, map[string]interface{}{"grant": fmt.Sprintf("GRANT SUPERUSER TO %s", uname)})
		} else {
			grants = append(grants, map[string]interface{}{"grant": fmt.Sprintf("-- 用户 %s 的权限需通过 GRANT 命令查看", uname)})
			grants = append(grants, map[string]interface{}{"grant": "-- 使用 \\dp 或 pg_permissions 查看具体权限"})
		}
	}
	return grants, nil
}
