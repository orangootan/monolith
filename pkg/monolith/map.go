package monolith

import "sync"

type SyncMap[K comparable, V any] struct {
	m    map[K]V
	lock *sync.RWMutex
}

func NewSyncMap[K comparable, V any]() SyncMap[K, V] {
	var lock sync.RWMutex
	return SyncMap[K, V]{
		m:    make(map[K]V),
		lock: &lock,
	}
}

func (sm *SyncMap[K, V]) put(key K, value V) {
	sm.lock.Lock()
	sm.m[key] = value
	sm.lock.Unlock()
}

func (sm *SyncMap[K, V]) get(key K) (V, bool) {
	sm.lock.RLock()
	value, ok := sm.m[key]
	sm.lock.RUnlock()
	return value, ok
}

func (sm *SyncMap[K, V]) delete(key K) {
	sm.lock.Lock()
	delete(sm.m, key)
	sm.lock.Unlock()
}
