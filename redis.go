package store

import (
	"fmt"
	"sync"

	"github.com/najeira/conv"
	"gopkg.in/redis.v4"
)

type Redis struct {
	client redisClient
	key    string

	mu           sync.RWMutex
	placeHolders map[string]*placeHolder
}

func NewRedis(client *redis.Client, key string) *Redis {
	return &Redis{
		client: &redisClientWrapper{
			client: client,
			key:    key,
		},
		placeHolders: make(map[string]*placeHolder),
	}
}

func (s *Redis) Fetch(field string, fn func() string) (string, error) {
	value, err := s.client.Get(field)
	if err == nil {
		// no more needs placeholder
		s.deletePlaceHolder(field)
		return value, nil
	}

	// lock to write
	s.mu.Lock()
	p, ok := s.placeHolders[field]
	if ok {
		s.mu.Unlock()
		return getValueAndError(p)
	}

	// lock placeholder to wait new value
	p = &placeHolder{}
	p.mu.Lock()
	defer p.mu.Unlock()

	// add placeholder to map then others wait on the placeholder
	s.placeHolders[field] = p
	s.mu.Unlock()

	// get new value
	value = conv.String(fn())

	// store the value
	s.client.Set(field, value)

	// return and unlock placeholder
	p.value = value
	return value, nil
}

func (s *Redis) Get(field string) (string, error) {
	v, err := s.client.Get(field)
	if err != nil && err != redis.Nil {
		return v, err
	}
	return v, nil
}

func (s *Redis) Set(field string, value string) error {
	return s.client.Set(field, conv.String(value))
}

func (s *Redis) Del(fields ...string) (int64, error) {
	return s.client.Del(fields...)
}

func (s *Redis) Incr(field string, incr int64) (int64, error) {
	return s.client.Incr(field, incr)
}

func (s *Redis) IncrF(field string, incr float64) (float64, error) {
	return s.client.IncrF(field, incr)
}

func (s *Redis) Decr(field string, decr int64) (int64, error) {
	return s.client.Incr(field, 0-decr)
}

func (s *Redis) Clear() error {
	return s.client.Clear()
}

func (s *Redis) deletePlaceHolder(field string) {
	s.mu.RLock()
	_, ok := s.placeHolders[field]
	s.mu.RUnlock()
	if ok {
		s.mu.Lock()
		delete(s.placeHolders, field)
		s.mu.Unlock()
	}
}

type redisClient interface {
	Get(field string) (string, error)
	Set(field string, value string) error
	Del(fields ...string) (int64, error)
	Incr(field string, incr int64) (int64, error)
	IncrF(field string, incr float64) (float64, error)
	Clear() error
}

type redisClientWrapper struct {
	client *redis.Client
	key    string
}

func (r *redisClientWrapper) Get(field string) (string, error) {
	return r.client.HGet(r.key, field).Result()
}

func (r *redisClientWrapper) Set(field string, value string) error {
	return r.client.HSet(r.key, field, value).Err()
}

func (r *redisClientWrapper) Del(fields ...string) (int64, error) {
	return r.client.HDel(r.key, fields...).Result()
}

func (r *redisClientWrapper) Incr(field string, incr int64) (int64, error) {
	n, err := r.client.HIncrBy(r.key, field, incr).Result()
	return n, err
}

func (r *redisClientWrapper) IncrF(field string, incr float64) (float64, error) {
	n, err := r.client.HIncrByFloat(r.key, field, incr).Result()
	return n, err
}

func (r *redisClientWrapper) Clear() error {
	_, err := r.client.Del(r.key).Result()
	return err
}

func getValueAndError(p *placeHolder) (string, error) {
	if p == nil {
		return "", nil
	}
	if v := p.get(); v == nil {
		return "", nil
	} else if err, ok := v.(error); ok {
		return "", err
	} else if str, ok := v.(string); ok {
		return str, nil
	} else {
		return "", fmt.Errorf("invalid type %T", v)
	}
}
