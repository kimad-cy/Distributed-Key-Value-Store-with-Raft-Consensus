package store

import (
    "sync"

)

type KVStore struct {
    data map[string]interface{}
    mu   sync.RWMutex
}

type LogEntry struct {
	Term int `json:"term"`
	Command string `json:"command"` 
	Key string `json:"key"`
	Value interface{} `json:"value"`
}



func NewKVStore() *KVStore {
    return &KVStore{
        data: make(map[string]interface{}),
    }
}

func (s *KVStore) Get(key string) (interface{}, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    value, exists := s.data[key]
    return value, exists
}

func (s *KVStore) Set(key string, value interface{}) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
}

func (s *KVStore) Delete(key string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.data, key)
}

func (s *KVStore) Apply(entry LogEntry) {
    switch entry.Command{
    case "SET":
        s.Set(entry.Key, entry.Value)
    case "DELETE":
        s.Delete(entry.Key)
    default:
        // ignore / error
    }
}


// For debugging/log compaction (implement later)
func (s *KVStore) Snapshot() map[string]interface{} {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    snapshot := make(map[string]interface{})
    for k, v := range s.data {
        snapshot[k] = v
    }
    return snapshot
}

func (s *KVStore) Restore(snapshot map[string]interface{}) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.data = make(map[string]interface{})
    for k, v := range snapshot {
        s.data[k] = v
    }
}