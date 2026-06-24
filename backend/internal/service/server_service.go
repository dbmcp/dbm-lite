/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dbm-lite/internal/database"
	"dbm-lite/internal/dbtype"
	"dbm-lite/internal/model"

	"golang.org/x/crypto/ssh"
)

type ServerService struct{}

func NewServerService() *ServerService { return &ServerService{} }

func (s *ServerService) Create(srv *model.Server, createdBy string) error {
	if srv == nil {
		return errors.New("服务器数据为空")
	}
	if strings.TrimSpace(srv.Name) == "" {
		return errors.New("服务器名称不能为空")
	}
	if strings.TrimSpace(srv.Host) == "" {
		return errors.New("主机地址不能为空")
	}
	if srv.Port <= 0 || srv.Port > 65535 {
		srv.Port = 22
	}
	if srv.AuthType != model.ServerAuthPassword && srv.AuthType != model.ServerAuthKey {
		srv.AuthType = model.ServerAuthPassword
	}
	// 密码/私钥加密存储
	if srv.Password != "" {
		if enc, err := dbtype.EncryptPassword(srv.Password); err == nil {
			srv.Password = enc
		}
	}
	if srv.PrivateKey != "" {
		if enc, err := dbtype.EncryptPassword(srv.PrivateKey); err == nil {
			srv.PrivateKey = enc
		}
	}
	if srv.KeyPassphrase != "" {
		if enc, err := dbtype.EncryptPassword(srv.KeyPassphrase); err == nil {
			srv.KeyPassphrase = enc
		}
	}
	srv.ServerID = "srv" + time.Now().Format("20060102150405")
	srv.CreatedBy = createdBy
	srv.CreatedAt = time.Now()
	srv.UpdatedAt = time.Now()
	if srv.Status == "" {
		srv.Status = model.ServerStatusActive
	}
	if srv.ConnStatus == "" {
		srv.ConnStatus = model.ServerConnNone
	}
	if srv.Timeout <= 0 {
		srv.Timeout = 30
	}
	return database.DB.Create(srv).Error
}

// GetById 根据 ID 获取服务器（不解密敏感字段）
func (s *ServerService) GetById(id string) (*model.Server, error) {
	if id == "" {
		return nil, errors.New("ID 不能为空")
	}
	var srv model.Server
	if err := database.DB.Where("server_id = ?", id).First(&srv).Error; err != nil {
		return nil, err
	}
	return &srv, nil
}

// GetDecrypted 获取服务器并解密敏感字段，仅用于内部连接
func (s *ServerService) GetDecrypted(id string) (*model.Server, error) {
	srv, err := s.GetById(id)
	if err != nil {
		return nil, err
	}
	if srv.Password != "" {
		if plain, decErr := dbtype.DecryptPassword(srv.Password); decErr == nil {
			srv.Password = plain
		}
	}
	if srv.PrivateKey != "" {
		if plain, decErr := dbtype.DecryptPassword(srv.PrivateKey); decErr == nil {
			srv.PrivateKey = plain
		}
	}
	if srv.KeyPassphrase != "" {
		if plain, decErr := dbtype.DecryptPassword(srv.KeyPassphrase); decErr == nil {
			srv.KeyPassphrase = plain
		}
	}
	return srv, nil
}

// List 分页查询
func (s *ServerService) List(page, pageSize int, keyword, env, status string) ([]model.Server, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Server
	var total int64
	q := database.DB.Model(&model.Server{})
	if keyword != "" {
		k := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR host LIKE ? OR remark LIKE ?", k, k, k)
	}
	if env != "" {
		q = q.Where("env = ?", env)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ListByProject 按项目查询
func (s *ServerService) ListByProject(page, pageSize int, keyword, projectId string) ([]model.Server, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
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
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *ServerService) All() ([]model.Server, error) {
	var list []model.Server
	err := database.DB.Model(&model.Server{}).Order("created_at DESC").Find(&list).Error
	return list, err
}

// EnsureByHost 根据 host 查找服务器，不存在则创建一条简单记录
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
		Status: model.ServerStatusActive,
	}
	if err := s.Create(srv, createdBy); err != nil {
		return nil, err
	}
	return srv, nil
}

