/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Module: 内存缓存工具（元数据接口、低频变更数据）
 *
 * 设计目的：
 *   1. 对数据源列表、数据库树、能力声明等元数据 GET 接口做服务端缓存；
 *   2. 减少 SQLite 查询频率，缩短接口响应时间，快速释放 HTTP 连接槽位，
 *      避免后端阻塞加剧前端 Monaco 资源与业务接口并发竞争；
 *   3. 不引入额外依赖（如 redis），完全基于标准库 + sync.Map。
 */

package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ====== 全局缓存实例 ======

// metaCache 默认 30 分钟 TTL 的通用缓存；可被任意模块 GetMetaCache() 使用。
var metaCache = newMemoryCache(30 * time.Minute)

// GetMetaCache 返回元数据缓存实例，供 Handler 层直接调用
func GetMetaCache() *MemoryCache { return metaCache }

// ====== MemoryCache 实现 ======

type cacheItem struct {
	value    interface{}
	expireAt time.Time
}

// MemoryCache 线程安全的内存缓存，过期自动失效，后台定时清理
type MemoryCache struct {
	store      sync.Map
	defaultTTL time.Duration
	stopCh     chan struct{}
	once       sync.Once
}

func newMemoryCache(defaultTTL time.Duration) *MemoryCache {
	c := &MemoryCache{defaultTTL: defaultTTL, stopCh: make(chan struct{})}
	go c.startCleanup(time.Minute)
	return c
}

// Get 读取缓存；命中返回 (value, true)，未命中返回 (nil, false)
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	if v, ok := c.store.Load(key); ok {
		item := v.(*cacheItem)
		if time.Now().After(item.expireAt) {
			c.store.Delete(key)
			return nil, false
		}
		return item.value, true
	}
	return nil, false
}

// Set 写入缓存；ttl <= 0 时使用默认 TTL
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	c.store.Store(key, &cacheItem{value: value, expireAt: time.Now().Add(ttl)})
}

// Delete 删除单条缓存
func (c *MemoryCache) Delete(key string) { c.store.Delete(key) }

// Clear 清空全部缓存（写操作之后调用）
func (c *MemoryCache) Clear() {
	c.store.Range(func(k, v interface{}) bool {
		c.store.Delete(k)
		return true
	})
}

// Stop 停止后台清理协程
func (c *MemoryCache) Stop() {
	c.once.Do(func() { close(c.stopCh) })
}

func (c *MemoryCache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			c.store.Range(func(k, v interface{}) bool {
				if item, ok := v.(*cacheItem); ok && now.After(item.expireAt) {
					c.store.Delete(k)
				}
				return true
			})
		}
	}
}

// ====== 中间件：通用缓存响应（只读 GET 接口使用） ======

type cachedResponse struct {
	Status int
	Header http.Header
	Body   []byte
}

// CacheReadThrough 对只读 GET 接口做服务端缓存：
// - key = 请求 Path + RawQuery（不包含用户相关信息，适用于全局一致的元数据接口）；
// - 命中：直接返回缓存体；
// - 未命中：执行下游 handler，记录 2xx 响应体至缓存。
//
// 适用：/api/datasources、/api/dataquery/*/tree、/api/dataquery/capabilities 等元数据接口。
func CacheReadThrough(ttl time.Duration) gin.HandlerFunc {
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		key := fmt.Sprintf("resp:%s?%s", c.Request.URL.Path, c.Request.URL.RawQuery)

		if hit, ok := metaCache.Get(key); ok {
			if resp, ok := hit.(*cachedResponse); ok {
				// 回放缓存：只回放 Content-Type 等业务相关 header，排除 Content-Encoding
				// Content-Encoding 由外层 Gzip 中间件动态设置
				for k, vals := range resp.Header {
					// 跳过 Content-Encoding 和 Content-Length，这些由 Gzip 中间件动态设置
					if k == "Content-Encoding" || k == "Content-Length" {
						continue
					}
					for _, v := range vals {
						c.Writer.Header().Add(k, v)
					}
				}
				c.Writer.Header().Set("X-DBM-Cache", "HIT")
				c.Data(resp.Status, resp.Header.Get("Content-Type"), resp.Body)
				c.Abort()
				return
			}
		}

		// 拦截下游响应并缓存
		cw := &cacheResponseWriter{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = cw
		c.Next()

		if cw.status >= 200 && cw.status < 300 && len(cw.body) > 0 {
			// 缓存时：排除 Content-Encoding 和 Content-Length，这些由 Gzip 中间件动态设置
			hdr := cw.Header().Clone()
			hdr.Del("Content-Encoding")
			hdr.Del("Content-Length")
			metaCache.Set(key, &cachedResponse{Status: cw.status, Header: hdr, Body: cw.body}, ttl)
			c.Writer.Header().Set("X-DBM-Cache", "MISS")
		}
	}
}

// cacheResponseWriter 拦截下游写入的响应内容，用于缓存回放
type cacheResponseWriter struct {
	gin.ResponseWriter
	status int
	body   []byte
}

func (w *cacheResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
func (w *cacheResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}
func (w *cacheResponseWriter) WriteString(s string) (int, error) {
	w.body = append(w.body, []byte(s)...)
	return w.ResponseWriter.WriteString(s)
}
