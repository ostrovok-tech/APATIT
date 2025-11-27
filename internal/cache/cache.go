package cache

import "sync"

var TaskDataCache = &TaskCache{}
var AllTasksInfoCache []byte

// TaskCache
// is a cache of TaskStat in JSON
type TaskCache struct {
	mu   sync.RWMutex // RWMutex allows readings
	data []byte
}

// UpdateCache safely updates data in cache
func (c *TaskCache) UpdateCache(data []byte) {
	c.mu.Lock()
	c.data = data
	c.mu.Unlock()
}

// GetFromCache safely reads data from cache
func (c *TaskCache) GetFromCache() []byte {
	c.mu.RLock()
	// Returns copy here to avoid data race
	dataCopy := make([]byte, len(c.data))
	copy(dataCopy, c.data)
	c.mu.RUnlock()
	return dataCopy
}