// Update 更新服务器（使用白名单字段，避免非数据库字段污染）
func (s *ServerService) Update(id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID 不能为空")
	}
	fieldMap := map[string]string{
		"name":          "name",
		"projectId":     "project_id",
		"businessId":    "business_id",
		"env":           "env",
		"host":          "host",
		"port":          "port",
		"username":      "username",
		"authType":      "auth_type",
		"password":      "password",
		"privateKey":    "private_key",
		"keyPassphrase": "key_passphrase",
		"os":            "os",
		"arch":          "arch",
		"version":       "version",
		"cpuCores":      "cpu_cores",
		"memoryGB":      "memory_gb",
		"diskGB":        "disk_gb",
		"status":        "status",
		"connStatus":    "conn_status",
		"connLatencyMs": "conn_latency_ms",
		"lastCheckTime": "last_check_time",
		"remark":        "remark",
		"tags":          "tags",
		"timeout":       "timeout",
	}
	cleaned := map[string]interface{}{}
	for k, v := range updates {
		// 敏感字段需加密
		if k == "password" || k == "privateKey" || k == "keyPassphrase" {
			if str, ok := v.(string); ok && str != "" {
				if enc, err := dbtype.EncryptPassword(str); err == nil {
					cleaned[fieldMap[k]] = enc
				}
			}
			continue
		}
		if colName, ok := fieldMap[k]; ok {
			cleaned[colName] = v
		}
	}
	cleaned["updated_at"] = time.Now()
	if len(cleaned) == 1 {
		return nil
	}
	return database.DB.Model(&model.Server{}).Where("server_id = ?", id).Updates(cleaned).Error
}

func (s *ServerService) Delete(id string) error {
	return database.DB.Where("server_id = ?", id).Delete(&model.Server{}).Error
}

// ToggleStatus 切换启用 / 禁用
func (s *ServerService) ToggleStatus(id string) (string, error) {
	srv, err := s.GetById(id)
	if err != nil {
		return "", err
	}
	newStatus := model.ServerStatusInactive
	if srv.Status == model.ServerStatusInactive || srv.Status == "" {
		newStatus = model.ServerStatusActive
	}
	return newStatus, database.DB.Model(srv).Update("status", newStatus).Error
}

// TestConnect 测试连接，并更新连接状态/延迟
func (s *ServerService) TestConnect(id string) (latencyMs int64, info map[string]interface{}, err error) {
	srv, err := s.GetDecrypted(id)
	if err != nil {
		_ = s.updateConnInfo(id, model.ServerConnFail, 0, nil)
		return 0, nil, err
	}
	start := time.Now()
	info, err = s.doSSHCheck(srv)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		_ = s.updateConnInfo(id, model.ServerConnFail, latency, nil)
		return latency, info, err
	}
	_ = s.updateConnInfo(id, model.ServerConnOK, latency, info)
	return latency, info, nil
}

// TestConnectByForm 直接通过传入的连接参数测试（不写库）
func (s *ServerService) TestConnectByForm(host string, port int, username, authType, password, privateKey, keyPassphrase string) (map[string]interface{}, error) {
	if host == "" || username == "" {
		return nil, errors.New("主机与用户名不能为空")
	}
	if port <= 0 || port > 65535 {
		port = 22
	}
	srv := &model.Server{
		Host:          host,
		Port:          port,
		Username:      username,
		AuthType:      authType,
		Password:      password,
		PrivateKey:    privateKey,
		KeyPassphrase: keyPassphrase,
		Timeout:       15,
	}
	return s.doSSHCheck(srv)
}

// ExecCommand 在指定服务器执行单条命令
func (s *ServerService) ExecCommand(id, command string) (stdout, stderr string, err error) {
	srv, err := s.GetDecrypted(id)
	if err != nil {
		return "", "", err
	}
	client, err := s.buildSSHClient(srv)
	if err != nil {
		return "", "", err
	}
	defer client.Close()
	sess, err := client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("创建会话失败: %w", err)
	}
	defer sess.Close()
	var outBuf, errBuf bytes.Buffer
	sess.Stdout = &outBuf
	sess.Stderr = &errBuf
	runErr := sess.Run(command)
	return outBuf.String(), errBuf.String(), runErr
}

// ===== 内部辅助方法 =====

func (s *ServerService) updateConnInfo(id string, connStatus string, latencyMs int64, info map[string]interface{}) error {
	updates := map[string]interface{}{
		"conn_status":       connStatus,
		"conn_latency_ms":   latencyMs,
		"last_check_time":   time.Now(),
		"updated_at":        time.Now(),
	}
	if info != nil {
		if osStr, ok := info["os"].(string); ok {
			updates["os"] = osStr
		}
		if archStr, ok := info["arch"].(string); ok {
			updates["arch"] = archStr
		}
		if versionStr, ok := info["version"].(string); ok {
			updates["version"] = versionStr
		}
		if v, ok := info["cpuCores"].(int); ok {
			updates["cpu_cores"] = v
		} else if v, ok := info["cpuCores"].(float64); ok {
			updates["cpu_cores"] = int(v)
		}
		if v, ok := info["memoryGB"].(float64); ok {
			updates["memory_gb"] = v
		}
		if v, ok := info["diskGB"].(float64); ok {
			updates["disk_gb"] = v
		}
	}
	return database.DB.Model(&model.Server{}).Where("server_id = ?", id).Updates(updates).Error
}

