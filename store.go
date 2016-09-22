package store

import (
	"sync"

	"github.com/najeira/conv"
)

type Memory struct {
	mu     sync.RWMutex
	values map[interface{}]interface{}
}

func NewMemory() *Memory {
	return &Memory{
		values: make(map[interface{}]interface{}),
	}
}

func (s *Memory) Fetch(key interface{}, fn func() interface{}) interface{} {
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

func (s *Memory) Get(key interface{}) (interface{}, bool) {
	s.mu.RLock()
	v, ok := s.values[key]
	s.mu.RUnlock()
	return getValue(v), ok
}

func (s *Memory) Set(key interface{}, value interface{}) {
	s.mu.Lock()
	s.values[key] = value
	s.mu.Unlock()
}

func (s *Memory) Del(key interface{}) {
	s.mu.Lock()
	delete(s.values, key)
	s.mu.Unlock()
}

func (s *Memory) Incr(key interface{}, incr int64) int64 {
	var next int64 = incr

	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.values[key]; ok {
		next = conv.Int(v) + incr
	}
	s.values[key] = next
	return next
}

func (s *Memory) Decr(key interface{}, decr int64) int64 {
	return s.Incr(key, 0-decr)
}

func (s *Memory) IncrF(key interface{}, incr float64) float64 {
	var next float64 = incr

	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.values[key]; ok {
		next = conv.Float(v) + incr
	}
	s.values[key] = next
	return next
}

func (s *Memory) DecrF(key interface{}, decr float64) float64 {
	return s.IncrF(key, 0-decr)
}

func (s *Memory) Clear() {
	s.mu.Lock()
	s.values = make(map[interface{}]interface{})
	s.mu.Unlock()
}

type placeHolder struct {
	mu    sync.RWMutex
	value interface{}
}

func (p *placeHolder) get() interface{} {
	p.mu.RLock()
	value := p.value
	p.mu.RUnlock()
	return value
}

func getValue(v interface{}) interface{} {
	if v == nil {
		return nil
	} else if p, ok := v.(*placeHolder); ok {
		return p.get()
	}
	return v
}
