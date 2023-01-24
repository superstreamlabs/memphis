package server

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestConcurrentMapBasics(t *testing.T) {
	integersMap := NewConcurrentMap[int]()

	if !integersMap.Add("one", 1) {
		t.Fatalf("Add failed")
	}

	if _, ok := integersMap.Load("one"); !ok {
		t.Fatalf("Load failed")
	}

	if integersMap.Add("one", 2) {
		t.Fatalf("Add not unique")
	}

	keys, vals := integersMap.Array()

	if len(keys) != 1 || len(vals) != 1 {
		t.Fatalf("Add not unique")
	}

	integersMap.Delete("one")

	if _, ok := integersMap.Load("one"); ok {
		t.Fatalf("Delete failed")
	}
}

func TestConcurrentMapConcurrency(t *testing.T) {
	integersMap := NewConcurrentMap[int]()

	errCh, waitCh := make(chan struct{}), make(chan struct{})
	go func(errCh, waitCh chan struct{}) {
		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(i int, m *concurrentMap[int], wg *sync.WaitGroup, errCh chan struct{}) {
				m.Add(fmt.Sprintf("%d", i), i)
				time.Sleep(1 * time.Second)
				if !m.Delete(fmt.Sprintf("%d", i)) {
					close(errCh)
				}
				wg.Done()
			}(i, integersMap, &wg, errCh)
		}
		wg.Wait()
		close(waitCh)
	}(errCh, waitCh)

	select {
	case <-waitCh:
		break
	case <-errCh:
		t.Fatalf("Concurrent deletion failed")
	}
}
