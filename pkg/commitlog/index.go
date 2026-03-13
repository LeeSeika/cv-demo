package commitlog

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/edsrzf/mmap-go"
	"github.com/rs/zerolog/log"
)

const (
	IndexEntryBytes  uint64 = 8
	IndexFileNameLen uint64 = 20
	IndexFileNameExt string = ".index"
)

var ErrIndexNotFound = errors.New("index not found")

type AccessMode string

const (
	// Only reads are permitted.
	AccessModeReadOnly AccessMode = "read-only"

	// This is the active index and can be read or written to.
	AccessModeReadWrite AccessMode = "read-write"
)

type Index struct {
	file *os.File

	path string

	// mmap
	mmap mmap.MMap

	mode AccessMode

	// next starting byte in index file offset to write
	nextWritePos    uint64
	lastFlushEndPos uint64
	baseOffset      uint64
}

func NewIndex(logDir string, baseOffset uint64, fileBytes uint64) (*Index, error) {
	indexPath := filepath.Join(logDir, fmt.Sprintf("%020d", baseOffset)+IndexFileNameExt)
	file, err := os.OpenFile(
		indexPath,
		os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_APPEND,
		0o644,
	)
	if err != nil {
		return nil, err
	}
	shouldClose := true
	defer func() {
		if shouldClose {
			_ = file.Close()
		}
	}()
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := stat.Size()
	if size == 0 {
		if err := file.Truncate(int64(fileBytes)); err != nil {
			return nil, err
		}
	}

	mmap, err := mmap.Map(file, mmap.RDWR, 0)
	if err != nil {
		return nil, err
	}

	shouldClose = false

	return &Index{
		file:            file,
		path:            indexPath,
		mmap:            mmap,
		mode:            AccessModeReadWrite,
		nextWritePos:    0,
		lastFlushEndPos: 0,
		baseOffset:      baseOffset,
	}, nil
}

func OpenIndex(indexPath string) (*Index, error) {
	file, err := os.OpenFile(
		indexPath,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0o644,
	)
	if err != nil {
		return nil, err
	}
	shouldClose := true
	defer func() {
		if shouldClose {
			_ = file.Close()
		}
	}()

	// get base offset from filename
	filename := filepath.Base(indexPath)
	if uint64(len(filename)) < IndexFileNameLen {
		return nil, errors.New("index file name is invalid")
	}
	baseOffsetStr := filename[:IndexFileNameLen]
	baseOffset, err := strconv.ParseInt(baseOffsetStr, 10, 64)
	if err != nil {
		return nil, err
	}

	indexMMap, err := mmap.Map(file, mmap.RDWR, 0)
	if err != nil {
		return nil, err
	}
	if len(indexMMap) == 0 || uint64(len(indexMMap))%IndexEntryBytes != 0 {
		return nil, fmt.Errorf("index file %d length is invalid", len(indexMMap))
	}

	offset, position := indexToEntry(indexMMap, uint64(len(indexMMap))-IndexEntryBytes)

	var nextWritePosition uint64
	var accessMode AccessMode

	// check if this is a full or partial index
	if offset == 0 && position == 0 {
		// partial index, search for break point
		compareF := func(relOffset uint32, fileOffset uint32) Ordering {
			// find the first unwritten index entry:
			// +#############-----|----------------+
			//  written msgs            empty msgs
			// if the file pos is 0, we're in the empty msgs, go left
			// otherwise, we're in the written msgs, go right
			//
			// NOTE: it is assumed the segment will never start at 0
			// since it contains at least 1 magic byte
			if fileOffset == 0 && relOffset == 0 {
				return OrderingGreater
			} else {
				return OrderingLess
			}
		}
		binSearchResult, err := binarySearchIndex(indexMMap, compareF)
		if err != nil {
			return nil, err
		}
		nextWritePosition = IndexEntryBytes * uint64(binSearchResult)
		accessMode = AccessModeReadWrite
	} else {
		nextWritePosition = uint64(len(indexMMap))
		accessMode = AccessModeReadOnly
	}

	log.Info().
		Str("index_filename", filename).
		Uint64("next_write_position", nextWritePosition).
		Str("access_mode", string(accessMode)).
		Msgf("opened index %s, next write pos %d, mode %s", filename, nextWritePosition, accessMode)

	shouldClose = false

	return &Index{
		file:            file,
		path:            indexPath,
		mmap:            indexMMap,
		mode:            accessMode,
		nextWritePos:    nextWritePosition,
		lastFlushEndPos: nextWritePosition,
		baseOffset:      uint64(baseOffset),
	}, nil
}

