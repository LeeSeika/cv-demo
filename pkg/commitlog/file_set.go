package commitlog

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/google/btree"
	"github.com/rs/zerolog/log"
)

const BTreeDegree = 32

type entry struct {
	index   *Index
	segment *Segment
}

type entryNode struct {
	offset uint64
	entry  *entry
}

type indexNode struct {
	offset uint64
	index  *Index
}

type segmentNode struct {
	offset  uint64
	segment *Segment
}

var (
	entryLess = func(a, b *entryNode) bool {
		return a.offset < b.offset
	}

	indexLess = func(a, b *indexNode) bool {
		return a.offset < b.offset
	}

	segmentLess = func(a, b *segmentNode) bool {
		return a.offset < b.offset
	}
)

type FileSet struct {
	active *entry
	closed *btree.BTreeG[*entryNode]
	opts   LogOptions
}

func NewFileSetFromLog(opts LogOptions) (*FileSet, error) {
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

	// try to reuse the last index if it is not full. otherwise, open a new index
	// at the correct offset
	closedNodes := allNodes(closed)
	var createNewEntry = true
	var newCreatedEntryBaseOffset = uint64(0)
	var active *entry
	if len(closedNodes) > 0 {
		lastEntry := closedNodes[len(closedNodes)-1]
		if lastEntry.entry.index.mode == AccessModeReadOnly {
			log.Trace().
				Uint64("last_index_offset", lastEntry.offset).
				Msg("last index is readonly, creating new entry")
			createNewEntry = true
			newCreatedEntryBaseOffset = uint64(lastEntry.entry.index.NextOffset())
		} else {
			createNewEntry = false

			log.Trace().
				Uint64("offset", lastEntry.offset).
				Msgf("reusing index and segment at offset %d", lastEntry.offset)
			_, _ = closed.Delete(&entryNode{
				offset: lastEntry.offset,
				entry:  nil,
			})

			lastEntry.entry.index.mode = AccessModeReadWrite
			active = &entry{
				index:   lastEntry.entry.index,
				segment: lastEntry.entry.segment,
			}
		}
	}
	if createNewEntry {
		log.Trace().Msgf("starting new index and segment at offset %d", newCreatedEntryBaseOffset)
		newIndex, err := NewIndex(opts.logDir, newCreatedEntryBaseOffset, opts.indexMaxBytes)
		if err != nil {
			return nil, err
		}
		newSegment, err := NewSegment(opts.logDir, newCreatedEntryBaseOffset, opts.logMaxBytes)
		if err != nil {
			return nil, err
		}
		active = &entry{
			index:   newIndex,
			segment: newSegment,
		}
	}

	// mark all closed indexes as readonly (indexes are not opened as readonly)
	for _, node := range allNodes(closed) {
		err := node.entry.index.SetReadonly()
		if err != nil {
			log.Error().
				Err(err).
				Uint64("offset", node.offset).
				Msgf("error setting index as read-only mode at offset %d", node.offset)
		}
	}

	return &FileSet{
		active: active,
		closed: closed,
		opts:   opts,
	}, nil
}

func (fs *FileSet) ActiveIndex() *Index {
	return fs.active.index
}

func (fs *FileSet) ActiveSegment() *Segment {
	return fs.active.segment
}

func (fs *FileSet) MinOffset() *Offset {
	closedNodes := allNodes(fs.closed)
	if len(closedNodes) > 0 {
		offset := Offset(closedNodes[0].offset)
		return &offset
	}

	if fs.active == nil {
		return nil
	}

	offset := Offset(fs.active.index.StartingOffset())
	return &offset
}

func (fs *FileSet) Find(offset uint64) *entry {
	activeSegmentStartOffset := fs.ActiveIndex().StartingOffset()
	if offset < activeSegmentStartOffset {
		log.Trace().
			Uint64("offset", offset).
			Msgf("index is not contained in the active index for offset %d, finding in closed entries", offset)

		nextBack := findNextBack(fs.closed, &entryNode{
			offset: offset,
			entry:  nil,
		})

		if nextBack != nil {
			log.Trace().
				Uint64("offset", offset).
				Msgf("found index for offset %d in closed entries at %d", offset, nextBack.offset)
			return nextBack.entry
		}
	}
	log.Trace().
		Uint64("offset", offset).
		Msg("using active index for offset")

	return fs.active
}

