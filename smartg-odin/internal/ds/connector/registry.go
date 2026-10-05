package connector

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = map[string]Connector{}
)

// Register 注册一个连接器实现，按其 Type() 归入注册表。
func Register(c Connector) {
	if c == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	registry[c.Type()] = c
}

// Get 根据数据源类型获取连接器。
func Get(dsType string) (Connector, error) {
	mu.RLock()
	defer mu.RUnlock()
	if c, ok := registry[dsType]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("unsupported datasource type: %s", dsType)
}

// SupportedTypes 返回所有已注册的数据源类型（按字母序）。
func SupportedTypes() []string {
	mu.RLock()
	defer mu.RUnlock()
	types := make([]string, 0, len(registry))
	for t := range registry {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func init() {
	Register(&SQLiteConnector{})
	Register(&PostgresConnector{})
	Register(&MySQLConnector{})
}