func (i *Index) Size() uint64 {
	return uint64(len(i.mmap))
}

func (i *Index) StartingOffset() uint64 {
	return i.baseOffset
}

func (i *Index) NextOffset() Offset {
	if i.nextWritePos == 0 {
		return Offset(i.baseOffset)
	} else {
		offset, _, err := i.ReadEntry((uint64(i.nextWritePos) / IndexEntryBytes) - 1) // get last entry
		if err != nil {
			log.Warn().
				Err(err).
				Uint64("next_write_position", i.nextWritePos).
				Uint64("base_offset", i.baseOffset).
				Msg("error reading last index entry, will return base offset")
			return Offset(i.baseOffset)
		}
		return Offset(offset + 1)
	}
}

func (i *Index) ReadEntry(entryIdx uint64) (Offset, uint32, error) {
	if i.Size() < (entryIdx+1)*IndexEntryBytes {
		return 0, 0, fmt.Errorf("index entry %d out of bounds", entryIdx)
	}

	start := entryIdx * IndexEntryBytes
	offset := binary.LittleEndian.Uint32(i.mmap[start : start+4])
	if offset == 0 && entryIdx > 0 {
		return 0, 0, fmt.Errorf("index entry %d is empty", entryIdx)
	} else {
		position := binary.LittleEndian.Uint32(i.mmap[start+4 : start+8])
		return Offset(i.baseOffset + uint64(offset)), position, nil
	}
}

func (i *Index) SetReadonly() error {
	if i.mode != AccessModeReadOnly {
		i.mode = AccessModeReadOnly

		err := i.Flush()
		if err != nil {
			return err
		}

		// trim un-used entries by reducing mmap view and truncating file
		if i.nextWritePos < uint64(len(i.mmap)) {
			if err := i.file.Truncate(int64(i.nextWritePos)); err != nil {
				filename := fmt.Sprintf("%020d%s", i.baseOffset, IndexFileNameExt)
				log.Warn().
					Err(err).
					Uint64("next_write_position", i.nextWritePos).
					Str("index_filename", filename).
					Msgf("failed to truncate index file %s", filename)
			}
		}
	}
	return nil
}

func (i *Index) Flush() error {
	err := i.mmap.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush index mmap: %w", err)
	}
	i.lastFlushEndPos = i.nextWritePos
	err = i.file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync index file: %w", err)
	}
	return nil
}

func (i *Index) FindSegmentRange(offset Offset, maxBytes uint32, segSize uint32) (*MessageSetRange, error) {
	startIdxPos, err := i.FindIndexPosition(offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find index position for offset %d: %w", offset, err)
	}

	_, startFilePos := indexToEntry(i.mmap, startIdxPos)

	// try to get until the end of the segment
	if segSize-startFilePos < maxBytes {
		log.Trace().
			Uint32("start_file_pos", startFilePos).
			Uint32("seg_size", segSize).
			Uint32("max_bytes", maxBytes).
			Msg("requested range contains the rest of the segment, does not exceed max bytes")
		return &MessageSetRange{
			filePosition: startFilePos,
			bytes:        segSize - startFilePos,
		}, nil
	}

	searchRange := i.mmap[startIdxPos:i.nextWritePos]
	if len(searchRange) == 0 {
		return nil, fmt.Errorf("message exceeded max bytes")
	}

	// find the first index entry that exceeds the max bytes
	compareF := func(_ uint32, pos uint32) Ordering {
		if (pos - startFilePos) < maxBytes {
			return OrderingLess
		} else if (pos - startFilePos) > maxBytes {
			return OrderingGreater
		}
		return OrderingEqual
	}
	endIdxPos, err := binarySearchIndex(searchRange, compareF)
	if err != nil {
		return nil, fmt.Errorf("failed to find index position for offset %d: %w", offset, err)
	}

	_, pos := indexToEntry(searchRange, uint64(endIdxPos)*IndexEntryBytes)
	if endIdxPos > 0 && pos-startFilePos > maxBytes {
		// binary search will choose the next entry when the left value is less, and the
		// right value is greater and not equal, so fix by grabbing the left
		log.Trace().
			Uint32("start_file_pos", startFilePos).
			Uint32("end_idx_pos", endIdxPos).
			Uint32("pos", pos).
			Uint32("max_bytes", maxBytes).
			Msg("binary search yielded a range too large, trying entry before")
		_, pos = indexToEntry(searchRange, (uint64(endIdxPos)-1)*IndexEntryBytes)
	}

	bytesInRange := pos - startFilePos
	if bytesInRange == 0 || bytesInRange > maxBytes {
		return nil, fmt.Errorf("message exceeded max bytes: %d > %d", bytesInRange, maxBytes)
	}

	log.Trace().
		Uint32("start_file_pos", startFilePos).
		Uint32("pos", pos).
		Msgf("found slice range %d..%d", startFilePos, pos)
	return &MessageSetRange{
		filePosition: startFilePos,
		bytes:        bytesInRange,
	}, nil
}

