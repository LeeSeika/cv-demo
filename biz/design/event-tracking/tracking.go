package eventtracking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/leeseika/cv-demo/pkg/commitlog"
)

var ErrTrackingEnqueueTimeout = errors.New("tracking enqueue timeout")
var ErrTrackingClosed = errors.New("tracking writer closed")

type TrackingPayload struct {
	ID                string          `json:"id" query:"id"`   // Shop ID
	SID               string          `json:"sid" query:"sid"` // Session ID
	FID               string          `json:"fid" query:"fid"` // Funnel ID
	FV                string          `json:"fv" query:"fv"`   // Funnel variant
	Events            []TrackingEvent `json:"e" query:"e"`     // Array of events
	UTMCampaign       string          `json:"uc" query:"uc"`   // UTM Campaign
	UTMContent        string          `json:"uo" query:"uo"`   // UTM Content
	UTMMedium         string          `json:"um" query:"um"`   // UTM Medium
	UTMSource         string          `json:"us" query:"us"`   // UTM Source
	UTMTerm           string          `json:"ut" query:"ut"`   // UTM Term
	ScreenWidth       int             `json:"sw" query:"sw"`   // Screen width
	ScreenHeight      int             `json:"sh" query:"sh"`   // Screen height
	Version           string          `json:"v" query:"v"`     // Version number
	From              string          `json:"f" query:"f"`     // Event source
	Referrer          string          `json:"rl" query:"rl"`   // Referrer page
	LandingPage       string          `json:"l" query:"l"`     // Landing page
	UserAgent         string          `json:"a" query:"a"`     // User agent
	Timestamp         time.Time       `json:"ts" query:"ts"`   // Event timestamp
	SearchString      string          `json:"wvt" query:"wvt"` // Search string
	RefererSearchTerm string          `json:"rst" query:"rst"`
	RefererType       string          `json:"rt" query:"rt"`
	RefererName       string          `json:"rn" query:"rn"`
	LandingPageType   string          `json:"lpt" query:"lpt"`

	IP      string `json:"ip"`
	Lang    string `json:"lan"`
	Country string `json:"ct"`
	Region  string `json:"rg"`
}

type TrackingEvent struct {
	Ev  string            `json:"ev" query:"ev"`   // Event type
	Cd  *ContentDirectory `json:"cd" query:"cd"`   // Content directory
	Dl  string            `json:"dl" query:"dl"`   // Page URL
	Pid string            `json:"pid" query:"pid"` // Page ID
	Pn  string            `json:"pn" query:"pn"`   // Page name
	Ac  string            `json:"ac" query:"ac"`   // Action
	Ca  string            `json:"ca" query:"ca"`   // Category
	La  string            `json:"la" query:"la"`   // Label
	Co  string            `json:"co" query:"co"`   // Container
	Po  any               `json:"po" query:"po"`   // Position
	Ot  string            `json:"ot" query:"ot"`   // Object Type
	On  string            `json:"on" query:"on"`   // Object Name
	It  int64             `json:"it" query:"it"`   // Event timestamp
}

type ContentDirectory struct {
	ContentName     string         `json:"content_name" query:"content_name"`         // Content name
	ContentCategory string         `json:"content_category" query:"content_category"` // Content category
	ContentIDs      string         `json:"content_ids" query:"content_ids"`           // Content IDs
	ContentType     string         `json:"content_type" query:"content_type"`         // Content type
	NumItems        string         `json:"num_items" query:"num_items"`               // Number of items
	Value           string         `json:"value" query:"value"`                       // Value
	Currency        string         `json:"currency" query:"currency"`                 // Currency
	Metadata        map[string]any `json:"-" query:"-"`                               // Other fields
}

type APIOptions struct {
	EnqueueTimeout time.Duration
}

type WriterOptions struct {
	QueueCapacity int
	FlushBatch    int
	FlushInterval time.Duration
}

type APIStats struct {
	Accepted  uint64
	Timeouts  uint64
	Persisted uint64
	Bytes     uint64
}

type TrackingAPI struct {
	enqueueTimeout time.Duration
	writer         *AsyncCommitLogWriter

	accepted atomic.Uint64
	timeouts atomic.Uint64
}

type AsyncCommitLogWriter struct {
	log           *commitlog.CommitLog
	queue         chan TrackingPayload
	flushBatch    int
	flushInterval time.Duration

	persisted atomic.Uint64
	bytes     atomic.Uint64

	stateMu sync.RWMutex
	closed  bool

	closeOnce sync.Once

	mu       sync.Mutex
	writeErr error

	stopCh chan struct{}
	doneCh chan struct{}
}

