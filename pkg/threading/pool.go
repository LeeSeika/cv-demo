package threading

import (
	"sync"
	"sync/atomic"
)

// workerResetThreshold defines how often the stack must be reset. Every
// N requests, by spawning a new goroutine in its place, a worker can reset its
// stack so that large stacks don't live in memory forever. 2^16 should allow
// each goroutine stack to live for at least a few seconds in a typical
// workload (assuming a QPS of a few thousand requests/sec).
const workerResetThreshold = 1 << 16

type Pool struct {
	taskChan      chan func()
	workerNum     int
	allowOverflow bool

	mu        *sync.RWMutex
	closed    *atomic.Bool
	closeOnce *sync.Once
	runOnce   *sync.Once
}

func NewPool(workerNum int, allowOverflow bool) *Pool {
	if workerNum <= 0 {
		panic("worker number must be greater than 0")
	}
	closed := atomic.Bool{}
	closed.Store(false)

	return &Pool{
		taskChan:      make(chan func()),
		workerNum:     workerNum,
		allowOverflow: allowOverflow,
		mu:            &sync.RWMutex{},
		closed:        &closed,
		closeOnce:     &sync.Once{},
		runOnce:       &sync.Once{},
	}
}

func NewPoolAndRun(workerNum int, allowOverflow bool) *Pool {
	pool := NewPool(workerNum, allowOverflow)
	pool.Run()
	return pool
}

func (p *Pool) Run() {
	p.runOnce.Do(func() {
		for range p.workerNum {
			workerGoSafe(p.worker)
		}
	})
}

func (p *Pool) worker() {
	for range workerResetThreshold {
		f, ok := <-p.taskChan
		if !ok {
			return
		}
		f()
	}
	workerGoSafe(p.worker)
}

func (p *Pool) Submit(f func()) {
	if p.closed.Load() {
		// fallback to goroutine if pool is closed
		workerGoSafe(f)
		return
	}
	p.mu.RLock()
	defer p.mu.RUnlock()

	// double check after acquiring lock
	if p.closed.Load() {
		workerGoSafe(f)
		return
	}

	if !p.allowOverflow {
		p.taskChan <- f
		return
	}

	select {
	case p.taskChan <- f:
	default:
		workerGoSafe(f)
	}
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed.Store(true)
	p.closeOnce.Do(func() {
		close(p.taskChan)
	})
}

func workerGoSafe(f func()) {
	GoSafe(f, "panic in worker pool", nil)
}
