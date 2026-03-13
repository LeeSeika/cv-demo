package commitlog

import (
	"os"
	"path/filepath"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestIndex_BasicAppendAndRead(t *testing.T) {
	path := t.TempDir()

	index, err := NewIndex(path, 9, 1000)
	assert.NoError(t, err, "Failed to create index")

	assert.Equal(t, uint64(1000), index.Size(), "Index size not matching expected value")

	buf := NewIndexBuf(2, 9)
	err = buf.Push(11, 0xffff)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(12, 0xeeee)
	assert.NoError(t, err, "Failed to push to index buffer")

	err = index.Append(buf)
	assert.NoError(t, err, "Failed to append to index")

	err = index.Flush()
	assert.NoError(t, err, "Failed to flush index")

	offset, pos, err := index.ReadEntry(0)
	assert.NoError(t, err, "Failed to read entry from index")
	assert.Equal(t, uint64(11), uint64(offset), "Offset not matching expected value")
	assert.Equal(t, uint32(0xffff), pos, "Position not matching expected value")

	offset, pos, err = index.ReadEntry(1)
	assert.NoError(t, err, "Failed to read entry from index")
	assert.Equal(t, uint64(12), uint64(offset), "Offset not matching expected value")
	assert.Equal(t, uint32(0xeeee), pos, "Position not matching expected value")

	// read an entry that does not exist
	_, _, err = index.ReadEntry(2)
	assert.Error(t, err, "Expected error when reading non-existent entry")
}

func TestIndex_SetReadOnly(t *testing.T) {
	path := t.TempDir()

	index, err := NewIndex(path, 10, 1000)
	assert.NoError(t, err, "Failed to create index")

	buf := NewIndexBuf(2, 10)
	err = buf.Push(11, 0xffff)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(12, 0xeeee)
	assert.NoError(t, err, "Failed to push to index buffer")

	err = index.Append(buf)
	assert.NoError(t, err, "Failed to append to index")

	err = index.Flush()
	assert.NoError(t, err, "Failed to flush index")

	// set read-only mode
	err = index.SetReadonly()
	assert.NoError(t, err, "Failed to set index to read-only mode")
	assert.Equal(t, index.mode, AccessModeReadOnly, "Index mode should be read-only")

	offset, pos, err := index.ReadEntry(1)
	assert.NoError(t, err, "Failed to read entry from index after setting read-only")
	assert.Equal(t, uint64(12), uint64(offset), "Offset not matching expected value after read-only")
	assert.Equal(t, uint32(0xeeee), pos, "Position not matching expected value after read-only")

	// read an entry that does not exist
	_, _, err = index.ReadEntry(2)
	assert.Error(t, err, "Expected error when reading non-existent entry after read-only")

	// try to append to read-only index
	err = index.Append(buf)
	assert.Error(t, err, "Expected error when appending to read-only index")
}

func TestIndex_Open(t *testing.T) {
	path := t.TempDir()

	// issue some writes
	{
		index, err := NewIndex(path, 10, 1000)
		assert.NoError(t, err, "Failed to create index")

		{
			buf := NewIndexBuf(3, 10)
			err = buf.Push(10, 0)
			assert.NoError(t, err, "Failed to push to index buffer")
			err = buf.Push(11, 10)
			assert.NoError(t, err, "Failed to push to index buffer")
			err = buf.Push(12, 20)
			assert.NoError(t, err, "Failed to push to index buffer")

			err = index.Append(buf)
			assert.NoError(t, err, "Failed to append to index")
		}

		{
			buf := NewIndexBuf(2, 10)
			err = buf.Push(13, 30)
			assert.NoError(t, err, "Failed to push to index buffer")
			err = buf.Push(14, 40)
			assert.NoError(t, err, "Failed to push to index buffer")

			err = index.Append(buf)
			assert.NoError(t, err, "Failed to append to index")
		}

		err = index.Flush()
		assert.NoError(t, err, "Failed to flush index")

		err = index.SetReadonly()
		assert.NoError(t, err, "Failed to set index to read-only mode")
	}

	// now open it
	{
		indexPath := filepath.Join(path, "00000000000000000010.index")

		file, err := os.Open(indexPath)
		assert.NoError(t, err, "Failed to open index file")
		defer file.Close()

		stat, err := file.Stat()
		assert.NoError(t, err, "Failed to get file stats")
		assert.False(t, stat.IsDir(), "Index file should not be a directory")

		index, err := OpenIndex(indexPath)
		if err != nil {
			t.Fatalf("Failed to open index: %v", err)
		}

		for i := range 5 {
			offset, pos, err := index.ReadEntry(uint64(i))

			assert.NoError(t, err, "Failed to read entry from index")
			assert.Equal(t, uint64(10+i), uint64(offset), "Offset not matching expected value")
			assert.Equal(t, uint32(i*10), pos, "Position not matching expected value")
		}
	}
}

func TestIndex_OpenWithOneMessage(t *testing.T) {
	path := t.TempDir()

	// issue some writes
	{
		index, err := NewIndex(path, 0, 1000)
		assert.NoError(t, err, "Failed to create index")

		{
			buf := NewIndexBuf(1, 0)
			err = buf.Push(0, 2)
			assert.NoError(t, err, "Failed to push to index buffer")

			err = index.Append(buf)
			assert.NoError(t, err, "Failed to append to index")
		}

		err = index.Flush()
		assert.NoError(t, err, "Failed to flush index")
	}

	// now open it
	{
		indexPath := filepath.Join(path, "00000000000000000000.index")

		file, err := os.Open(indexPath)
		assert.NoError(t, err, "Failed to open index file")
		defer file.Close()

		stat, err := file.Stat()
		assert.NoError(t, err, "Failed to get file stats")
		assert.False(t, stat.IsDir(), "Index file should not be a directory")

		index, err := OpenIndex(indexPath)
		assert.NoError(t, err, "Failed to open index")

		// issue a new write, to make sure we're not overwriting things
		{
			buf := NewIndexBuf(1, 0)
			err = buf.Push(1, 3)
			assert.NoError(t, err, "Failed to push to index buffer")

			err = index.Append(buf)
			assert.NoError(t, err, "Failed to append to index")
		}

		assert.Equal(t, uint64(16), index.nextWritePos, "Next write position does not match expected value")

		offset, pos, err := index.ReadEntry(0)
		assert.NoError(t, err, "Failed to read entry from index")

		assert.Equal(t, uint64(0), uint64(offset), "Offset not matching expected value")
		assert.Equal(t, uint32(2), pos, "Position not matching expected value")
	}
}

func TestIndex_OpenWithOneMessageClosed(t *testing.T) {
	path := t.TempDir()

	// issue some writes
	{
		index, err := NewIndex(path, 0, 1000)
		assert.NoError(t, err, "Failed to create index")

		{
			buf := NewIndexBuf(1, 0)
			err = buf.Push(0, 2)
			assert.NoError(t, err, "Failed to push to index buffer")

			err = index.Append(buf)
			assert.NoError(t, err, "Failed to append to index")
		}

		err = index.Flush()
		assert.NoError(t, err, "Failed to flush index")

		err = index.SetReadonly()
		assert.NoError(t, err, "Failed to set index to read-only mode")
	}

	// now open it
	{
		indexPath := filepath.Join(path, "00000000000000000000.index")

		file, err := os.Open(indexPath)
		assert.NoError(t, err, "Failed to open index file")
		defer file.Close()

		stat, err := file.Stat()
		assert.NoError(t, err, "Failed to get file stats")
		assert.False(t, stat.IsDir(), "Index file should not be a directory")

		index, err := OpenIndex(indexPath)
		assert.NoError(t, err, "Failed to open index")
		assert.Equal(t, AccessModeReadOnly, index.mode, "Index mode should be read-only")

		offset, pos, err := index.ReadEntry(0)
		assert.NoError(t, err, "Failed to read entry from index")

		assert.Equal(t, uint64(0), uint64(offset), "Offset not matching expected value")
		assert.Equal(t, uint32(2), pos, "Position not matching expected value")
	}
}

func TestIndex_Find(t *testing.T) {
	path := t.TempDir()

	index, err := NewIndex(path, 10, 1000)
	assert.NoError(t, err, "Failed to create index")

	buf := NewIndexBuf(8, 10)
	err = buf.Push(10, 1)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(11, 2)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(12, 3)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(15, 4)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(16, 5)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(17, 6)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(18, 7)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(20, 8)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = index.Append(buf)
	assert.NoError(t, err, "Failed to append to index")

	// find exact
	offset, pos, err := index.Find(16)
	assert.NoError(t, err, "Failed to find entry in index")
	assert.Equal(t, uint64(16), uint64(offset), "Offset not matching expected value")
	assert.Equal(t, uint32(5), pos, "Position not matching expected value")

	// find approximate
	offset, pos, err = index.Find(14)
	assert.NoError(t, err, "Failed to find approximate entry in index")
	assert.Equal(t, uint64(15), uint64(offset), "Offset not matching expected value for approximate find")
	assert.Equal(t, uint32(4), pos, "Position not matching expected value for approximate find")
}

func TestIndex_FindOutOfBounds(t *testing.T) {
	path := t.TempDir()

	index, err := NewIndex(path, 10, 1000)
	assert.NoError(t, err, "Failed to create index")

	buf := NewIndexBuf(8, 10)
	err = buf.Push(10, 1)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(11, 2)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(12, 3)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(15, 4)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(16, 5)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(17, 6)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(18, 7)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(20, 8)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = index.Append(buf)
	assert.NoError(t, err, "Failed to append to index")

	// find out of bounds
	_, _, err = index.Find(21)
	assert.Error(t, err, "Expected error when finding out of bounds entry")

	_, _, err = index.Find(7)
	assert.Error(t, err, "Expected error when finding out of bounds entry")
}

func TestIndex_ReopenPartialIndex(t *testing.T) {
	path := t.TempDir()

	{
		index, err := NewIndex(path, 10, 1000)
		assert.NoError(t, err, "Failed to create index")
		buf := NewIndexBuf(8, 10)
		err = buf.Push(10, 1)
		assert.NoError(t, err, "Failed to push to index buffer")
		err = buf.Push(11, 2)
		assert.NoError(t, err, "Failed to push to index buffer")
		err = index.Append(buf)
		assert.NoError(t, err, "Failed to append to index")
		err = index.Flush()
		assert.NoError(t, err, "Failed to flush index")
	}

	{
		indexPath := filepath.Join(path, "00000000000000000010.index")

		file, err := os.Open(indexPath)
		assert.NoError(t, err, "Failed to open index file")
		defer file.Close()

		stat, err := file.Stat()
		assert.NoError(t, err, "Failed to get file stats")
		assert.False(t, stat.IsDir(), "Index file should not be a directory")

		index, err := OpenIndex(indexPath)
		assert.NoError(t, err, "Failed to open index")

		offset, pos, err := index.Find(10)
		assert.NoError(t, err, "Failed to find entry in index")
		assert.Equal(t, uint64(10), uint64(offset), "Offset not matching expected value")
		assert.Equal(t, uint32(1), pos, "Position not matching expected value")

		offset, pos, err = index.Find(11)
		assert.NoError(t, err, "Failed to find entry in index")
		assert.Equal(t, uint64(11), uint64(offset), "Offset not matching expected value")
		assert.Equal(t, uint32(2), pos, "Position not matching expected value")

		// try to find an entry that does not exist
		_, _, err = index.Find(12)
		assert.Error(t, err, "Expected error when finding non-existent entry")

		assert.Equal(t, Offset(12), index.NextOffset())
		assert.Equal(t, AccessModeReadWrite, index.mode, "Index mode should be ReadWrite")
	}
}

func TestIndex_ReopenFullIndex(t *testing.T) {
	path := t.TempDir()

	{
		index, err := NewIndex(path, 10, 16)
		assert.NoError(t, err, "Failed to create index")
		buf := NewIndexBuf(2, 10)
		err = buf.Push(10, 1)
		assert.NoError(t, err, "Failed to push to index buffer")
		err = buf.Push(11, 2)
		assert.NoError(t, err, "Failed to push to index buffer")
		err = index.Append(buf)
		assert.NoError(t, err, "Failed to append to index")

		err = index.Flush()
		assert.NoError(t, err, "Failed to flush index")
	}

	{
		indexPath := filepath.Join(path, "00000000000000000010.index")

		index, err := OpenIndex(indexPath)
		assert.NoError(t, err, "Failed to open index")

		offset, pos, err := index.Find(10)
		assert.NoError(t, err, "Failed to find entry in index")
		assert.Equal(t, uint64(10), uint64(offset), "Offset not matching expected value")
		assert.Equal(t, uint32(1), pos, "Position not matching expected value")

		offset, pos, err = index.Find(11)
		assert.NoError(t, err, "Failed to find entry in index")
		assert.Equal(t, uint64(11), uint64(offset), "Offset not matching expected value")
		assert.Equal(t, uint32(2), pos, "Position not matching expected value")

		// try to find an entry that does not exist
		_, _, err = index.Find(12)
		assert.Error(t, err, "Expected error when finding non-existent entry")
		assert.Equal(t, Offset(12), index.NextOffset(), "Next offset should be 12")
		assert.Equal(t, AccessModeReadOnly, index.mode, "Index mode should be read-only")
	}
}

func TestIndex_FindSegmentRangeOffset(t *testing.T) {
	path := t.TempDir()

	index, err := NewIndex(path, 10, 40)
	assert.NoError(t, err, "Failed to create index")

	buf := NewIndexBuf(5, 10)
	err = buf.Push(10, 10)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(11, 20)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(12, 30)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(13, 40)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(14, 50)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = index.Append(buf)
	assert.NoError(t, err, "Failed to append to index")

	// test offset not in index
	msgSetRange, err := index.FindSegmentRange(9, 50, 60)
	assert.Error(t, err, "Expected error when finding segment range with offset not in index")
	assert.Nil(t, msgSetRange, "Expected nil message set range when offset not in index")

	// test message exceeds max bytes
	msgSetRange, err = index.FindSegmentRange(10, 5, 60)
	assert.Error(t, err, "Expected error when finding segment range with message exceeding max bytes")
	assert.Nil(t, msgSetRange, "Expected nil message set range when message exceeds max bytes")

	// test message within range, not including last message
	msgSetRange, err = index.FindSegmentRange(10, 20, 60)
	assert.NoError(t, err, "Failed to find segment range for offset 10")
	assert.NotNil(t, msgSetRange, "Expected non-nil message set range for offset 10")
	assert.Equal(t, uint32(10), msgSetRange.FilePosition(), "Start offset of message set range should be 10")
	assert.Equal(t, uint32(20), msgSetRange.Bytes(), "Bytes of message set range should be 20")

	// test message within range, not including last message, not first
	msgSetRange, err = index.FindSegmentRange(11, 20, 60)
	assert.NoError(t, err, "Failed to find segment range for offset 11")
	assert.NotNil(t, msgSetRange, "Expected non-nil message set range for offset 11")
	assert.Equal(t, uint32(20), msgSetRange.FilePosition(), "Start offset of message set range should be 20")
	assert.Equal(t, uint32(20), msgSetRange.Bytes(), "Bytes of message set range should be 20")

	// test message within rest of range, including last message
	msgSetRange, err = index.FindSegmentRange(11, 80, 60)
	assert.NoError(t, err, "Failed to find segment range for offset 11 with larger max bytes")
	assert.NotNil(t, msgSetRange, "Expected non-nil message set range for offset 11 with larger max bytes")
	assert.Equal(t, uint32(20), msgSetRange.FilePosition(), "Start offset of message set range should be 20")
	assert.Equal(t, uint32(40), msgSetRange.Bytes(), "Bytes of message set range should be 30")
}

func TestIndex_Resize(t *testing.T) {
	path := t.TempDir()

	index, err := NewIndex(path, 10, 32)
	assert.NoError(t, err, "Failed to create index")

	buf := NewIndexBuf(4, 10)
	err = buf.Push(10, 10)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(11, 20)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(12, 30)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(13, 40)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = index.Append(buf)
	assert.NoError(t, err, "Failed to append to index")

	buf = NewIndexBuf(1, 10)
	err = buf.Push(14, 50)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = index.Append(buf)
	assert.NoError(t, err, "Failed to append to index")

	// make sure the index is resized
	assert.Equal(t, uint64(48), index.Size(), "Index should be resized to 48 bytes")
	// check the entries
	offset, pos, err := index.Find(14)
	assert.NoError(t, err, "Failed to find entry in index after resize")
	assert.Equal(t, uint64(14), uint64(offset), "Offset not matching expected value after resize")
	assert.Equal(t, uint32(50), pos, "Position not matching expected value after resize")
}

func TestIndex_Remove(t *testing.T) {
	path := t.TempDir()

	index, err := NewIndex(path, 0, 1000)
	assert.NoError(t, err, "Failed to create index")

	found := false

	err = filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == "00000000000000000000.index" {
			found = true
		}
		return nil
	})
	assert.NoError(t, err, "Failed to walk directory")
	assert.True(t, found, "Index file should exist before removal")

	err = index.Remove()
	assert.NoError(t, err, "Failed to remove index")

	found = false
	err = filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == "00000000000000000000.index" {
			found = true
		}
		return nil
	})
	assert.NoError(t, err, "Failed to walk directory after removal")
	assert.False(t, found, "Index file should not exist after removal")
}

