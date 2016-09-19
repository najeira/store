package store

import (
	"sync"
)

type Store struct {
	values map[interface{}]interface{}
	mu     sync.RWMutex
}

func New() *Store {
	return &Store{
		values: make(map[interface{}]interface{}),
	}
}

func (s *Store) Get(key interface{}, fn func() interface{}) interface{} {
	// lock to read
	s.mu.RLock()
	v, ok := s.values[key]
	s.mu.RUnlock()
	if ok {
		return getValue(v)
	}

	// lock to write
	s.mu.Lock()
	v, ok = s.values[key]
	if ok {
		s.mu.Unlock()
		return getValue(v)
	}

	// lock placeholder to wait new value
	p := &placeHolder{}
	p.mu.Lock()
	defer p.mu.Unlock()

	// add placeholder to map then others wait on the placeholder
	s.values[key] = p
	s.mu.Unlock()

	// get new value
	p.value = fn()

	// store the value
	s.mu.Lock()
	s.values[key] = p.value
	s.mu.Unlock()

	// return and unlock placeholder
	return p.value
}

func (s *Store) Del(key interface{}) {
	s.mu.Lock()
	delete(s.values, key)
	s.mu.Unlock()
}

type placeHolder struct {
	value interface{}
	mu    sync.RWMutex
}

func getValue(v interface{}) interface{} {
	if ver, ok := v.(*placeHolder); ok {
		ver.mu.RLock()
		ret := ver.value
		ver.mu.RUnlock()
		return ret
	}
	return v
}