func (i *Index) Find(offset Offset) (Offset, uint32, error) {
	indexPos, err := i.FindIndexPosition(offset)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to find index position for offset %d: %w", offset, err)
	}
	relOffset, filePos := indexToEntry(i.mmap, indexPos)
	absOffset := uint64(relOffset) + i.baseOffset
	if absOffset < uint64(offset) {
		return 0, 0, ErrIndexNotFound
	} else {
		return Offset(absOffset), filePos, nil
	}
}

func (i *Index) FindIndexPosition(_offset Offset) (uint64, error) {
	offset := uint64(_offset)
	if offset < i.baseOffset {
		return 0, fmt.Errorf("offset %d is less than index base offset %d", offset, i.baseOffset)
	}
	relOffset := uint64(offset) - i.baseOffset

	log.Trace().
		Msgf("offset=%d, next write pos = %d", offset, i.nextWritePos)

	// attempt to find the offset assuming no truncation
	if (i.nextWritePos / IndexEntryBytes) > relOffset {
		log.Trace().
			Uint64("offset", offset).
			Msg("attempting to read offset from exact location")
		// read exact entry
		entryPos := relOffset * IndexEntryBytes
		relOffsetVal := binary.LittleEndian.Uint32(i.mmap[entryPos : entryPos+4])

		log.Trace().
			Msgf("found relative offset. relOffset=%d, entry offset=%d", relOffset, relOffsetVal)

		if relOffsetVal == uint32(relOffset) {
			return entryPos, nil
		}
	}

	log.Trace().
		Uint64("offset", offset).
		Msg("exact offset not found, searching for index position")

	// fall back to binary search otherwise
	targetOffset := relOffset
	compareF := func(_relOffset uint32, _ uint32) Ordering {
		if _relOffset < uint32(targetOffset) {
			return OrderingLess
		}
		if _relOffset > uint32(targetOffset) {
			return OrderingGreater
		}
		return OrderingEqual
	}

	foundIdx, err := binarySearchIndex(i.mmap[:i.nextWritePos], compareF)
	if err != nil {
		return 0, fmt.Errorf("failed to find index position for offset %d: %w", offset, err)
	}

	// check if the found index is within the bounds
	if uint64(foundIdx) < i.nextWritePos/IndexEntryBytes {
		entryPos := uint64(foundIdx) * IndexEntryBytes
		return entryPos, nil
	} else {
		log.Error().
			Uint64("offset", offset).
			Uint64("next_write_pos", i.nextWritePos).
			Msgf("index entry for offset %d not found, next write pos %d", offset, i.nextWritePos)
		return 0, ErrIndexNotFound
	}
}

func (i *Index) Append(offsets *IndexBuf) error {
	if i.baseOffset != offsets.start {
		return fmt.Errorf("index base offset %d does not match offsets start %d", i.baseOffset, offsets.start)
	}

	if i.mode != AccessModeReadWrite {
		return fmt.Errorf("attempt to append to index in read-only mode")
	}

	// check if we need to resize
	if i.Size() < (i.nextWritePos + uint64(len(offsets.buf))) {
		log.Info().
			Uint64("current_size", i.Size()).
			Msg("index size is too small, resizing")
		err := i.resize()
		if err != nil {
			return fmt.Errorf("failed to resize index: %w", err)
		}
	}

	start := i.nextWritePos
	end := start + uint64(len(offsets.buf))

	if end > i.Size() {
		return fmt.Errorf("index write position %d exceeds index size %d", end, i.Size())
	}

	copy(i.mmap[start:end], offsets.buf)

	i.nextWritePos = end

	return nil
}

