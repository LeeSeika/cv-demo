package commitlog

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestSegment_AppendLog(t *testing.T) {
	path := t.TempDir()

	seg, err := NewSegment(path, 0, 1024)
	assert.NoError(t, err, "Failed to create new segment")

	{
		buf := DefaultMessageBuf()
		err := buf.Push([]byte("12345"))
		assert.NoError(t, err, "Failed to push data into message buffer")

		startPos, err := seg.Append(buf)
		assert.NoError(t, err, "Failed to append message to segment")
		assert.Equal(t, uint64(2), startPos, "Expected start position to be 2")
	}

	{
		buf := DefaultMessageBuf()
		err := buf.Push([]byte("66666"))
		assert.NoError(t, err, "Failed to push data into message buffer")
		err = buf.Push([]byte("77777"))
		assert.NoError(t, err, "Failed to push data into message buffer")

		startPos, err := seg.Append(buf)
		assert.NoError(t, err, "Failed to append message to segment")
		assert.Equal(t, uint64(27), startPos, "Expected start position to be 9")

		messages := buf.Messages()
		assert.Equal(t, 2, len(messages), "Expected 2 messages in the buffer after appending")

		firstMsg := messages[0]
		assert.Equal(t, uint64(25), firstMsg.TotalBytes(), "Expected first message total bytes to be 25")
		secondMsg := messages[1]
		assert.Equal(t, uint64(25), secondMsg.TotalBytes(), "Expected second message total bytes to be 25")
	}

	err = seg.Flush()
	assert.NoError(t, err, "Failed to flush segment")
}

func TestSegment_OpenLog(t *testing.T) {
	path := t.TempDir()

	{
		seg, err := NewSegment(path, 0, 1024)
		assert.NoError(t, err, "Failed to create new segment")
		buf := DefaultMessageBuf()
		err = buf.Push([]byte("12345"))
		assert.NoError(t, err, "Failed to push data into message buffer")
		err = buf.Push([]byte("66666"))
		assert.NoError(t, err, "Failed to push data into message buffer")
		_, err = seg.Append(buf)
		assert.NoError(t, err, "Failed to append message to segment")

		err = seg.Flush()
		assert.NoError(t, err, "Failed to flush segment")
	}

	// open it
	{
		segPath := filepath.Join(path, "00000000000000000000.log")
		seg, err := OpenSegment(segPath, 1024)
		assert.NoError(t, err, "Failed to open segment")

		assert.Equal(t, uint64(0), seg.StartingOffset(), "Expected segment starting offset to be 0")
	}
}

func TestSegment_ReadLog(t *testing.T) {
	path := t.TempDir()
	seg, err := NewSegment(path, 0, 1024)
	assert.NoError(t, err, "Failed to create new segment")

	{
		buf := DefaultMessageBuf()
		err := buf.Push([]byte("0123456789"))
		assert.NoError(t, err, "Failed to push data into message buffer")
		err = buf.Push([]byte("aaaaaaaaaa"))
		assert.NoError(t, err, "Failed to push data into message buffer")
		err = buf.Push([]byte("abc"))
		assert.NoError(t, err, "Failed to push data into message buffer")

		setMessagesOffsets(buf, 0)
		_, err = seg.Append(buf)
		assert.NoError(t, err, "Failed to append message to segment")
	}

	reader := MessageBufReader{}
	msgBuf, err := seg.ReadSlice(&reader, 2, 83)
	assert.NoError(t, err, "Failed to read messages from segment")
	assert.Equal(t, uint64(3), msgBuf.Len(), "Expected to read 3 messages from segment")

	messages := msgBuf.Messages()
	for i := range len(messages) {
		msg := messages[i]
		assert.Equal(t, uint64(i), msg.Offset(), "Expected message offset to match index")
	}
}

func TestSegment_ReadLogWithSizeLimit(t *testing.T) {
	path := t.TempDir()
	seg, err := NewSegment(path, 0, 1024)
	assert.NoError(t, err, "Failed to create new segment")

	buf := DefaultMessageBuf()
	err = buf.Push([]byte("0123456789"))
	assert.NoError(t, err, "Failed to push data into message buffer")
	err = buf.Push([]byte("aaaaaaaaaa"))
	assert.NoError(t, err, "Failed to push data into message buffer")
	err = buf.Push([]byte("abc"))
	assert.NoError(t, err, "Failed to push data into message buffer")

	setMessagesOffsets(buf, 0)
	startPos, err := seg.Append(buf)
	assert.NoError(t, err, "Failed to append message to segment")

	msgs := buf.Messages()
	assert.Equal(t, 3, len(msgs), "Expected 3 messages in the buffer after appending")

	secondMsgStartPos := startPos + msgs[0].TotalBytes()
	reader := MessageBufReader{}
	readSize := uint32(secondMsgStartPos - 2)
	msgBuf, err := seg.ReadSlice(&reader, uint32(2), readSize)
	assert.NoError(t, err, "Failed to read messages from segment with size limit")

	assert.Equal(t, uint64(1), msgBuf.Len(), "Expected to read 2 messages from segment with size limit")
}