func (fs *FileSet) RollSegment() (*Index, *Segment, error) {
	err := fs.ActiveIndex().SetReadonly()
	if err != nil {
		return nil, nil, err
	}

	err = fs.ActiveSegment().Flush()
	if err != nil {
		return nil, nil, err
	}

	nextOffset := fs.ActiveIndex().NextOffset()

	log.Info().
		Uint64("next_offset", uint64(nextOffset)).
		Msgf("rolling new segment and index at offset %d", nextOffset)

	newActiveIndex, err := NewIndex(fs.opts.logDir, uint64(nextOffset), fs.opts.indexMaxBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new index: %w", err)
	}
	newActiveSegment, err := NewSegment(fs.opts.logDir, uint64(nextOffset), fs.opts.logMaxBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new segment: %w", err)
	}

	newClosedIndex := fs.active.index
	newClosedSegment := fs.active.segment

	// put the old active index and segment into the closed set
	fs.closed.ReplaceOrInsert(&entryNode{
		offset: fs.active.index.StartingOffset(),
		entry: &entry{
			index:   fs.active.index,
			segment: fs.active.segment,
		},
	})

	fs.active = &entry{
		index:   newActiveIndex,
		segment: newActiveSegment,
	}

	log.Info().
		Uint64("new_active_index_offset", newActiveIndex.StartingOffset()).
		Str("new_active_segment_path", newActiveSegment.path).
		Msgf("new active index at offset %d, segment at %s", newActiveIndex.StartingOffset(), newActiveSegment.path)

	return newClosedIndex, newClosedSegment, nil
}

func (fs *FileSet) RemoveAfter(offset uint64) []*entry {
	if offset >= fs.active.index.StartingOffset() {
		return []*entry{}
	}

	// find the midpoint
	//
	// E.g:
	//    offset = 6
	//    [0 5 10 15] => split key 5
	//
	// midpoint  is then used as the active index/segment pair
	splitKey := findNextBack(fs.closed, &entryNode{
		offset: offset,
		entry:  nil,
	})

	if splitKey == nil {
		log.Warn().
			Uint64("offset", offset).
			Msgf("split key before offset %d not found", offset)
		return []*entry{}
	}

	log.Trace().
		Uint64("offset", offset).
		Uint64("split_key_offset", splitKey.offset).
		Msgf("split key %d found at offset %d", splitKey.offset, offset)

	// split off the range of close segment/index pairs including
	// the midpoint (which will become the new active index/segment)
	splitAt := &entryNode{
		offset: splitKey.offset,
		entry:  nil,
	}
	var after *btree.BTreeG[*entryNode]
	fs.closed, after = splitOff(fs.closed, splitAt, entryLess)

	active, ok := after.Delete(splitAt)
	if !ok {
		log.Warn().Msg("no active entry found after split, this should not happen")
		return nil
	}

	if active.entry.index.StartingOffset() > offset {
		log.Warn().
			Uint64("active_index_offset", active.entry.index.StartingOffset()).
			Uint64("split_offset", offset).
			Msgf("active index starting offset %d is beyond the split offset %d, this should not happen", active.entry.index.StartingOffset(), offset)
		return nil
	}

	log.Trace().
		Uint64("active_index_offset", active.entry.index.StartingOffset()).
		Msgf("setting active to segment starting at offset %d", active.entry.index.StartingOffset())

	// FileSet will start using the new active index/segment pair
	// swap pointers
	fs.active, active.entry = active.entry, fs.active
	active.offset = active.entry.index.StartingOffset()

	// set new active index to read-write mode
	fs.active.index.mode = AccessModeReadWrite

	pairs := make([]*entry, 0, after.Len()+1)
	afterNodes := allNodes(after)
	for _, node := range afterNodes {
		pairs = append(pairs, &entry{
			index:   node.entry.index,
			segment: node.entry.segment,
		})
	}
	pairs = append(pairs, &entry{
		index:   active.entry.index,
		segment: active.entry.segment,
	})

	return pairs
}

func (fs *FileSet) RemoveBefore(offset uint64) []*entry {
	splitKey := findNextBack(fs.closed, &entryNode{
		offset: offset,
		entry:  nil,
	})

	if splitKey == nil {
		log.Warn().
			Uint64("offset", offset).
			Msgf("split key before offset %d not found", offset)
		return []*entry{}
	}

	log.Trace().
		Uint64("split_key_offset", splitKey.offset).
		Uint64("offset", offset).
		Msgf("split key %d found at offset %d", splitKey.offset, offset)

	var suffix *btree.BTreeG[*entryNode]
	fs.closed, suffix = splitOff(fs.closed, &entryNode{
		offset: splitKey.offset,
		entry:  nil,
	}, entryLess)

	// get prefix entries before the split key
	prefix := make([]*entry, 0, fs.closed.Len())
	fs.closed.AscendLessThan(&entryNode{
		offset: splitKey.offset,
		entry:  nil,
	}, func(item *entryNode) bool {
		prefix = append(prefix, item.entry)
		return true
	})

	// put the suffix back into the closed segments
	fs.closed = suffix

	log.Trace().
		Int("removed_entries_count", len(prefix)).
		Uint64("offset", offset).
		Msgf("removed %d entries before offset %d", len(prefix), offset)
	return prefix
}
