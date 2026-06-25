package adapter

import (
	"database/sql"
	"fmt"
	"strings"

	"dbm-lite/internal/model"
)

type PermissionAdapter interface {
	CreateUser(conn *sql.DB, username, host, password string) error
	AlterUser(conn *sql.DB, username, host, password string) error
	DropUser(conn *sql.DB, username, host string) error
	EnableUser(conn *sql.DB, username, host string, enable bool) error
	Grant(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error
	Revoke(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error
	ListUsers(conn *sql.DB) ([]map[string]interface{}, error)
	ListUserPrivileges(conn *sql.DB) ([]map[string]interface{}, error)
	ListRoles(conn *sql.DB) ([]map[string]interface{}, error)
	FlushPrivileges(conn *sql.DB) error
	SupportColumnLevel() bool
	SupportDynamicPrivileges() bool
	GetDynamicPrivileges() []string
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

func (a *BasePermissionAdapter) GetDynamicPrivileges() []string {
	return nil
}

func (a *BasePermissionAdapter) ListUserPrivileges(conn *sql.DB) ([]map[string]interface{}, error) {
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

func (a *MySQLPermissionAdapter) ListUsers(conn *sql.DB) ([]map[string]interface{}, error) {
	rows, err := conn.Query("SELECT User, Host FROM mysql.user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []map[string]interface{}
	for rows.Next() {
		var user, host string
		if err := rows.Scan(&user, &host); err != nil {
			return nil, err
		}
		users = append(users, map[string]interface{}{"username": user, "host": host})
	}
	return users, nil
}

func (a *MySQLPermissionAdapter) ListRoles(conn *sql.DB) ([]map[string]interface{}, error) {
	rows, err := conn.Query("SELECT Role_name FROM mysql.role_edges")
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

func (a *MySQLPermissionAdapter) FlushPrivileges(conn *sql.DB) error {
	_, err := conn.Exec("FLUSH PRIVILEGES")
	return err
}

func (a *MySQLPermissionAdapter) SupportColumnLevel() bool {
	return true
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
			"username":      user,
			"host":          host,
			"database_name": dbName,
			"table_name":    tableName,
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
			"username":      user,
			"host":          host,
			"database_name": dbName,
			"table_name":    tableName,
			"privilege_type": privType,
		})
	}

	return privileges, nil
}

type TiDBPermissionAdapter struct {
	MySQLPermissionAdapter
}

func (a *TiDBPermissionAdapter) CreateUser(conn *sql.DB, username, host, password string) error {
	stmt := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'", username, host, password)
	_, err := conn.Exec(stmt)
	return err
}

func (a *TiDBPermissionAdapter) AlterUser(conn *sql.DB, username, host, password string) error {
	stmt := fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED BY '%s'", username, host, password)
	_, err := conn.Exec(stmt)
	return err
}

func (a *TiDBPermissionAdapter) Grant(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error {
	err := a.MySQLPermissionAdapter.Grant(conn, username, host, privType, dbName, tableName, columns)
	if err != nil {
		return err
	}
	return a.FlushPrivileges(conn)
}

func (a *TiDBPermissionAdapter) Revoke(conn *sql.DB, username, host, privType, dbName, tableName string, columns []string) error {
	err := a.MySQLPermissionAdapter.Revoke(conn, username, host, privType, dbName, tableName, columns)
	if err != nil {
		return err
	}
	return a.FlushPrivileges(conn)
}

func (a *TiDBPermissionAdapter) EnableUser(conn *sql.DB, username, host string, enable bool) error {
	err := a.MySQLPermissionAdapter.EnableUser(conn, username, host, enable)
	if err != nil {
		return err
	}
	return a.FlushPrivileges(conn)
}

func (a *TiDBPermissionAdapter) FlushPrivileges(conn *sql.DB) error {
	_, err := conn.Exec("FLUSH PRIVILEGES")
	return err
}

func (a *TiDBPermissionAdapter) SupportDynamicPrivileges() bool {
	return true
}

func (a *TiDBPermissionAdapter) GetDynamicPrivileges() []string {
	return []string{
		"BACKUP_ADMIN",
		"ROLE_ADMIN",
		"CONNECTION_ADMIN",
		"PLACEMENT_ADMIN",
		"RESTRICTED_USER_ADMIN",
		"RESTRICTED_ROLE_ADMIN",
		"RESTRICTED_TABLE_ADMIN",
		"RESTRICTED_GRANT_ADMIN",
	}
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

func (a *SQLitePermissionAdapter) SupportColumnLevel() bool {
	return false
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