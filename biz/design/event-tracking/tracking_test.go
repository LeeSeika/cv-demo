package eventtracking

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/leeseika/cv-demo/pkg/commitlog"
	assert "github.com/stretchr/testify/require"
)

func TestTrackingAPI_TrackTimeoutWhenQueueFull(t *testing.T) {
	writer := &AsyncCommitLogWriter{
		queue:  make(chan TrackingPayload, 1),
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
	writer.queue <- benchmarkTrackingPayload(t, 1)

	api := &TrackingAPI{
		enqueueTimeout: 10 * time.Millisecond,
		writer:         writer,
	}

	err := api.Track(context.Background(), benchmarkTrackingPayload(t, 1))
	assert.ErrorIs(t, err, ErrTrackingEnqueueTimeout)

	stats := api.Stats()
	assert.Equal(t, uint64(0), stats.Accepted)
	assert.Equal(t, uint64(1), stats.Timeouts)
}

func TestTrackingAPI_CloseDrainsQueuedPayloads(t *testing.T) {
	log := newTrackingTestCommitLog(t, benchmarkPayloadJSONSize(t, 3), 10)
	api := NewTrackingAPI(log, APIOptions{
		EnqueueTimeout: time.Second,
	}, WriterOptions{
		QueueCapacity: 64,
		FlushBatch:    10,
		FlushInterval: time.Millisecond,
	})

	const total = 53
	for i := 0; i < total; i++ {
		err := api.Track(context.Background(), benchmarkTrackingPayload(t, 3))
		assert.NoError(t, err)
	}

	err := api.Close()
	assert.NoError(t, err)

	stats := api.Stats()
	assert.Equal(t, uint64(total), stats.Accepted)
	assert.Equal(t, uint64(total), stats.Persisted)
	assert.Equal(t, uint64(0), stats.Timeouts)
	assert.Equal(t, commitlog.Offset(total), log.NextOffset())
}

func TestTrackingAPI_TrackAfterClose(t *testing.T) {
	log := newTrackingTestCommitLog(t, benchmarkPayloadJSONSize(t, 1), 1)
	api := NewTrackingAPI(log, APIOptions{
		EnqueueTimeout: time.Second,
	}, WriterOptions{
		QueueCapacity: 4,
		FlushBatch:    1,
		FlushInterval: time.Millisecond,
	})

	assert.NoError(t, api.Close())

	err := api.Track(context.Background(), benchmarkTrackingPayload(t, 1))
	assert.ErrorIs(t, err, ErrTrackingClosed)
}

func newTrackingTestCommitLog(t testing.TB, payloadBytes int, flushBatch int) *commitlog.CommitLog {
	t.Helper()

	opts := commitlog.DefaultLogOptions(t.TempDir())
	opts.SetSegmentMaxBytes(1 << 30)
	opts.SetIndexMaxItems(10_000_000)

	requiredBatchBytes := uint64((payloadBytes + commitlog.HeaderSize) * flushBatch)
	if requiredBatchBytes > 0 && requiredBatchBytes >= 1_000_000 {
		opts.SetMessageMaxBytes(requiredBatchBytes * 2)
	}

	log, err := commitlog.NewCommitLog(opts)
	assert.NoError(t, err)
	return log
}

func benchmarkTrackingPayload(t testing.TB, eventCount int) TrackingPayload {
	t.Helper()

	events := make([]TrackingEvent, 0, eventCount)
	for i := 0; i < eventCount; i++ {
		events = append(events, TrackingEvent{
			Ev: "click",
			Cd: &ContentDirectory{
				ContentName:     fmt.Sprintf("product-%02d", i),
				ContentCategory: "product-card",
				ContentIDs:      fmt.Sprintf("sku-%04d", i),
				ContentType:     "variant",
				NumItems:        "1",
				Value:           "129.99",
				Currency:        "USD",
				Metadata: map[string]any{
					"source": "homepage",
				},
			},
			Dl:  fmt.Sprintf("https://shop.example.com/products/%02d", i),
			Pid: fmt.Sprintf("pid-%02d", i),
			Pn:  "ProductDetail",
			Ac:  "tap",
			Ca:  "engagement",
			La:  "buy-now",
			Co:  "hero",
			Po: map[string]any{
				"slot": i,
				"row":  1,
			},
			Ot: "button",
			On: "purchase",
			It: 1_710_000_000_000 + int64(i),
		})
	}

	return TrackingPayload{
		ID:                "shop-001",
		SID:               "session-0001",
		FID:               "funnel-checkout",
		FV:                "variant-a",
		Events:            events,
		UTMCampaign:       "spring-sale",
		UTMContent:        "hero-cta",
		UTMMedium:         "cpc",
		UTMSource:         "google",
		UTMTerm:           "running shoes",
		ScreenWidth:       1440,
		ScreenHeight:      900,
		Version:           "1.0.0",
		From:              "web",
		Referrer:          "https://www.google.com/search?q=running+shoes",
		LandingPage:       "https://shop.example.com/products/shoes",
		UserAgent:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
		Timestamp:         time.Unix(1_710_000_000, 0).UTC(),
		SearchString:      "running shoes",
		RefererSearchTerm: "running shoes",
		RefererType:       "search",
		RefererName:       "google",
		LandingPageType:   "product",
		IP:                "203.0.113.42",
		Lang:              "en-US",
		Country:           "US",
		Region:            "CA",
	}
}

func benchmarkPayloadJSONSize(t testing.TB, eventCount int) int {
	t.Helper()

	bytes, err := json.Marshal(benchmarkTrackingPayload(t, eventCount))
	assert.NoError(t, err)
	return len(bytes)
}
