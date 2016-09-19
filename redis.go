package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/najeira/conv"
	"gopkg.in/redis.v4"
)

type RedisOptions struct {
	Client *redis.Client
	Prefix string
	Age    time.Duration
}

type Redis struct {
	client redisClient
	prefix string
	age    time.Duration

	mu           sync.RWMutex
	placeHolders map[string]*placeHolder
}

var _ Store = (*Memory)(nil)

func NewRedis(opt RedisOptions) *Redis {
	return &Redis{
		client:       &redisClientWrapper{opt.Client},
		prefix:       opt.Prefix,
		age:          opt.Age,
		placeHolders: make(map[string]*placeHolder),
	}
}

func (s *Redis) key(key interface{}) string {
	keyStr := conv.String(key)
	if keyStr == "" {
		panic(fmt.Errorf("invalid key %v", key))
	}
	return s.prefix + keyStr
}

func (s *Redis) Fetch(key interface{}, fn func() interface{}) interface{} {
	keyStr := s.key(key)
	value, err := s.client.Get(keyStr)
	if err == nil {
		return value
	}

	// lock to write
	s.mu.Lock()
	p, ok := s.placeHolders[keyStr]
	if ok {
		s.mu.Unlock()
		return p.get()
	}

	// lock placeholder to wait new value
	p = &placeHolder{}
	p.mu.Lock()
	defer p.mu.Unlock()

	// add placeholder to map then others wait on the placeholder
	s.placeHolders[keyStr] = p
	s.mu.Unlock()

	// get new value
	value = conv.String(fn())

	// store the value
	s.client.Set(keyStr, value, s.age)

	// no more needs placeholder
	s.mu.Lock()
	delete(s.placeHolders, keyStr)
	s.mu.Unlock()

	// return and unlock placeholder
	p.value = value
	return p.value
}

func (s *Redis) Get(key interface{}) (interface{}, bool) {
	keyStr := s.key(key)
	value, err := s.client.Get(keyStr)
	return value, (err == nil)
}

func (s *Redis) Set(key interface{}, value interface{}) {
	s.client.Set(s.key(key), conv.String(value), s.age)
}

func (s *Redis) Del(key interface{}) {
	s.client.Del(s.key(key))
}

type redisClient interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
	Del(key string) (bool, error)
}

type redisClientWrapper struct {
	client *redis.Client
}

func (r *redisClientWrapper) Get(key string) (string, error) {
	return r.client.Get(key).Result()
}

func (r *redisClientWrapper) Set(key string, value string, expiration time.Duration) error {
	return r.client.Set(key, value, expiration).Err()
}

func (r *redisClientWrapper) Del(key string) (bool, error) {
	n, err := r.client.Del(key).Result()
	return n > 0, err
}