// buildSSHClient 根据认证方式构造 SSH Client
func (s *ServerService) buildSSHClient(srv *model.Server) (*ssh.Client, error) {
	var auth []ssh.AuthMethod
	if srv.AuthType == model.ServerAuthKey {
		key := strings.TrimSpace(srv.PrivateKey)
		if key == "" {
			return nil, errors.New("私钥为空")
		}
		var signer ssh.Signer
		var err error
		if srv.KeyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(key), []byte(srv.KeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(key))
		}
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	} else {
		if srv.Password == "" {
			return nil, errors.New("密码为空")
		}
		auth = append(auth, ssh.Password(srv.Password))
	}
	timeout := time.Duration(srv.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	config := &ssh.ClientConfig{
		User:            srv.Username,
		Auth:            auth,
		Timeout:         timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	addr := net.JoinHostPort(srv.Host, strconv.Itoa(srv.Port))
	return ssh.Dial("tcp", addr, config)
}

// doSSHCheck 执行连接测试，并采集 OS / CPU / 内存 等信息
func (s *ServerService) doSSHCheck(srv *model.Server) (map[string]interface{}, error) {
	client, err := s.buildSSHClient(srv)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer client.Close()

	info := map[string]interface{}{}
	// 执行多条命令收集信息
	commands := map[string]string{
		"uname":    "uname -s",
		"uname_m":  "uname -m",
		"release":  "(cat /etc/os-release 2>/dev/null | head -n 5) || uname -sr",
		"cpu":      "nproc 2>/dev/null || grep -c '^processor' /proc/cpuinfo",
		"mem_kb":   "awk '/MemTotal/ {print $2}' /proc/meminfo",
		"disk_gb":  "df -Pk / 2>/dev/null | awk 'NR==2 {print $2}'",
	}
	results := map[string]string{}
	for k, cmd := range commands {
		out, _, runErr := runOnce(client, cmd)
		if runErr == nil {
			results[k] = strings.TrimSpace(out)
		}
	}
	info["os"] = results["uname"]
	info["arch"] = results["uname_m"]
	if results["release"] != "" {
		info["version"] = results["release"]
	}
	if n, err := strconv.Atoi(results["cpu"]); err == nil && n > 0 {
		info["cpuCores"] = n
	}
	if memKb, err := strconv.ParseFloat(results["mem_kb"], 64); err == nil && memKb > 0 {
		info["memoryGB"] = round2(memKb / 1024 / 1024)
	}
	if diskKb, err := strconv.ParseFloat(results["disk_gb"], 64); err == nil && diskKb > 0 {
		info["diskGB"] = round2(diskKb / 1024 / 1024)
	}
	return info, nil
}

func runOnce(client *ssh.Client, cmd string) (string, string, error) {
	sess, err := client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer sess.Close()
	var out, errB bytes.Buffer
	sess.Stdout = &out
	sess.Stderr = &errB
	runErr := sess.Run(cmd)
	return out.String(), errB.String(), runErr
}

func round2(v float64) float64 {
	n := int64(v*100 + 0.5)
	return float64(n) / 100
}

// ParseOSRelease 解析 NAME=xxx VERSION=xxx 格式字符串
func ParseOSRelease(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	name := ""
	version := ""
	reName := regexp.MustCompile(`(?i)^NAME\s*=\s*"?([^""]+)"?`)
	reVer := regexp.MustCompile(`(?i)^VERSION\s*=\s*"?([^""]+)"?`)
	for _, l := range lines {
		if sm := reName.FindStringSubmatch(l); sm != nil && name == "" {
			name = sm[1]
		}
		if sm := reVer.FindStringSubmatch(l); sm != nil && version == "" {
			version = sm[1]
		}
	}
	if name != "" && version != "" {
		return name + " " + version
	}
	if name != "" {
		return name
	}
	return strings.TrimSpace(raw)
}

// Stats 返回服务器统计
func (s *ServerService) Stats() (map[string]interface{}, error) {
	var total int64
	database.DB.Model(&model.Server{}).Count(&total)
	var active, inactive, connected, failed int64
	database.DB.Model(&model.Server{}).Where("status = ?", model.ServerStatusActive).Count(&active)
	database.DB.Model(&model.Server{}).Where("status = ?", model.ServerStatusInactive).Count(&inactive)
	database.DB.Model(&model.Server{}).Where("conn_status = ?", model.ServerConnOK).Count(&connected)
	database.DB.Model(&model.Server{}).Where("conn_status = ?", model.ServerConnFail).Count(&failed)
	return map[string]interface{}{
		"total":     total,
		"active":    active,
		"inactive":  inactive,
		"connected": connected,
		"failed":    failed,
	}, nil
}
