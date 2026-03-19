package eventtracking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func BenchmarkTrackingAPI_EndToEnd(b *testing.B) {
	restore := disableTrackingBenchmarkLogging()
	defer restore()

	for _, tc := range []struct {
		name           string
		eventsPerTrack int
		queueCapacity  int
		flushBatch     int
		flushInterval  time.Duration
		enqueueTimeout time.Duration
		parallelism    int
		allowTimeouts  bool
	}{
		{
			name:           "1event_flush1",
			eventsPerTrack: 1,
			queueCapacity:  4096,
			flushBatch:     1,
			flushInterval:  2 * time.Millisecond,
			enqueueTimeout: 200 * time.Millisecond,
			parallelism:    4,
		},
		{
			name:           "1event_flush20",
			eventsPerTrack: 1,
			queueCapacity:  4096,
			flushBatch:     20,
			flushInterval:  2 * time.Millisecond,
			enqueueTimeout: 200 * time.Millisecond,
			parallelism:    4,
		},
		{
			name:           "5events_flush20",
			eventsPerTrack: 5,
			queueCapacity:  4096,
			flushBatch:     20,
			flushInterval:  2 * time.Millisecond,
			enqueueTimeout: 200 * time.Millisecond,
			parallelism:    4,
		},
		{
			name:           "5events_flush100",
			eventsPerTrack: 5,
			queueCapacity:  8192,
			flushBatch:     100,
			flushInterval:  2 * time.Millisecond,
			enqueueTimeout: 200 * time.Millisecond,
			parallelism:    4,
		},
		{
			name:           "5events_flush100_tight_timeout",
			eventsPerTrack: 5,
			queueCapacity:  64,
			flushBatch:     100,
			flushInterval:  10 * time.Millisecond,
			enqueueTimeout: 200 * time.Microsecond,
			parallelism:    8,
			allowTimeouts:  true,
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			payload := benchmarkTrackingPayload(b, tc.eventsPerTrack)
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				b.Fatal(err)
			}

			log := newTrackingTestCommitLog(b, len(payloadBytes), tc.flushBatch)
			api := NewTrackingAPI(log, APIOptions{
				EnqueueTimeout: tc.enqueueTimeout,
			}, WriterOptions{
				QueueCapacity: tc.queueCapacity,
				FlushBatch:    tc.flushBatch,
				FlushInterval: tc.flushInterval,
			})

			var unexpectedErrs atomic.Uint64
			var firstErr atomic.Value

			b.ReportAllocs()
			b.SetBytes(int64(len(payloadBytes)))
			b.SetParallelism(tc.parallelism)
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				ctx := context.Background()
				for pb.Next() {
					err := api.Track(ctx, payload)
					if err == nil {
						continue
					}
					if tc.allowTimeouts && errors.Is(err, ErrTrackingEnqueueTimeout) {
						continue
					}
					if unexpectedErrs.Add(1) == 1 {
						firstErr.Store(err.Error())
					}
				}
			})

			if err := api.Close(); err != nil {
				b.Fatal(err)
			}

			elapsed := b.Elapsed()
			b.StopTimer()

			if unexpected := unexpectedErrs.Load(); unexpected > 0 {
				errMsg, _ := firstErr.Load().(string)
				b.Fatalf("unexpected Track errors: %d, first=%s", unexpected, errMsg)
			}

			stats := api.Stats()
			if got, want := stats.Accepted+stats.Timeouts, uint64(b.N); got != want {
				b.Fatalf("accepted + timeouts mismatch: got %d want %d", got, want)
			}
			if stats.Persisted != stats.Accepted {
				b.Fatalf("persisted mismatch: got %d want %d", stats.Persisted, stats.Accepted)
			}
			if !tc.allowTimeouts && stats.Timeouts > 0 {
				b.Fatalf("unexpected enqueue timeouts: %d", stats.Timeouts)
			}

			seconds := elapsed.Seconds()
			if seconds > 0 {
				b.ReportMetric(float64(stats.Accepted)/seconds, "accepted/s")
				b.ReportMetric(float64(stats.Persisted)/seconds, "persisted/s")
				b.ReportMetric(float64(stats.Timeouts)/seconds, "timeout/s")
				b.ReportMetric(float64(stats.Bytes)/seconds/(1024*1024), "MiB/s")
				b.ReportMetric(float64(stats.Timeouts)/float64(b.N)*100, "timeout-%")
			}
		})
	}
}

