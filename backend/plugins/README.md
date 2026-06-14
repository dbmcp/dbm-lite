# 插件目录（v0.2 预留）

运行时动态加载的插件目录，遵循 `pkg/plugin/plugin.go` 定义的接口。

## 插件分类

| 目录 | 功能 | 版本 |
|------|------|------|
| `backup/` | 数据备份 / 恢复插件 | v0.2 |
| `review/` | SQL 审核规则插件 | v0.2 |
| `ops/` | 运维脚本插件 | v0.2 |

## 插件接口（pkg/plugin/plugin.go）

每个插件必须实现：

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Init() error
    Close() error
}
```

## 加载机制（v0.2 实现）

- 启动时扫描 `plugins/*/plugin.so`
- 使用 Go `plugin` 包动态加载
- 插件通过 `Manager.Register()` 注册
- 健康检查失败的插件会被禁用并记录日志
