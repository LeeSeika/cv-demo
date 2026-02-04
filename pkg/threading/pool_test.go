package threading

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPool(t *testing.T) {
	t.Run("valid worker number", func(t *testing.T) {
		pool := NewPool(5, false)
		assert.NotNil(t, pool)
		assert.Equal(t, 5, pool.workerNum)
		assert.False(t, pool.allowOverflow)
		assert.NotNil(t, pool.taskChan)
		assert.NotNil(t, pool.mu)
		assert.NotNil(t, pool.closed)
		assert.NotNil(t, pool.closeOnce)
		assert.False(t, pool.closed.Load())
	})

	t.Run("panic with invalid worker number", func(t *testing.T) {
		assert.Panics(t, func() {
			NewPool(0, false)
		})

		assert.Panics(t, func() {
			NewPool(-1, false)
		})
	})

	t.Run("with overflow allowed", func(t *testing.T) {
		pool := NewPool(3, true)
		assert.True(t, pool.allowOverflow)
	})
}

func TestPool_Run(t *testing.T) {
	t.Run("starts correct number of workers", func(t *testing.T) {
		pool := NewPool(3, false)

		// Start the pool
		pool.Run()

		// Give workers time to start
		time.Sleep(100 * time.Millisecond)

		// Submit tasks to verify workers are running
		var counter atomic.Int32
		done := make(chan struct{})

		for i := 0; i < 3; i++ {
			pool.Submit(func() {
				counter.Add(1)
				if counter.Load() == 3 {
					close(done)
				}
			})
		}

		select {
		case <-done:
			// Success - all workers processed tasks
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for workers to process tasks")
		}

		pool.Close()
	})
}

func TestPool_Submit(t *testing.T) {
	t.Run("submit task to running pool", func(t *testing.T) {
		pool := NewPool(2, false)
		pool.Run()
		defer pool.Close()

		var result atomic.Int32
		done := make(chan struct{})

		pool.Submit(func() {
			result.Store(42)
			close(done)
		})

		select {
		case <-done:
			assert.Equal(t, int32(42), result.Load())
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for task execution")
		}
	})

	t.Run("submit task to closed pool", func(t *testing.T) {
		pool := NewPool(2, false)
		pool.Run()
		pool.Close()

		var executed atomic.Bool

		// Submit should still execute the task via goroutine fallback
		pool.Submit(func() {
			executed.Store(true)
		})

		// Give time for goroutine to execute
		time.Sleep(100 * time.Millisecond)
		assert.True(t, executed.Load())
	})

	t.Run("submit with overflow not allowed", func(t *testing.T) {
		pool := NewPool(1, false) // single worker, no overflow
		pool.Run()
		defer pool.Close()

		// Block the single worker
		workerBlocked := make(chan struct{})
		workerCanContinue := make(chan struct{})

		pool.Submit(func() {
			close(workerBlocked)
			<-workerCanContinue
		})

		// Wait for worker to be blocked
		<-workerBlocked

		// Submit another task - should block since no overflow allowed
		taskExecuted := make(chan struct{})
		go func() {
			pool.Submit(func() {
				close(taskExecuted)
			})
		}()

		// Ensure task doesn't execute immediately
		select {
		case <-taskExecuted:
			t.Fatal("task should not execute immediately when pool is full")
		case <-time.After(100 * time.Millisecond):
			// Expected - task is waiting
		}

		// Unblock the worker
		close(workerCanContinue)

		// Now the queued task should execute
		select {
		case <-taskExecuted:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for queued task to execute")
		}
	})

	t.Run("submit with overflow allowed", func(t *testing.T) {
		pool := NewPool(1, true) // single worker, overflow allowed
		pool.Run()
		defer pool.Close()

		// Block the single worker
		workerBlocked := make(chan struct{})
		workerCanContinue := make(chan struct{})

		pool.Submit(func() {
			close(workerBlocked)
			<-workerCanContinue
		})

		// Wait for worker to be blocked
		<-workerBlocked

		// Submit another task - should execute via goroutine overflow
		var overflowExecuted atomic.Bool
		pool.Submit(func() {
			overflowExecuted.Store(true)
		})

		// Give time for overflow goroutine to execute
		time.Sleep(100 * time.Millisecond)
		assert.True(t, overflowExecuted.Load())

		// Unblock the worker
		close(workerCanContinue)
	})
}

