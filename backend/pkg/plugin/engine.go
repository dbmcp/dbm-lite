/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package plugin

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type PluginInfo struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Path        string            `json:"path"`
	Params      map[string]string `json:"params"`
}

type ExecuteResult struct {
	ExitCode int    `json:"exitCode"`
	Output   string `json:"output"`
	Error    string `json:"error"`
	Duration int64  `json:"durationMs"`
}

type Engine struct {
	pluginDir string
}

func NewEngine(pluginDir string) *Engine {
	if pluginDir == "" {
		pluginDir = "./plugins"
	}
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		pluginDir = "."
	}
	return &Engine{pluginDir: pluginDir}
}

func (e *Engine) List() []PluginInfo {
	var plugins []PluginInfo

	entries, err := os.ReadDir(e.pluginDir)
	if err != nil {
		return plugins
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(e.pluginDir, name)
		info := PluginInfo{
			Name:        name,
			Description: "自定义脚本插件",
			Type:        filepath.Ext(name),
			Path:        fullPath,
			Params:      make(map[string]string),
		}

		metaPath := fullPath + ".json"
		if metaData, err := os.ReadFile(metaPath); err == nil {
			var meta map[string]interface{}
			if json.Unmarshal(metaData, &meta) == nil {
				if desc, ok := meta["description"].(string); ok {
					info.Description = desc
				}
			}
		}

		plugins = append(plugins, info)
	}

	return plugins
}

func (e *Engine) Execute(pluginName string, args []string, env map[string]string) (*ExecuteResult, error) {
	startTime := time.Now()
	result := &ExecuteResult{}

	pluginPath := filepath.Join(e.pluginDir, pluginName)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		result.Error = "插件不存在: " + pluginName
		result.ExitCode = -1
		return result, err
	}

	var cmd *exec.Cmd
	ext := strings.ToLower(filepath.Ext(pluginName))

	if runtime.GOOS == "windows" {
		if ext == ".bat" || ext == ".cmd" {
			cmd = exec.Command(pluginPath)
		} else if ext == ".ps1" {
			cmdArgs := []string{"-ExecutionPolicy", "Bypass", "-File", pluginPath}
			cmdArgs = append(cmdArgs, args...)
			cmd = exec.Command("powershell.exe", cmdArgs...)
		} else {
			cmd = exec.Command(pluginPath, args...)
		}
	} else {
		if ext == ".sh" || ext == "" {
			cmdArgs := []string{pluginPath}
			cmdArgs = append(cmdArgs, args...)
			cmd = exec.Command("bash", cmdArgs...)
		} else if ext == ".py" {
			cmdArgs := []string{pluginPath}
			cmdArgs = append(cmdArgs, args...)
			cmd = exec.Command("python3", cmdArgs...)
		} else {
			cmd = exec.Command(pluginPath, args...)
		}
	}

	cmd.Dir = e.pluginDir
	envList := os.Environ()
	for k, v := range env {
		envList = append(envList, k+"="+v)
	}
	cmd.Env = envList

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(startTime).Milliseconds()
	result.Output = stdout.String()
	result.Error = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Error += ": " + err.Error()
	} else {
		result.ExitCode = 0
	}

	return result, nil
}

