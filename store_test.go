package store

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	rwmu sync.RWMutex
	mu   sync.Mutex
	tmp  int64
)

type testValuer struct {
	counter int64
}

func (v *testValuer) do() int64 {
	v.counter += 1
	return v.counter
}

func TestMemoryFetch(t *testing.T) {
	var v testValuer
	store := NewMemory()

	for i := 0; i < 3; i++ {
		ret := store.Fetch("key", func() interface{} {
			return v.do()
		})
		retint := ret.(int64)
		if retint != 1 {
			t.Errorf("got %d expect %d", retint, 1)
		}
	}

	store.Del("key")

	for i := 0; i < 3; i++ {
		ret := store.Fetch("key", func() interface{} {
			return v.do()
		})
		retint := ret.(int64)
		if retint != 2 {
			t.Errorf("got %d expect %d", retint, 2)
		}
	}
}

func TestMemoryConcurrency(t *testing.T) {
	concurrency := 1000

	var counter int64 = 0
	store := NewMemory()
	creater := func() interface{} {
		n := atomic.AddInt64(&counter, 1)
		if n != 1 {
			t.Fatalf("%d != 1", n)
		}
		time.Sleep(time.Millisecond * 100)
		n = atomic.AddInt64(&counter, -1)
		if n != 0 {
			t.Fatalf("%d != 0", n)
		}
		return "ok"
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				v := store.Fetch("key", creater)
				if s, _ := v.(string); s != "ok" {
					t.Fatal(s)
				}
			}
		}()
	}
	wg.Wait()
}

func BenchmarkMemory(b *testing.B) {
	b.ReportAllocs()

	concurrency := 1000
	store := NewMemory()
	creater := func() interface{} {
		time.Sleep(time.Millisecond * 100)
		return "ok"
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < b.N; j++ {
				store.Fetch("key", creater)
				if rand.Int63()%1000 == 0 {
					store.Del("key")
				}
			}
		}()
	}
	wg.Wait()
}

func getWithRWMutex() {
	rwmu.RLock()
	if tmp != 0 {
		rwmu.RUnlock()
		return
	}
	rwmu.RUnlock()

	rwmu.Lock()
	if tmp != 0 {
		rwmu.Unlock()
		return
	}

	tmp = 1
	rwmu.Unlock()
}

func getWithMutex() {
	mu.Lock()
	if tmp != 0 {
		mu.Unlock()
		return
	}
	tmp = 1
	mu.Unlock()
}

func clearTestValueWithRWMutex() {
	rwmu.Lock()
	tmp = 0
	rwmu.Unlock()
}

func clearTestValueWithMutex() {
	mu.Lock()
	tmp = 0
	mu.Unlock()
}

func BenchmarkLock(b *testing.B) {
	b.ReportAllocs()
	concurrency := 100
	n := b.N / concurrency

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < n; j++ {
				getWithMutex()
				if rand.Int63()%100 == 0 {
					clearTestValueWithMutex()
				}
			}
		}()
	}
	wg.Wait()
}

func BenchmarkRWLock(b *testing.B) {
	b.ReportAllocs()
	concurrency := 100
	n := b.N / concurrency

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < n; j++ {
				getWithRWMutex()
				if rand.Int63()%100 == 0 {
					clearTestValueWithRWMutex()
				}
			}
		}()
	}
	wg.Wait()
}
