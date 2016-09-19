package store

import (
	"sync"
)

type Store interface {
	Get(key interface{}, fn func() interface{}) interface{}
	Del(key interface{})
}

type Memory struct {
	mu     sync.RWMutex
	values map[interface{}]interface{}
}

var _ Store = (*Memory)(nil)

func New() *Memory {
	return &Memory{
		values: make(map[interface{}]interface{}),
	}
}

func (s *Memory) Get(key interface{}, fn func() interface{}) interface{} {
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

func (s *Memory) Del(key interface{}) {
	s.mu.Lock()
	delete(s.values, key)
	s.mu.Unlock()
}

type placeHolder struct {
	mu    sync.RWMutex
	value interface{}
}

func (p *placeHolder) get() interface{} {
	p.mu.RLock()
	v := p.value
	p.mu.RUnlock()
	return v
}

func getValue(v interface{}) interface{} {
	if p, ok := v.(*placeHolder); ok {
		return p.get()
	}
	return v
}