func (i *Index) resize() error {
	// increase length by 50% -= 7 for alignment
	currSize := i.Size()
	if currSize == 0 {
		return fmt.Errorf("cannot resize index with size 0")
	}
	newSize := currSize + (currSize / 2)
	// align to byte size
	newSize -= (newSize % IndexEntryBytes)

	// unmap the file (set to dummy anonymous map)
	err := i.mmap.Unmap()
	if err != nil {
		return fmt.Errorf("failed to unmap index mmap: %w", err)
	}

	anonMMap, err := mmap.MapRegion(nil, 32, mmap.RDWR, mmap.ANON, 0)
	if err != nil {
		return fmt.Errorf("failed to create anonymous mmap: %w", err)
	}
	i.mmap = anonMMap

	err = i.file.Truncate(int64(newSize))
	if err != nil {
		return fmt.Errorf("failed to truncate index file: %w", err)
	}

	mappedMMap, err := mmap.Map(i.file, mmap.RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to remap index file: %w", err)
	}
	i.mmap = mappedMMap

	return nil
}

func (i *Index) Remove() error {
	err := i.mmap.Unmap()
	if err != nil {
		return fmt.Errorf("failed to unmap index mmap: %w", err)
	}
	err = i.file.Close()
	if err != nil {
		return fmt.Errorf("failed to close index file: %w", err)
	}
	if err := os.Remove(i.path); err != nil {
		return fmt.Errorf("failed to remove index file %s: %w", i.path, err)
	}
	log.Info().
		Str("index_path", i.path).
		Msgf("removed index file %s", i.path)
	return nil
}

func (i *Index) Truncate(offset Offset) (uint32, error) {
	// find the next offset position in order to inform
	// the truncation of the segment
	nextPos, err := i.FindIndexPosition(offset + 1)
	if err != nil {
		return 0, err
	}

	off, fileLen := indexToEntry(i.mmap, nextPos)

	// find_index_pos will find the right-most position, which may include
	// something <= the offset passed in, which we should reject for
	// truncation. This likely occurs when the last offset is the offset
	// requested for truncation OR the offset for truncation is > than the
	// last offset.
	if uint64(off)+i.baseOffset <= uint64(offset) {
		log.Trace().Msg("truncated to exact segment boundary, no need to truncate segment")
		return 0, nil
	}

	log.Info().
		Uint64("offset", uint64(offset)).
		Uint32("file_len", fileLen).
		Msgf("start truncating index at offset %d, file position %d", off, fileLen)

	// override file positions > offset
	mmapRange := i.mmap[nextPos:i.nextWritePos]
	for pos := uint64(0); pos < uint64(len(mmapRange)); pos++ {
		mmapRange[pos] = 0 // set relative offset to 0
	}

	// re-adjust the next write position
	i.nextWritePos = nextPos

	return fileLen, nil
}

func toPageSize(size uint64) uint64 {
	memPageSize := os.Getpagesize()
	truncated := size - (size & uint64((memPageSize - 1)))

	// TODO
	// assert_eq!(truncated % page_size::get(), 0);
	// assert!(truncated <= size);

	return truncated
}

func indexToEntry(mMap mmap.MMap, pos uint64) (uint32, uint32) {
	offset := binary.LittleEndian.Uint32(mMap[pos : pos+4])
	position := binary.LittleEndian.Uint32(mMap[pos+4 : pos+8])
	return offset, position
}

type IndexBuf struct {
	buf   []byte
	start uint64
}

func NewIndexBuf(len uint64, start uint64) *IndexBuf {
	return &IndexBuf{
		buf:   make([]byte, 0, len*IndexEntryBytes),
		start: start,
	}
}

func (ib *IndexBuf) Push(absOffset uint64, position uint32) error {
	if ib.start > absOffset {
		return errors.New("attempt to append to an offset before base offset in index")
	}

	tmpBuf := make([]byte, IndexEntryBytes)
	binary.LittleEndian.PutUint32(tmpBuf[:4], uint32(absOffset)-uint32(ib.start))
	binary.LittleEndian.PutUint32(tmpBuf[4:], position)
	ib.buf = append(ib.buf, tmpBuf...)
	return nil
}

type MessageSetRange struct {
	filePosition uint32
	bytes        uint32
}

func (msr *MessageSetRange) FilePosition() uint32 {
	return msr.filePosition
}

func (msr *MessageSetRange) Bytes() uint32 {
	return msr.bytes
}
