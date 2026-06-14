/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package plugin

type Plugin interface {
	Name() string
	Version() string
	Description() string
	Init() error
	Close() error
}

type Manager struct {
	plugins map[string]Plugin
}

func NewManager() *Manager {
	return &Manager{plugins: make(map[string]Plugin)}
}

func (m *Manager) Register(p Plugin) error {
	if err := p.Init(); err != nil {
		return err
	}
	m.plugins[p.Name()] = p
	return nil
}

func (m *Manager) List() []string {
	names := make([]string, 0, len(m.plugins))
	for n := range m.plugins {
		names = append(names, n)
	}
	return names
}

func (m *Manager) Get(name string) Plugin {
	return m.plugins[name]
}

