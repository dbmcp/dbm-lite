/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Module: 静态资源性能优化中间件
 *
 * 设计目的：
 *   1. 对 JS/CSS/JSON/HTML/Text 等可压缩响应启用 gzip，降低传输字节数；
 *   2. 对 JS/CSS/图片 等静态资源文件响应头写入 Cache-Control: public, max-age=604800, immutable，
 *      浏览器直接走本地缓存，避免重复请求消耗 HTTP 连接；
 *   3. 完全基于标准库实现，不引入 gin-contrib/gzip 外部依赖。
 */

package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// ====== Gzip 中间件 ======

const gzipMinSize = 1024 // 低于 1KB 不压缩，避免浪费 CPU

var gzipCompressible = map[string]bool{
	"text/html":              true,
	"text/css":               true,
	"text/plain":             true,
	"text/javascript":        true,
	"application/javascript": true,
	"application/json":       true,
	"application/xml":        true,
	"image/svg+xml":          true,
	"font/woff":              true,
	"font/woff2":             true,
}

// gzipWriterPool gzip.Writer 复用池，减少分配开销
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression)
		return w
	},
}

// Gzip 判断客户端支持 gzip 时压缩响应体
// 注意：仅对非 API 请求启用 gzip，避免与前端开发服务器的代理产生兼容性问题
func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检测客户端是否支持 gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// 跳过 API 请求：避免 Vite 开发服务器代理 gzip 响应时出现解码问题
		// API 响应通常是 JSON，体积不大，压缩收益有限
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}

		// 保存原始 response writer
		originalWriter := c.Writer

		// 创建 gzipResponseWriter
		gw := &gzipResponseWriter{
			ResponseWriter: originalWriter,
		}
		// 设置 c.Writer 为 gw，后续中间件和 handler 都通过 gw 写入
		c.Writer = gw

		// defer：确保 gzip writer 被正确关闭
		defer func() {
			// 如果有 gzip writer，关闭它以写入 footer
			if gw.gzipWriter != nil {
				// 先 flush gzip writer 的内部缓冲区
				_ = gw.gzipWriter.Close()
				gw.gzipWriter.Reset(nil)
				gzipWriterPool.Put(gw.gzipWriter)
				gw.gzipWriter = nil
			}
			// 如果实际写入了 gzip 数据，确保 Content-Encoding header 存在
			if gw.wroteGzip {
				// Content-Encoding 已经在 Gzip 中间件开始时设置了
				// 删除 Content-Length（因为压缩后长度改变）
				originalWriter.Header().Del("Content-Length")
			} else {
				// 没有写入 gzip 数据，移除 Content-Encoding header
				originalWriter.Header().Del("Content-Encoding")
			}
		}()

		// 执行后续中间件和 handler
		c.Next()
	}
}

// gzipResponseWriter 延迟初始化 gzip.Writer，只有在第一次写入且
// Content-Type 可压缩时才开启压缩。
type gzipResponseWriter struct {
	gin.ResponseWriter
	gzipWriter  *gzip.Writer
	wroteHeader bool
	wroteGzip   bool
}

func (w *gzipResponseWriter) lazyInit() {
	if w.gzipWriter != nil {
		return
	}

	// 获取 Content-Type
	ct := w.Header().Get("Content-Type")
	// 如果 Content-Type 为空，暂时不压缩（等第一次 Write 时自动检测）
	if ct == "" {
		return
	}

	// 解析主类型
	mainType := ct
	if idx := strings.Index(ct, ";"); idx >= 0 {
		mainType = strings.TrimSpace(ct[:idx])
	}
	mainType = strings.ToLower(mainType)

	// 检查是否可压缩
	if _, ok := gzipCompressible[mainType]; ok {
		// 可压缩，创建 gzip writer
		writer := gzipWriterPool.Get().(*gzip.Writer)
		writer.Reset(w.ResponseWriter)
		w.gzipWriter = writer
		w.wroteGzip = true
		// 设置 Content-Encoding: gzip
		w.Header().Set("Content-Encoding", "gzip")
		// 删除 Content-Length（压缩后长度会改变）
		w.Header().Del("Content-Length")
	} else {
		// 不可压缩，确保移除 Content-Encoding header
		w.Header().Del("Content-Encoding")
	}
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.wroteHeader = true
	w.lazyInit()
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		// 如果 Content-Type 为空，自动检测
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(b))
		}
		w.WriteHeader(http.StatusOK)
	}

	// 如果 gzip writer 存在，写入压缩数据
	if w.gzipWriter != nil {
		return w.gzipWriter.Write(b)
	}
	// 否则直接写入原始数据
	return w.ResponseWriter.Write(b)
}

func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// Flush 刷新 gzip writer 和底层 writer
func (w *gzipResponseWriter) Flush() {
	if w.gzipWriter != nil {
		_ = w.gzipWriter.Flush()
	}
	// 调用底层 writer 的 Flush
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	} else {
		// 尝试通过 gin.ResponseWriter 的 Flush 方法
		w.ResponseWriter.(gin.ResponseWriter).Flush()
	}
}

// ====== 静态资源 Cache-Control 中间件 ======

// StaticCacheControl 对静态资源路径添加强缓存头，默认 7 天，不可变资源加 immutable。
// 匹配规则：请求路径包含 /assets/ 或以 .js/.css/.svg/.png/.jpg/.jpeg/.gif/.ico/.woff/.woff2 结尾。
func StaticCacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := strings.ToLower(c.Request.URL.Path)
		staticSuffixes := []string{
			".js", ".css", ".svg", ".png", ".jpg", ".jpeg",
			".gif", ".ico", ".woff", ".woff2", ".ttf", ".map",
		}
		isStatic := strings.Contains(p, "/assets/")
		if !isStatic {
			for _, suf := range staticSuffixes {
				if strings.HasSuffix(p, suf) {
					isStatic = true
					break
				}
			}
		}
		if isStatic {
			c.Header("Cache-Control", "public, max-age=604800, immutable")
		}
		c.Next()
	}
}
