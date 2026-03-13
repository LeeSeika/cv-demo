package commitlog

import (
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type EnhancedCommitLogAnalytics struct {
	*CommitLog
}

func NewEnhancedCommitLogAnalytics(opts LogOptions) (*EnhancedCommitLogAnalytics, error) {
	innerLog, err := NewCommitLog(opts)
	if err != nil {
		return nil, err
	}
	return &EnhancedCommitLogAnalytics{
		CommitLog: innerLog,
	}, nil
}

func (c *EnhancedCommitLogAnalytics) AppendByEntryPath(indexPath, segmentPath string) (Offset, error) {
	return c.fileSet.appendEntryFromPath(indexPath, segmentPath)
}

func (fs *FileSet) appendEntryFromPath(indexPath, segmentPath string) (Offset, error) {
	extIdx := filepath.Ext(indexPath)
	if extIdx != IndexFileNameExt {
		return Offset(0), fmt.Errorf("invalid index file path: %s, expected .index extension", indexPath)
	}
	extSeg := filepath.Ext(segmentPath)
	if extSeg != SegmentFileNameExt {
		return Offset(0), fmt.Errorf("invalid segment file path: %s, expected .log extension", segmentPath)
	}

	index, err := OpenIndex(indexPath)
	if err != nil {
		return Offset(0), fmt.Errorf("failed to open index %s: %w", indexPath, err)
	}

	segment, err := OpenSegment(segmentPath, fs.opts.logMaxBytes)
	if err != nil {
		return Offset(0), fmt.Errorf("failed to open segment %s: %w", segmentPath, err)
	}

	fs.closed.ReplaceOrInsert(&entryNode{
		offset: fs.active.index.StartingOffset(),
		entry: &entry{
			index:   fs.ActiveIndex(),
			segment: fs.ActiveSegment(),
		},
	})

	offset := index.StartingOffset()
	fs.active = &entry{
		index:   index,
		segment: segment,
	}

	log.Info().
		Uint64("offset", offset).
		Str("index_path", indexPath).
		Str("segment_path", segmentPath).
		Msgf("appended new entry from path at offset %d", offset)

	return Offset(offset), nil
}