func TestSegment_ReadFromWrite(t *testing.T) {
	path := t.TempDir()
	seg, err := NewSegment(path, 0, 1024)
	assert.NoError(t, err, "Failed to create new segment")

	{
		buf := DefaultMessageBuf()
		err := buf.Push([]byte("0123456789"))
		assert.NoError(t, err, "Failed to push data into message buffer")
		err = buf.Push([]byte("aaaaaaaaaa"))
		assert.NoError(t, err, "Failed to push data into message buffer")
		err = buf.Push([]byte("abc"))
		assert.NoError(t, err, "Failed to push data into message buffer")
		setMessagesOffsets(buf, 0)
		_, err = seg.Append(buf)
		assert.NoError(t, err, "Failed to append message to segment")
	}

	reader := MessageBufReader{}
	msgBuf, err := seg.ReadSlice(&reader, 2, 83)
	assert.NoError(t, err, "Failed to read messages from segment")
	assert.Equal(t, uint64(3), msgBuf.Len(), "Expected to read 3 messages from segment")

	{
		buf := DefaultMessageBuf()
		err := buf.Push([]byte("foo"))
		assert.NoError(t, err, "Failed to push data into message buffer")
		setMessagesOffsets(buf, 3)
		_, err = seg.Append(buf)
		assert.NoError(t, err, "Failed to append message to segment")
	}

	// read again
	msgBuf, err = seg.ReadSlice(&reader, 2, 106)
	assert.NoError(t, err, "Failed to read messages from segment after writing more data")
	assert.Equal(t, uint64(4), msgBuf.Len(), "Expected to read 4 messages from segment after writing more data")

	for i, msg := range msgBuf.Messages() {
		assert.Equal(t, uint64(i), msg.Offset(), "Expected message offset to match index after reading again")
	}
}

func TestSegment_RemoveLog(t *testing.T) {
	path := t.TempDir()
	seg, err := NewSegment(path, 0, 1024)
	assert.NoError(t, err, "Failed to create new segment")

	found := false

	err = filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == "00000000000000000000.log" {
			found = true
		}
		return nil
	})
	assert.NoError(t, err, "Failed to walk directory")
	assert.True(t, found, "Segment file should exist before removal")

	err = seg.Remove()
	assert.NoError(t, err, "Failed to remove segment")

	found = false
	err = filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == "00000000000000000000.log" {
			found = true
		}
		return nil
	})
	assert.NoError(t, err, "Failed to walk directory after removal")
	assert.False(t, found, "Segment file should not exist after removal")
}

func TestSegment_TruncateLog(t *testing.T) {
	path := t.TempDir()
	seg, err := NewSegment(path, 0, 1024)
	assert.NoError(t, err, "Failed to create new segment")

	buf := DefaultMessageBuf()
	err = buf.Push([]byte("0123456789"))
	assert.NoError(t, err, "Failed to push data into message buffer")
	err = buf.Push([]byte("aaaaaaaaaa"))
	assert.NoError(t, err, "Failed to push data into message buffer")
	err = buf.Push([]byte("abc"))
	assert.NoError(t, err, "Failed to push data into message buffer")

	setMessagesOffsets(buf, 0)
	startPos, err := seg.Append(buf)
	assert.NoError(t, err, "Failed to append message to segment")

	reader := MessageBufReader{}
	msgBuf, err := seg.ReadSlice(&reader, 2, uint32(seg.Size()-2))
	assert.NoError(t, err, "Failed to read messages from segment")
	assert.Equal(t, uint64(3), msgBuf.Len(), "Expected to read 3 messages from segment")

	// find the second message starting point position in the segment
	msgs := msgBuf.Messages()
	secondMsgStartPos := startPos + msgs[0].TotalBytes()

	// truncate to first message
	err = seg.Truncate(uint32(secondMsgStartPos))
	assert.NoError(t, err, "Failed to truncate segment")

	assert.Equal(t, uint64(secondMsgStartPos), seg.Size(), "Expected segment size to match truncated position")

	file, err := os.OpenFile(seg.path, os.O_RDONLY, fs.ModePerm)
	assert.NoError(t, err, "Failed to open segment file after truncation")
	defer file.Close()

	stat, err := file.Stat()
	assert.NoError(t, err, "Failed to get segment file stats after truncation")

	assert.Equal(t, int64(secondMsgStartPos), stat.Size(), "Expected segment file size to match truncated position")

	buf = DefaultMessageBuf()
	err = buf.Push([]byte("zzzzzzzzzz"))
	assert.NoError(t, err, "Failed to push data into message buffer after truncation")

	setMessagesOffsets(buf, 1) // start from offset 1 after truncation
	startPos, err = seg.Append(buf)
	assert.NoError(t, err, "Failed to append message to segment after truncation")

	assert.Equal(t, uint64(secondMsgStartPos), startPos, "Expected start position to match truncated position after appending new message")

	stat, err = file.Stat()
	assert.NoError(t, err, "Failed to get segment file stats after appending new message")
	assert.Equal(t, seg.Size(), uint64(stat.Size()), "Expected segment file size to match truncated position after appending new message")

	// read again after truncation and appending new message
	reader = MessageBufReader{}
	msgBuf, err = seg.ReadSlice(&reader, 2, uint32(seg.Size()-2))
	assert.NoError(t, err, "Failed to read messages from segment after truncation and appending new message")
	assert.Equal(t, uint64(2), msgBuf.Len(), "Expected to read 2 messages from segment after truncation and appending new message")
}