func BenchmarkTrackingAPI_FlushBatchSweep(b *testing.B) {
	restore := disableTrackingBenchmarkLogging()
	defer restore()

	const (
		eventsPerTrack = 5
		queueCapacity  = 8192
		parallelism    = 4
	)

	for _, flushBatch := range []int{1, 5, 10, 20, 50, 100, 200} {
		b.Run(fmt.Sprintf("flush_%d", flushBatch), func(b *testing.B) {
			payload := benchmarkTrackingPayload(b, eventsPerTrack)
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				b.Fatal(err)
			}

			log := newTrackingTestCommitLog(b, len(payloadBytes), flushBatch)
			api := NewTrackingAPI(log, APIOptions{
				EnqueueTimeout: 200 * time.Millisecond,
			}, WriterOptions{
				QueueCapacity: queueCapacity,
				FlushBatch:    flushBatch,
				FlushInterval: 2 * time.Millisecond,
			})

			var unexpectedErrs atomic.Uint64
			var firstErr atomic.Value

			b.ReportAllocs()
			b.SetBytes(int64(len(payloadBytes)))
			b.SetParallelism(parallelism)
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				ctx := context.Background()
				for pb.Next() {
					if err := api.Track(ctx, payload); err != nil {
						if unexpectedErrs.Add(1) == 1 {
							firstErr.Store(err.Error())
						}
					}
				}
			})

			if err := api.Close(); err != nil {
				b.Fatal(err)
			}

			elapsed := b.Elapsed()
			b.StopTimer()

			if unexpected := unexpectedErrs.Load(); unexpected > 0 {
				errMsg, _ := firstErr.Load().(string)
				b.Fatalf("unexpected Track errors: %d, first=%s", unexpected, errMsg)
			}

			stats := api.Stats()
			if stats.Accepted != uint64(b.N) {
				b.Fatalf("accepted mismatch: got %d want %d", stats.Accepted, b.N)
			}
			if stats.Persisted != stats.Accepted {
				b.Fatalf("persisted mismatch: got %d want %d", stats.Persisted, stats.Accepted)
			}
			if stats.Timeouts > 0 {
				b.Fatalf("unexpected enqueue timeouts: %d", stats.Timeouts)
			}

			if seconds := elapsed.Seconds(); seconds > 0 {
				b.ReportMetric(float64(stats.Accepted)/seconds, "accepted/s")
				b.ReportMetric(float64(stats.Persisted)/seconds, "persisted/s")
				b.ReportMetric(float64(stats.Bytes)/seconds/(1024*1024), "MiB/s")
			}
		})
	}
}

func disableTrackingBenchmarkLogging() func() {
	prev := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.Disabled)

	return func() {
		zerolog.SetGlobalLevel(prev)
	}
}

func BenchmarkTrackingAPI_SerialEndToEnd(b *testing.B) {
	restore := disableTrackingBenchmarkLogging()
	defer restore()

	payload := benchmarkTrackingPayload(b, 1)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		b.Fatal(err)
	}

	log := newTrackingTestCommitLog(b, len(payloadBytes), 20)
	api := NewTrackingAPI(log, APIOptions{
		EnqueueTimeout: 200 * time.Millisecond,
	}, WriterOptions{
		QueueCapacity: 4096,
		FlushBatch:    20,
		FlushInterval: 2 * time.Millisecond,
	})

	b.ReportAllocs()
	b.SetBytes(int64(len(payloadBytes)))
	b.ResetTimer()

	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		if err := api.Track(ctx, payload); err != nil {
			b.Fatalf("track serial: %v", err)
		}
	}

	if err := api.Close(); err != nil {
		b.Fatalf("close serial benchmark: %v", err)
	}

	elapsed := b.Elapsed()
	b.StopTimer()

	stats := api.Stats()
	if stats.Accepted != uint64(b.N) {
		b.Fatalf("accepted mismatch: got %d want %d", stats.Accepted, b.N)
	}
	if stats.Persisted != uint64(b.N) {
		b.Fatalf("persisted mismatch: got %d want %d", stats.Persisted, b.N)
	}
	if stats.Timeouts > 0 {
		b.Fatalf("unexpected enqueue timeouts: %d", stats.Timeouts)
	}

	if seconds := elapsed.Seconds(); seconds > 0 {
		b.ReportMetric(float64(stats.Accepted)/seconds, "accepted/s")
		b.ReportMetric(float64(stats.Persisted)/seconds, "persisted/s")
		b.ReportMetric(float64(stats.Bytes)/seconds/(1024*1024), "MiB/s")
	}
}
