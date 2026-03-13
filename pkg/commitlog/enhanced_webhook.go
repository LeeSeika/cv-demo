package commitlog

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/btree"
	"github.com/rs/zerolog/log"
)

type EnhancedCommitLogWebhook struct {
	*CommitLog
}

func NewEnhancedCommitLogWebhook(opts LogOptions) (*EnhancedCommitLogWebhook, error) {
	err := os.MkdirAll(opts.logDir, 0o755)
	if err != nil {
		return nil, err
	}
	fileSet, err := newFileSetFromLogForEnhancedWebhook(opts)
	if err != nil {
		return nil, err
	}
	innerLog := &CommitLog{
		fileSet: fileSet,
	}

	return &EnhancedCommitLogWebhook{
		CommitLog: innerLog,
	}, nil
}

func newFileSetFromLogForEnhancedWebhook(opts LogOptions) (*FileSet, error) {
	segments := btree.NewG(BTreeDegree, segmentLess)
	indexes := btree.NewG(BTreeDegree, indexLess)

	filepath.WalkDir(opts.logDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)

		switch ext {
		case SegmentFileNameExt:
			// xxx.log
			segment, err := OpenSegment(path, opts.logMaxBytes)
			if err != nil {
				return err
			}

			offset := segment.StartingOffset()
			segments.ReplaceOrInsert(&segmentNode{
				offset:  offset,
				segment: segment,
			})
		case IndexFileNameExt:
			// xxx.index
			index, err := OpenIndex(path)
			if err != nil {
				return err
			}

			offset := index.StartingOffset()
			indexes.ReplaceOrInsert(&indexNode{
				offset: offset,
				index:  index,
			})
		default:
			// ignore unknown files
			return nil
		}

		return nil
	})

	if segments.Len() != 1 || indexes.Len() != 1 {
		log.Error().Msg("expected exactly one segment and one index")
		return nil, fmt.Errorf("expected exactly one segment and one index, got %d segments and %d indexes", segments.Len(), indexes.Len())
	}

	// pair up the index and segments (there should be an index per segment)
	closed := btree.NewG(BTreeDegree, entryLess)
	segmentSlice := allNodes(segments)
	for _, segmentNode := range segmentSlice {
		offset := segmentNode.offset
		foundIdxNode, found := indexes.Delete(&indexNode{offset: offset})
		if !found {
			return nil, fmt.Errorf("no index found for segment at offset: %d", offset)
		}

		closed.ReplaceOrInsert(&entryNode{
			offset: offset,
			entry: &entry{
				index:   foundIdxNode.index,
				segment: segmentNode.segment,
			},
		})
	}

	closedNodes := allNodes(closed)
	var active *entry
	if len(closedNodes) != 1 {
		log.Error().Msg("expected exactly one closed entry")
		return nil, fmt.Errorf("expected exactly one closed entry, got %d", len(closedNodes))
	}
	lastEntry := closedNodes[len(closedNodes)-1]
	if lastEntry.entry.index.mode != AccessModeReadOnly {
		log.Error().Msg("expected last index to be in read-only mode")
		return nil, fmt.Errorf("expected last index to be in read-only mode, got %v", lastEntry.entry.index.mode)
	}
	// force change to read write mode
	lastEntry.entry.index.mode = AccessModeReadWrite
	active = &entry{
		index:   lastEntry.entry.index,
		segment: lastEntry.entry.segment,
	}

	return &FileSet{
		active: active,
		closed: btree.NewG(BTreeDegree, entryLess), // no node is closed, only one active node
		opts:   opts,
	}, nil
}
