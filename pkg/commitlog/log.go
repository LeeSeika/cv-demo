package commitlog

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

type Offset uint64

type OffsetRange struct {
	start Offset
	len   Offset
}

type NewClosedEntryPath struct {
	SegmentPath string
	IndexPath   string
}

type CommitLog struct {
	fileSet *FileSet
}

func (c *CommitLog) NextOffset() Offset {
	return Offset(c.fileSet.ActiveIndex().NextOffset())
}

func (c *CommitLog) Read(start Offset, limit *ReadLimit) (*MessageBuf, error) {
	mbf := &MessageBufReader{}
	return c.Reader(mbf, start, *limit)
}

func (c *CommitLog) Reader(reader LogSliceReader, start Offset, rl ReadLimit) (*MessageBuf, error) {
	nextOffset := c.fileSet.ActiveIndex().NextOffset()
	if start >= nextOffset {
		return nil, fmt.Errorf("start offset %d is beyond the end of the log, next offset: %d", start, nextOffset)
	}

	// adjust for the minimum offset (e.g. reader requests an offset that was truncated)
	if minOffset := c.fileSet.MinOffset(); minOffset != nil {
		if start < *minOffset {
			start = *minOffset
		}
	} else {
		return nil, fmt.Errorf("no active index found in the commit log")
	}

	maxBytes := rl.limit

	// find the correct segment
	foundEntry := c.fileSet.Find(uint64(start))
	if foundEntry == nil {
		return nil, fmt.Errorf("no segment found for offset %d", start)
	}
	foundIdx := foundEntry.index
	foundSeg := foundEntry.segment
	segBytes := foundEntry.segment.Size()

	// grab the range from the contained index
	msgRange, err := foundIdx.FindSegmentRange(start, uint32(maxBytes), uint32(segBytes))
	if err != nil {
		return nil, fmt.Errorf("error finding segment range for offset %d: %w", start, err)
	}
	if msgRange == nil || msgRange.Bytes() == 0 {
		return nil, fmt.Errorf("no messages found for offset %d", start)
	}

	return foundSeg.ReadSlice(reader, msgRange.FilePosition(), msgRange.Bytes())
}

func (c *CommitLog) Truncate(offset Offset) error {
	log.Info().
		Uint64("offset", uint64(offset)).
		Msgf("starting to truncate commit log at offset %d", offset)

	// remove index/segment files rolled after the offset
	entriesToRemove := c.fileSet.RemoveAfter(uint64(offset))
	err := removeEntries(entriesToRemove)
	if err != nil {
		return fmt.Errorf("failed to remove entries after truncation: %w", err)
	}

	// truncate the current index
	len, err := c.fileSet.ActiveIndex().Truncate(offset)
	if err != nil || len == 0 {
		log.Warn().
			Err(err).
			Uint64("offset", uint64(offset)).
			Msg("index outside of appended range")
		return nil
	}
	err = c.fileSet.ActiveSegment().Truncate(len)
	if err != nil {
		return fmt.Errorf("failed to truncate active segment: %w", err)
	}

	log.Info().
		Uint64("offset", uint64(offset)).
		Msgf("truncated commit log at offset %d", offset)
	return nil
}

func (c *CommitLog) TrimSegmentsBefore(offset Offset) error {
	entriesToRemove := c.fileSet.RemoveBefore(uint64(offset))
	err := removeEntries(entriesToRemove)
	if err != nil {
		return fmt.Errorf("failed to remove segments before offset %d: %w", offset, err)
	}
	return nil
}

func (c *CommitLog) TrimInactiveSegments() error {
	activeStartOffset := c.fileSet.ActiveIndex().StartingOffset()
	err := c.TrimSegmentsBefore(Offset(activeStartOffset))
	if err != nil {
		return fmt.Errorf("failed to trim inactive segments: %w", err)
	}
	return nil
}

func removeEntries(entries []*entry) error {
	for _, entry := range entries {
		log.Info().
			Uint64("offset", entry.index.StartingOffset()).
			Str("segment_path", entry.segment.path).
			Str("index_path", entry.index.path).
			Msgf("removing segment and index files at offset %d", entry.index.StartingOffset())
		if err := os.Remove(entry.segment.path); err != nil {
			return fmt.Errorf("failed to remove segment file %s: %w", entry.segment.path, err)
		}
		if err := os.Remove(entry.index.path); err != nil {
			return fmt.Errorf("failed to remove index file %s: %w", entry.index.path, err)
		}
	}
	return nil
}

func (c *CommitLog) LastOffset() Offset {
	nextOffset := c.fileSet.ActiveIndex().NextOffset()
	if nextOffset > 0 {
		return Offset(nextOffset - 1)
	}
	return Offset(0)
}

