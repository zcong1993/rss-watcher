package kv

type MemStore struct {
	store map[string]string
}

func NewMemStore() *MemStore {
	return &MemStore{store: make(map[string]string)}
}

func (ms *MemStore) Get(key string) (string, error) {
	v, ok := ms.store[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (ms *MemStore) Set(key string, value string) error {
	ms.store[key] = value
	return nil
}

func (ms *MemStore) Close() error {
	return nil
}

func (ms *MemStore) Name() string {
	return "mem"
}