func TestIndex_Truncate(t *testing.T) {
	path := t.TempDir()

	index, err := NewIndex(path, 10, 128)
	assert.NoError(t, err, "Failed to create index")
	buf := NewIndexBuf(10, 10)
	err = buf.Push(10, 10)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(11, 20)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(12, 30)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(13, 40)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(14, 50)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = index.Append(buf)
	assert.NoError(t, err, "Failed to append to index")

	fileLen, err := index.Truncate(12)
	assert.NoError(t, err, "Failed to truncate index")

	assert.Equal(t, uint32(40), fileLen, "Truncated file length does not match expected value")
	assert.Equal(t, Offset(13), index.NextOffset(), "Next offset after truncate does not match expected value")
	assert.Equal(t, 3*IndexEntryBytes, index.nextWritePos, "Next write position after truncate does not match expected value")

	// ensure we've zeroed the entries
	for i := 3 * IndexEntryBytes; i < 5*IndexEntryBytes; i++ {
		assert.Equal(t, byte(0), index.mmap[i], "Index entry should be zeroed after truncate")
	}
}

func TestIndex_TruncateAtBoundary(t *testing.T) {
	path := t.TempDir()

	index, err := NewIndex(path, 10, 128)
	assert.NoError(t, err, "Failed to create index")
	buf := NewIndexBuf(5, 10)
	err = buf.Push(10, 10)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(11, 20)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(12, 30)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(13, 40)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = buf.Push(14, 50)
	assert.NoError(t, err, "Failed to push to index buffer")
	err = index.Append(buf)
	assert.NoError(t, err, "Failed to append to index")

	fileLen, err := index.Truncate(14)
	assert.Error(t, ErrIndexNotFound, err, "Expected ErrIndexNotFound error when truncating to an offset that is entry boundary")
	assert.Equal(t, uint32(0), fileLen, "Truncated file length should be 0 when truncating to an entry boundary")

	assert.Equal(t, Offset(15), index.NextOffset(), "Next offset after truncate should be 15")
	assert.Equal(t, 5*IndexEntryBytes, index.nextWritePos, "Next write position after truncate should be 5*IndexEntryBytes")
}
