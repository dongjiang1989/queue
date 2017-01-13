package queue

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	log "github.com/cihub/seelog"
)

func TestQueue(t *testing.T) {
	log.Flush()
	var wg sync.WaitGroup
	var id int32

	producter := 100
	consumer := 5

	wg.Add(producter)

	q := NewQueue(1024 * 1024)

	for i := 0; i < producter; i++ {
		go func(g int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				t := fmt.Sprintf("Node.%d.%d.%d", g, j, atomic.AddInt32(&id, 1))
				q.Put(t)
			}
		}(i)
	}
	wg.Wait()

	wg.Add(consumer)
	for i := 0; i < consumer; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; {
				_, ok, _ := q.Get()
				if !ok {
					runtime.Gosched()
				} else {
					j++
				}
			}
		}()
	}
	wg.Wait()
	log.Info("Len:", q)

	if q := q.Len(); q != 0 {
		log.Error("Len Error: r.len == 0", q, 0)
	} else {
		log.Info("Len:", q)
	}
}