func TestPool_Close(t *testing.T) {
	t.Run("close pool", func(t *testing.T) {
		pool := NewPool(2, false)
		pool.Run()

		assert.False(t, pool.closed.Load())

		pool.Close()

		assert.True(t, pool.closed.Load())

		// Verify channel is closed
		select {
		case _, ok := <-pool.taskChan:
			assert.False(t, ok, "task channel should be closed")
		default:
			t.Fatal("task channel should be closed and readable")
		}
	})

	t.Run("close pool multiple times", func(t *testing.T) {
		pool := NewPool(2, false)
		pool.Run()

		// Close multiple times should not panic
		pool.Close()
		pool.Close()
		pool.Close()

		assert.True(t, pool.closed.Load())
	})
}

func TestPool_WorkerReset(t *testing.T) {
	t.Run("worker resets after threshold", func(t *testing.T) {
		// This test verifies that workers reset themselves after processing
		// workerResetThreshold tasks. This is harder to test directly,
		// so we'll test indirectly by ensuring the pool continues to work
		// after processing many tasks.

		pool := NewPool(2, false)
		pool.Run()
		defer pool.Close()

		// Submit more tasks than the reset threshold to trigger worker reset
		numTasks := workerResetThreshold + 100
		var counter atomic.Int32
		done := make(chan struct{})

		for i := 0; i < numTasks; i++ {
			pool.Submit(func() {
				if counter.Add(1) >= int32(numTasks) {
					close(done)
				}
			})
		}

		select {
		case <-done:
			assert.Equal(t, int32(numTasks), counter.Load())
		case <-time.After(30 * time.Second):
			t.Fatal("timeout waiting for all tasks to complete")
		}
	})
}

func TestPool_ConcurrentOperations(t *testing.T) {
	t.Run("concurrent submit and close", func(t *testing.T) {
		pool := NewPool(5, true)
		pool.Run()

		var wg sync.WaitGroup
		var executed atomic.Int32

		// Start goroutines that submit tasks
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					pool.Submit(func() {
						executed.Add(1)
					})
				}
			}()
		}

		// Close the pool while submissions are happening
		go func() {
			time.Sleep(50 * time.Millisecond)
			pool.Close()
		}()

		wg.Wait()

		// Give time for all tasks to complete
		time.Sleep(500 * time.Millisecond)

		// Should have executed all 1000 tasks
		assert.Equal(t, int32(1000), executed.Load())
	})
}

func TestPool_ErrorHandling(t *testing.T) {
	t.Run("panic in task should not crash pool", func(t *testing.T) {
		pool := NewPool(2, false)
		pool.Run()
		defer pool.Close()

		// Submit a task that panics
		panicTaskDone := make(chan struct{})
		pool.Submit(func() {
			defer close(panicTaskDone)
			panic("test panic")
		})

		// Wait for panic task to complete
		<-panicTaskDone

		// Pool should still be functional
		var normalTaskExecuted atomic.Bool
		normalTaskDone := make(chan struct{})

		pool.Submit(func() {
			normalTaskExecuted.Store(true)
			close(normalTaskDone)
		})

		select {
		case <-normalTaskDone:
			assert.True(t, normalTaskExecuted.Load())
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for normal task after panic")
		}
	})
}

func TestPool_Performance(t *testing.T) {
	t.Run("performance benchmark", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping performance test in short mode")
		}

		pool := NewPool(10, true)
		pool.Run()
		defer pool.Close()

		numTasks := 10000
		var counter atomic.Int32
		done := make(chan struct{})

		start := time.Now()

		for i := 0; i < numTasks; i++ {
			pool.Submit(func() {
				// Simulate some work
				time.Sleep(time.Microsecond)
				if counter.Add(1) >= int32(numTasks) {
					close(done)
				}
			})
		}

		<-done
		duration := time.Since(start)

		t.Logf("Processed %d tasks in %v (%.2f tasks/sec)",
			numTasks, duration, float64(numTasks)/duration.Seconds())

		assert.Equal(t, int32(numTasks), counter.Load())
	})
}

// Benchmark tests
func BenchmarkPool_Submit(b *testing.B) {
	pool := NewPool(10, false)
	pool.Run()
	defer pool.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			done := make(chan struct{})
			pool.Submit(func() {
				close(done)
			})
			<-done
		}
	})
}

func BenchmarkPool_SubmitWithOverflow(b *testing.B) {
	pool := NewPool(2, true) // Small pool to trigger overflow
	pool.Run()
	defer pool.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			done := make(chan struct{})
			pool.Submit(func() {
				close(done)
			})
			<-done
		}
	})
}