func (c *CommitLog) AppendMsg(payload []byte) (Offset, *NewClosedEntryPath, error) {
	buf := NewMessageBufWithCapacity(len(payload))
	err := buf.Push(payload)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to push message to buffer: %w", err)
	}
	offsetRange, newClosedEntryPath, err := c.Append(buf)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to append message: %w", err)
	}
	if offsetRange.len != 1 {
		return 0, nil, fmt.Errorf("expected to append 1 message, got %d", offsetRange.len)
	}
	return offsetRange.start, newClosedEntryPath, nil
}

func (c *CommitLog) Append(buf *MessageBuf) (OffsetRange, *NewClosedEntryPath, error) {
	startOffset := c.fileSet.ActiveIndex().NextOffset()
	setMessagesOffsets(buf, startOffset)
	return c.AppendWithOffsets(buf)
}

func (c *CommitLog) AppendWithOffsets(buf MessageSet) (OffsetRange, *NewClosedEntryPath, error) {
	bLen := buf.Len()
	if bLen == 0 {
		return OffsetRange{
			start: 0,
			len:   0,
		}, nil, nil
	}

	// check if the given message size exceeded the max bytes
	msgBytes := buf.Bytes()
	if uint64(len(msgBytes)) > c.fileSet.opts.messageMaxBytes {
		return OffsetRange{}, nil, fmt.Errorf("message size %d exceeds max bytes %d", uint64(len(msgBytes)), c.fileSet.opts.messageMaxBytes)
	}

	// first write to the current segment
	startOffset := c.fileSet.ActiveIndex().NextOffset()

	// check to make sure the first message matches the starting offset
	msgs := buf.Messages()
	if len(msgs) > 0 && msgs[0].Offset() != uint64(startOffset) {
		return OffsetRange{}, nil, fmt.Errorf("first message offset %d does not match the starting offset %d", msgs[0].Offset(), startOffset)
	}

	// write the messages to the segment
	var newClosedEntryPath *NewClosedEntryPath
	startPos, err := c.fileSet.ActiveSegment().Append(buf)
	if err != nil {
		switch err {
		case ErrSegmentFull:
			// segment is full, roll to a new segment
			newClosedIndex, newClosedSegment, err := c.fileSet.RollSegment()
			if err != nil {
				return OffsetRange{}, nil, fmt.Errorf("failed to roll segment: %w", err)
			}
			// try appending again to the new segment
			startPos, err = c.fileSet.ActiveSegment().Append(buf)
			if err != nil {
				return OffsetRange{}, nil, fmt.Errorf("failed to append messages to new segment: %w", err)
			}
			log.Info().
				Str("new_closed_segment_path", newClosedSegment.path).
				Str("new_closed_index_path", newClosedIndex.path).
				Msgf("rolled segment at offset %d", startOffset)
			newClosedEntryPath = &NewClosedEntryPath{
				SegmentPath: newClosedSegment.path,
				IndexPath:   newClosedIndex.path,
			}

		default:
			return OffsetRange{}, nil, fmt.Errorf("failed to append messages to segment: %w", err)
		}
	}

	// write the index
	activeIndex := c.fileSet.ActiveIndex()
	indexPosBuf := NewIndexBuf(bLen, activeIndex.StartingOffset())
	pos := startPos
	for _, msg := range msgs {
		indexPosBuf.Push(msg.Offset(), uint32(pos))
		pos += msg.TotalBytes()
	}
	err = activeIndex.Append(indexPosBuf)
	if err != nil {
		return OffsetRange{}, nil, fmt.Errorf("failed to append index: %w", err)
	}

	return OffsetRange{
		start: Offset(startOffset),
		len:   Offset(bLen),
	}, newClosedEntryPath, nil
}

func (c *CommitLog) Flush() error {
	err := c.fileSet.ActiveSegment().Flush()
	if err != nil {
		return fmt.Errorf("failed to flush active segment: %w", err)
	}
	err = c.fileSet.ActiveIndex().Flush()
	if err != nil {
		return fmt.Errorf("failed to flush active index: %w", err)
	}

	log.Info().
		Str("log_dir", c.fileSet.opts.logDir).
		Msgf("flushed commit log")
	return nil
}

func NewCommitLog(opts LogOptions) (*CommitLog, error) {
	err := os.MkdirAll(opts.logDir, 0o755)
	if err != nil {
		return nil, err
	}
	fileSet, err := NewFileSetFromLog(opts)
	if err != nil {
		return nil, err
	}
	return &CommitLog{
		fileSet: fileSet,
	}, nil
}
