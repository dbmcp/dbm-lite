package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const backendPort = "8088"

var resultFile = "D:\\dbm\\dbm-lite\\ssl_test_report.log"

type logWriter struct{}

func (w *logWriter) Write(p []byte) (n int, err error) {
	f, err := os.OpenFile(resultFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		os.Stdout.Write(p)
		return len(p), nil
	}
	defer f.Close()
	f.Write(p)
	os.Stdout.Write(p)
	return len(p), nil
}

var out io.Writer = &logWriter{}

func logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format+"\n", args...)
	out.Write([]byte(msg))
}

type loginResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

type testReq struct {
	DBType    string `json:"dbType"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Database  string `json:"database"`
	SSLMode   string `json:"sslMode"`
	SSLCAFile string `json:"sslCaFile"`
	Charset   string `json:"charset"`
	Timezone  string `json:"timezone"`
}

type testResp struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	LatencyMs int64  `json:"latencyMs"`
	Version   string `json:"version"`
}

func login(baseUrl string) (string, error) {
	body := map[string]string{"username": "admin", "password": "admin123"}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", baseUrl+"/api/auth/login", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logf("  登录响应: HTTP %d", resp.StatusCode)
	logf("  %s", string(respBody))

	var lr loginResp
	if err := json.Unmarshal(respBody, &lr); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	if !lr.Success {
		return "", fmt.Errorf("登录失败: %s", lr.Message)
	}
	return lr.Data.Token, nil
}

func testConnection(baseUrl, token, name string, tr testReq) {
	logf("")
	logf("========== %s ==========", name)
	logf("  dbType=%s, host=%s, port=%d, user=%s, db=%s, sslMode=%s, sslCAFile=%s",
		tr.DBType, tr.Host, tr.Port, tr.Username, tr.Database, tr.SSLMode, tr.SSLCAFile)

	jsonBody, _ := json.Marshal(tr)
	req, err := http.NewRequest("POST", baseUrl+"/api/datasources/test", bytes.NewReader(jsonBody))
	if err != nil {
		logf("  构建请求失败: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 45 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		logf("  请求失败 (耗时 %dms): %v", latency, err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logf("  响应状态: HTTP %d (耗时 %dms)", resp.StatusCode, latency)
	logf("  响应内容: %s", string(respBody))

	var trr testResp
	if err := json.Unmarshal(respBody, &trr); err == nil {
		if trr.Success {
			logf("  结果: 成功 ✓ (版本: %s, 延迟: %dms)", trr.Version, trr.LatencyMs)
		} else {
			logf("  结果: 失败 ✗ (%s)", trr.Message)
		}
	}
}

func waitForBackend(baseUrl string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseUrl + "/api/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return true
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func readBackendLogs() {
	logDir := "D:\\dbm\\dbm-lite\\backend\\logs"
	logf("")
	logf("========== 后端日志检查 ==========")

	entries, err := os.ReadDir(logDir)
	if err != nil {
		logf("  无法读取日志目录: %v", err)
		return
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, _ := e.Info()
		logf("  文件: %s (%d 字节, 修改: %s)", e.Name(), info.Size(), info.ModTime().Format("2006-01-02 15:04:05"))
	}

	today := time.Now().Format("2006-01-02")
	for _, logName := range []string{"backend-" + today + ".log", "server-stderr.log", "server-stdout.log", "server-err.log", "server-out.log"} {
		path := filepath.Join(logDir, logName)
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			continue
		}
		logf("")
		logf("  --- %s (最近 80 行) ---", logName)
		lines := strings.Split(string(data), "\n")
		start := 0
		if len(lines) > 80 {
			start = len(lines) - 80
		}
		for i := start; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) != "" {
				logf("  %4d: %s", i+1, lines[i])
			}
		}
	}
}

func main() {
	os.Remove(resultFile)

	baseUrl := "http://localhost:" + backendPort

	logf("================================================================")
	logf("  TiDB Cloud SSL 连接测试 - %s", time.Now().Format("2006-01-02 15:04:05"))
	logf("  后端: %s", baseUrl)
	logf("================================================================")

	running := false
	resp, err := http.Get(baseUrl + "/api/health")
	if err == nil {
		resp.Body.Close()
		running = resp.StatusCode == 200
	}

	var proc *exec.Cmd
	if !running {
		logf("后端服务未运行，启动中...")
		os.Setenv("DBM_LITE_SERVER_PORT", backendPort)
		os.Setenv("DBM_LITE_DB_TYPE", "sqlite")
		os.Setenv("DBM_LITE_DB_PATH", "./data/dbm-lite.db")
		os.Setenv("DBM_LITE_JWT_SECRET", "dbm-lite-jwt-secret-key-change-in-production-please-12345")
		os.Setenv("DBM_LITE_TOKEN_TTL_SECONDS", "86400")
		os.Setenv("DBM_LITE_AES_KEY", "dbm-lite-aes-key-32bytes-long-please-")
		os.Setenv("DBM_LITE_ADMIN_USERNAME", "admin")
		os.Setenv("DBM_LITE_ADMIN_PASSWORD", "admin123")

		exePath := "D:\\dbm\\dbm-lite\\backend\\dbm-lite.exe"
		proc = exec.Command(exePath)
		proc.Dir = "D:\\dbm\\dbm-lite\\backend"
		proc.Env = os.Environ()
		if err := proc.Start(); err != nil {
			logf("启动后端失败: %v", err)
			return
		}
		logf("后端 PID: %d，等待服务就绪...", proc.Process.Pid)
		defer func() {
			if proc != nil && proc.Process != nil {
				proc.Process.Kill()
			}
		}()

		if !waitForBackend(baseUrl, 20*time.Second) {
			logf("后端服务未能在超时时间内就绪")
			readBackendLogs()
			return
		}
		logf("后端服务就绪 ✓")
	} else {
		logf("后端服务已在运行 ✓")
	}

	logf("")
	logf("========== 登录 ==========")
	token, err := login(baseUrl)
	if err != nil {
		logf("登录失败: %v", err)
		readBackendLogs()
		return
	}
	logf("登录成功 ✓")

	tests := []struct {
		name string
		req  testReq
	}{
		{
			"测试 1: TiDB Cloud + SSL ON, 无 CA 文件 (使用系统 CA)",
			testReq{
				DBType: "tidb", Host: "gateway01.us-west-2.prod.aws.tidbcloud.com",
				Port: 4000, Username: "test_user", Password: "test_password",
				Database: "test", SSLMode: "true", SSLCAFile: "",
			},
		},
		{
			"测试 2: MySQL Local, 无 SSL",
			testReq{
				DBType: "mysql", Host: "localhost", Port: 3306,
				Username: "root", Password: "root",
				Database: "test", SSLMode: "false", SSLCAFile: "",
			},
		},
		{
			"测试 3: TiDB Cloud + SSL ON + CA 文件路径",
			testReq{
				DBType: "tidb", Host: "gateway01.us-west-2.prod.aws.tidbcloud.com",
				Port: 4000, Username: "test_user", Password: "test_password",
				Database: "test", SSLMode: "true",
				SSLCAFile: "/etc/ssl/certs/ca-certificates.crt",
			},
		},
		{
			"测试 4: TiDB Cloud + SSL OFF",
			testReq{
				DBType: "tidb", Host: "gateway01.us-west-2.prod.aws.tidbcloud.com",
				Port: 4000, Username: "test_user", Password: "test_password",
				Database: "test", SSLMode: "false", SSLCAFile: "",
			},
		},
		{
			"测试 5: TiDB Cloud + SSL mode 'require'",
			testReq{
				DBType: "tidb", Host: "gateway01.us-west-2.prod.aws.tidbcloud.com",
				Port: 4000, Username: "test_user", Password: "test_password",
				Database: "test", SSLMode: "require", SSLCAFile: "",
			},
		},
	}

	for _, t := range tests {
		testConnection(baseUrl, token, t.name, t.req)
		time.Sleep(500 * time.Millisecond)
	}

	readBackendLogs()

	logf("")
	logf("================================================================")
	logf("  测试完成 - %s", time.Now().Format("2006-01-02 15:04:05"))
	logf("================================================================")
	fmt.Printf("\n完整测试报告已保存到: %s\n", resultFile)
}
