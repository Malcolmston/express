package plimit

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCapsConcurrency(t *testing.T) {
	const limit = 3
	const tasks = 20
	l := New(limit)

	var active int64
	var maxActive int64
	var completed int64

	for i := 0; i < tasks; i++ {
		l.Go(func() {
			cur := atomic.AddInt64(&active, 1)
			for {
				old := atomic.LoadInt64(&maxActive)
				if cur <= old || atomic.CompareAndSwapInt64(&maxActive, old, cur) {
					break
				}
			}
			// Tiny sleep to force overlap; well under 20ms.
			time.Sleep(2 * time.Millisecond)
			atomic.AddInt64(&active, -1)
			atomic.AddInt64(&completed, 1)
		})
	}

	l.Wait()

	if completed != tasks {
		t.Fatalf("want %d completed, got %d", tasks, completed)
	}
	if maxActive > limit {
		t.Fatalf("max concurrency %d exceeded limit %d", maxActive, limit)
	}
	if maxActive == 0 {
		t.Fatal("expected some concurrency")
	}
}

func TestSerializeWithLimitOne(t *testing.T) {
	l := New(1)
	var active int64
	var maxActive int64
	var mu sync.Mutex
	var order []int

	for i := 0; i < 5; i++ {
		i := i
		l.Go(func() {
			cur := atomic.AddInt64(&active, 1)
			if cur > atomic.LoadInt64(&maxActive) {
				atomic.StoreInt64(&maxActive, cur)
			}
			time.Sleep(time.Millisecond)
			mu.Lock()
			order = append(order, i)
			mu.Unlock()
			atomic.AddInt64(&active, -1)
		})
	}
	l.Wait()

	if maxActive != 1 {
		t.Fatalf("limit 1 must serialize, got max %d", maxActive)
	}
	if len(order) != 5 {
		t.Fatalf("want 5 executions, got %d", len(order))
	}
}

func TestAllTasksComplete(t *testing.T) {
	l := New(4)
	var sum int64
	for i := 1; i <= 100; i++ {
		i := i
		l.Go(func() {
			atomic.AddInt64(&sum, int64(i))
		})
	}
	l.Wait()

	const want = 100 * 101 / 2
	if sum != want {
		t.Fatalf("want sum %d, got %d", want, sum)
	}
}

func TestRunBlocksAtCapacity(t *testing.T) {
	l := New(2)
	release := make(chan struct{})
	var active int64
	var maxActive int64

	start := func() {
		l.Run(func() {
			cur := atomic.AddInt64(&active, 1)
			for {
				old := atomic.LoadInt64(&maxActive)
				if cur <= old || atomic.CompareAndSwapInt64(&maxActive, old, cur) {
					break
				}
			}
			<-release
			atomic.AddInt64(&active, -1)
		})
	}

	// Two Runs fill capacity and return quickly.
	start()
	start()

	// The third Run should block until we release a slot. Launch it in a
	// goroutine and confirm it has not returned yet.
	returned := make(chan struct{})
	go func() {
		start()
		close(returned)
	}()

	select {
	case <-returned:
		t.Fatal("Run should block while at capacity")
	case <-time.After(10 * time.Millisecond):
		// Expected: still blocked.
	}

	close(release)
	<-returned
	l.Wait()

	if maxActive != 2 {
		t.Fatalf("want max active 2, got %d", maxActive)
	}
}

func TestCounts(t *testing.T) {
	l := New(2)
	release := make(chan struct{})
	var started int64

	for i := 0; i < 6; i++ {
		l.Go(func() {
			atomic.AddInt64(&started, 1)
			<-release
		})
	}

	// Give the scheduler a brief moment to fill the 2 slots.
	deadline := time.Now().Add(15 * time.Millisecond)
	for time.Now().Before(deadline) {
		if atomic.LoadInt64(&started) >= 2 {
			break
		}
		time.Sleep(time.Millisecond)
	}

	if got := l.ActiveCount(); got != 2 {
		t.Fatalf("want ActiveCount 2, got %d", got)
	}
	if got := l.PendingCount(); got != 4 {
		t.Fatalf("want PendingCount 4, got %d", got)
	}

	close(release)
	l.Wait()

	if got := l.ActiveCount(); got != 0 {
		t.Fatalf("want ActiveCount 0 after wait, got %d", got)
	}
	if got := l.PendingCount(); got != 0 {
		t.Fatalf("want PendingCount 0 after wait, got %d", got)
	}
}

func TestNewClampsConcurrency(t *testing.T) {
	l := New(0)
	var maxActive int64
	var active int64
	for i := 0; i < 4; i++ {
		l.Go(func() {
			cur := atomic.AddInt64(&active, 1)
			if cur > atomic.LoadInt64(&maxActive) {
				atomic.StoreInt64(&maxActive, cur)
			}
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&active, -1)
		})
	}
	l.Wait()
	if maxActive != 1 {
		t.Fatalf("New(0) should clamp to 1, got max %d", maxActive)
	}
}