func NewTrackingAPI(log *commitlog.CommitLog, apiOpts APIOptions, writerOpts WriterOptions) *TrackingAPI {
	writer := NewAsyncCommitLogWriter(log, writerOpts)
	return &TrackingAPI{
		enqueueTimeout: apiOpts.EnqueueTimeout,
		writer:         writer,
	}
}

func NewAsyncCommitLogWriter(log *commitlog.CommitLog, opts WriterOptions) *AsyncCommitLogWriter {
	if opts.QueueCapacity <= 0 {
		opts.QueueCapacity = 4096
	}
	if opts.FlushBatch <= 0 {
		opts.FlushBatch = 20
	}
	if opts.FlushInterval <= 0 {
		opts.FlushInterval = 5 * time.Millisecond
	}

	writer := &AsyncCommitLogWriter{
		log:           log,
		queue:         make(chan TrackingPayload, opts.QueueCapacity),
		flushBatch:    opts.FlushBatch,
		flushInterval: opts.FlushInterval,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
	}

	go writer.loop()

	return writer
}

func (a *TrackingAPI) Track(ctx context.Context, payload TrackingPayload) error {
	a.writer.stateMu.RLock()
	defer a.writer.stateMu.RUnlock()

	if a.writer.closed {
		return ErrTrackingClosed
	}
	if err := a.writer.Err(); err != nil {
		return err
	}

	timeout := a.enqueueTimeout
	if timeout <= 0 {
		timeout = 100 * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case a.writer.queue <- payload:
		a.accepted.Add(1)
		return nil
	case <-ctx.Done():
		a.timeouts.Add(1)
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ErrTrackingEnqueueTimeout
		}
		return ctx.Err()
	}
}

func (a *TrackingAPI) Close() error {
	return a.writer.Close()
}

func (a *TrackingAPI) Stats() APIStats {
	return APIStats{
		Accepted:  a.accepted.Load(),
		Timeouts:  a.timeouts.Load(),
		Persisted: a.writer.persisted.Load(),
		Bytes:     a.writer.bytes.Load(),
	}
}

func (w *AsyncCommitLogWriter) Close() error {
	w.closeOnce.Do(func() {
		w.stateMu.Lock()
		w.closed = true
		close(w.stopCh)
		w.stateMu.Unlock()
	})
	<-w.doneCh
	return w.Err()
}

func (w *AsyncCommitLogWriter) Err() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writeErr
}

func (w *AsyncCommitLogWriter) setErr(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.writeErr == nil {
		w.writeErr = err
	}
}

func (w *AsyncCommitLogWriter) loop() {
	defer close(w.doneCh)

	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	batch := make([]TrackingPayload, 0, w.flushBatch)

	flush := func() bool {
		if len(batch) == 0 {
			return true
		}
		if err := w.flushBatchPayloads(batch); err != nil {
			w.setErr(err)
			return false
		}
		batch = batch[:0]
		return true
	}

	for {
		select {
		case payload := <-w.queue:
			batch = append(batch, payload)
			if len(batch) >= w.flushBatch {
				if !flush() {
					return
				}
			}
		case <-ticker.C:
			if !flush() {
				return
			}
		case <-w.stopCh:
			for {
				select {
				case payload := <-w.queue:
					batch = append(batch, payload)
					if len(batch) >= w.flushBatch {
						if !flush() {
							return
						}
					}
				default:
					if !flush() {
						return
					}
					return
				}
			}
		}
	}
}

func (w *AsyncCommitLogWriter) flushBatchPayloads(batch []TrackingPayload) error {
	msgBuf := commitlog.NewMessageBufWithCapacity(len(batch) * 2048)
	var totalBytes uint64

	for _, payload := range batch {
		bytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal tracking payload: %w", err)
		}
		totalBytes += uint64(len(bytes))
		if err := msgBuf.Push(bytes); err != nil {
			return fmt.Errorf("push payload into commitlog buffer: %w", err)
		}
	}

	if _, _, err := w.log.Append(msgBuf); err != nil {
		return fmt.Errorf("append tracking payload batch: %w", err)
	}
	if err := w.log.Flush(); err != nil {
		return fmt.Errorf("flush tracking payload batch: %w", err)
	}

	w.persisted.Add(uint64(len(batch)))
	w.bytes.Add(totalBytes)

	return nil
}
