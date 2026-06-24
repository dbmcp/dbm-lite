/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 */

package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 常量
const (
	defaultDir        = "logs"
	maxFileSize       = 50 * 1024 * 1024 // 单个日志文件最大 50MB
	maxFilesPerSource = 10                // 每个来源保留最近 10 个文件
)

type DailyWriter struct {
	mu      sync.Mutex
	dir     string
	source  string
	curDate string
	file    *os.File
	curSize int64
}

func NewBackend(dir ...string) *DailyWriter {
	d := defaultDir
	if len(dir) > 0 && dir[0] != "" { d = dir[0] }
	return &DailyWriter{dir: d, source: "backend"}
}

func NewFrontend(dir ...string) *DailyWriter {
	d := defaultDir
	if len(dir) > 0 && dir[0] != "" { d = dir[0] }
	return &DailyWriter{dir: d, source: "frontend"}
}

func (w *DailyWriter) ensureReady() error {
	now := time.Now().Format("2006-01-02")
	if w.file != nil && w.curDate == now && w.curSize < maxFileSize { return nil }
	if w.file != nil { _ = w.file.Close(); w.file = nil }
	if err := os.MkdirAll(w.dir, 0o755); err != nil { return err }
	cleanOldFiles(w.dir, w.source)
	fname := fmt.Sprintf("%s-%s.log", w.source, now)
	f, err := os.OpenFile(filepath.Join(w.dir, fname), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil { return err }
	if info, statErr := f.Stat(); statErr == nil && info != nil { w.curSize = info.Size() }
	w.file = f
	w.curDate = now
	return nil
}

func (w *DailyWriter) Write(p []byte) (int, error) {
	if p == nil || len(p) == 0 { return 0, nil }
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.ensureReady(); err != nil { return 0, err }
	n, err := w.file.Write(p)
	if err == nil { w.curSize += int64(n) }
	return n, err
}

func (w *DailyWriter) WriteLine(line string) error {
	if !strings.HasSuffix(line, "\n") { line = line + "\n" }
	_, err := w.Write([]byte(fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), line)))
	return err
}

func (w *DailyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil { err := w.file.Close(); w.file = nil; return err }
	return nil
}

func cleanOldFiles(dir string, source string) {
	// 保留最近 maxFilesPerSource 份（按文件名排序即可，按日期）
	entries, err := os.ReadDir(dir)
	if err != nil { return }
	var matched []string
	prefix := source + "-"
	for _, e := range entries {
		if e.IsDir() { continue }
		name := e.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".log") {
			matched = append(matched, name)
		}
	}
	if len(matched) <= maxFilesPerSource { return }
	// 按文件名排序（日期字符串），删除最旧的
	for i := 0; i < len(matched)-1; i++ {
		for j := i + 1; j < len(matched); j++ {
			if matched[i] > matched[j] { matched[i], matched[j] = matched[j], matched[i] }
		}
	}
	removeCount := len(matched) - maxFilesPerSource
	for i := 0; i < removeCount; i++ {
		_ = os.Remove(filepath.Join(dir, matched[i]))
	}
}

// 全局实例
var (
	backendLogger  = NewBackend()
	frontendLogger = NewFrontend()
)

func Backend() *DailyWriter  { return backendLogger }
func Frontend() *DailyWriter { return frontendLogger }

// Print 形式的辅助
func BackendLog(format string, args ...interface{}) {
	_ = backendLogger.WriteLine(fmt.Sprintf(format, args...))
}
func FrontendLog(format string, args ...interface{}) {
	_ = frontendLogger.WriteLine(fmt.Sprintf(format, args...))
}
