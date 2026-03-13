package commitlog

import (
	"strconv"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCommitLog_Append(t *testing.T) {
	path := t.TempDir()

	log, err := NewCommitLog(DefaultLogOptions(path))
	assert.NoError(t, err, "Failed to create new commit log")

	startPos, _, err := log.AppendMsg([]byte("test message 1"))
	assert.NoError(t, err, "Failed to append message to commit log")
	assert.Equal(t, Offset(0), startPos, "Expected start position to be 0")
	startPos, _, err = log.AppendMsg([]byte("test message 2"))
	assert.NoError(t, err, "Failed to append message to commit log")
	assert.Equal(t, Offset(1), startPos, "Expected start position to be 1")
	startPos, _, err = log.AppendMsg([]byte("test message 3"))
	assert.NoError(t, err, "Failed to append message to commit log")
	assert.Equal(t, Offset(2), startPos, "Expected start position to be 2")

	err = log.Flush()
	assert.NoError(t, err, "Failed to flush commit log")
}

func TestCommitLog_AppendMultiple(t *testing.T) {
	path := t.TempDir()

	log, err := NewCommitLog(DefaultLogOptions(path))
	assert.NoError(t, err, "Failed to create new commit log")

	buf := DefaultMessageBuf()
	err = buf.Push([]byte("message 1"))
	assert.NoError(t, err, "Failed to push data into message buffer")
	err = buf.Push([]byte("message 2"))
	assert.NoError(t, err, "Failed to push data into message buffer")
	err = buf.Push([]byte("message 3"))
	assert.NoError(t, err, "Failed to push data into message buffer")

	offRange, _, err := log.Append(buf)
	assert.NoError(t, err, "Failed to append messages to commit log")
	assert.Equal(t, OffsetRange{start: 0, len: 3}, offRange, "Expected offset range to be {0, 3}")
}

func TestCommitLog_NewSegment(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	opts.logMaxBytes = 62

	{
		log, err := NewCommitLog(opts)
		assert.NoError(t, err, "Failed to create new commit log")
		// first 2 entries fit (both 30 bytes with encoding)
		_, _, err = log.AppendMsg([]byte("0123456789"))
		assert.NoError(t, err, "Failed to append first message to commit log")
		_, _, err = log.AppendMsg([]byte("0123456789"))
		assert.NoError(t, err, "Failed to append second message to commit log")

		// this one should roll the log
		_, _, err = log.AppendMsg([]byte("0123456789"))
		assert.NoError(t, err, "Failed to append third message to commit log")

		err = log.Flush()
		assert.NoError(t, err, "Failed to flush commit log")
	}

	match, err := expectFiles(
		path,
		[]string{
			"00000000000000000000.log",
			"00000000000000000000.index",
			"00000000000000000002.log",
			"00000000000000000002.index",
		},
	)
	assert.NoError(t, err, "Failed to check expected files in commit log directory")
	assert.True(t, match, "Expected files in commit log directory do not match actual files")
}

func TestCommitLog_ReadEntries(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(1000)

	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	for i := range 100 {
		msg := []byte("-data " + strconv.Itoa(i))
		_, _, err := log.AppendMsg(msg)
		assert.NoError(t, err, "Failed to append message to commit log")
	}

	err = log.Flush()
	assert.NoError(t, err, "Failed to flush commit log")

	{
		activeIndexRead, err := log.Read(82, NewReadLimitWithMaxBytes(168))
		assert.NoError(t, err, "Failed to read active index")
		assert.Equal(t, uint64(6), activeIndexRead.Len(), "Expected to read 6 messages from active index")

		expectedIndex := []uint64{82, 83, 84, 85, 86, 87}
		for i, msg := range activeIndexRead.Messages() {
			assert.Equal(t, msg.Offset(), expectedIndex[i], "Expected message offset to match")
		}
	}

	{
		oldIndexRead, err := log.Read(5, NewReadLimitWithMaxBytes(112))
		assert.NoError(t, err, "Failed to read old index")
		assert.Equal(t, uint64(4), oldIndexRead.Len(), "Expected to read 4 messages from old index")

		expectedIndex := []uint64{5, 6, 7, 8}
		for i, msg := range oldIndexRead.Messages() {
			assert.Equal(t, msg.Offset(), expectedIndex[i], "Expected message offset to match")
		}
	}

	// read at the boundary (not going to get full message limit)
	{
		// log rolls at offset 36
		boundaryRead, err := log.Read(33, NewReadLimitWithMaxBytes(100))
		assert.NoError(t, err, "Failed to read at boundary")
		assert.Equal(t, uint64(3), boundaryRead.Len(), "Expected to read 3 messages at the boundary")

		expectedIndex := []uint64{33, 34, 35}
		for i, msg := range boundaryRead.Messages() {
			assert.Equal(t, msg.Offset(), expectedIndex[i], "Expected message offset to match at boundary")
		}
	}
}

func TestCommitLog_Reopen(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(1000)

	{
		log, err := NewCommitLog(opts)
		assert.NoError(t, err, "Failed to create new commit log")

		for i := range 99 {
			msg := []byte("some data " + strconv.Itoa(i))
			_, _, err := log.AppendMsg(msg)
			assert.NoError(t, err, "Failed to append message to commit log")
		}

		err = log.Flush()
		assert.NoError(t, err, "Failed to flush commit log")
	}

	{
		log, err := NewCommitLog(opts)
		assert.NoError(t, err, "Failed to reopen commit log")

		activeIndexRead, err := log.Read(82, NewReadLimitWithMaxBytes(130))
		assert.NoError(t, err, "Failed to read active index after reopening")
		assert.Equal(t, uint64(4), activeIndexRead.Len(), "Expected to read 4 messages from active index after reopening")

		expectedIndex := []uint64{82, 83, 84, 85}
		for i, msg := range activeIndexRead.Messages() {
			assert.Equal(t, msg.Offset(), expectedIndex[i], "Expected message offset to match after reopening")
		}

		offset, _, err := log.AppendMsg([]byte("new message after reopen"))
		assert.NoError(t, err, "Failed to append message after reopening commit log")
		assert.Equal(t, Offset(99), offset, "Expected appended message offset to be 99")
	}
}

func TestCommitLog_ReopenWithoutSegmentWrite(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(1000)

	{
		log, err := NewCommitLog(opts)
		assert.NoError(t, err, "Failed to create new commit log")

		err = log.Flush()
		assert.NoError(t, err, "Failed to flush commit log without writing segments")
	}

	{
		_, err := NewCommitLog(opts)
		assert.NoError(t, err, "Failed to reopen commit log without writing segments")
	}
}

func TestCommitLog_ReopenWithOneSegmentWrite(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)

	{
		log, err := NewCommitLog(opts)
		assert.NoError(t, err, "Failed to create new commit log")

		// write one message to create a segment
		_, _, err = log.AppendMsg([]byte("test message"))
		assert.NoError(t, err, "Failed to append message to commit log")
		err = log.Flush()
		assert.NoError(t, err, "Failed to flush commit log after writing one segment")
	}

	{
		log, err := NewCommitLog(opts)
		assert.NoError(t, err, "Failed to reopen commit log after writing one segment")

		assert.Equal(t, Offset(1), log.NextOffset(), "Expected next offset to be 1 after reopening")
	}
}

func TestCommitLog_AppendMessageGreaterThanMax(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	value := make([]byte, 1200000)
	target := 0
	for target != 1200000 {
		value[target] = 0x61 // fill with some data
		target++
	}

	_, _, err = log.AppendMsg(value)
	assert.Error(t, err, "Expected error when appending message greater than max bytes")

	err = log.Flush()
	assert.NoError(t, err, "Failed to flush commit log after attempting to append large message")
}

func TestCommitLog_TruncateFromActive(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	// append 5 messages
	{
		buf := DefaultMessageBuf()
		err = buf.Push([]byte("message 1"))
		assert.NoError(t, err, "Failed to push message 1 into buffer")
		err = buf.Push([]byte("message 2"))
		assert.NoError(t, err, "Failed to push message 2 into buffer")
		err = buf.Push([]byte("message 3"))
		assert.NoError(t, err, "Failed to push message 3 into buffer")
		err = buf.Push([]byte("message 4"))
		assert.NoError(t, err, "Failed to push message 4 into buffer")
		err = buf.Push([]byte("message 5"))
		assert.NoError(t, err, "Failed to push message 5 into buffer")

		_, _, err = log.Append(buf)
		assert.NoError(t, err, "Failed to append messages to commit log")
	}

	// truncate to offset 2 (should remove 2 messages)
	err = log.Truncate(Offset(2))
	assert.NoError(t, err, "Failed to truncate commit log from active segment")

	assert.Equal(t, Offset(2), log.LastOffset(), "Expected last offset after truncation to be 2")
}

func TestCommitLog_TruncateAfterOffsetRemovesSegments(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(52)

	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	// append 7 messages (4 segments)
	{
		for range 7 {
			msg := []byte("12345")
			_, _, err := log.AppendMsg(msg)
			assert.NoError(t, err, "Failed to append message to commit log")
		}
	}

	// ensure we have the expected index/segment files
	match, err := expectFiles(
		path,
		[]string{
			"00000000000000000000.index",
			"00000000000000000000.log",
			"00000000000000000002.log",
			"00000000000000000002.index",
			"00000000000000000004.log",
			"00000000000000000004.index",
			"00000000000000000006.log",
			"00000000000000000006.index",
		},
	)

	assert.NoError(t, err, "Failed to check expected files in commit log directory")
	assert.True(t, match, "Expected files in commit log directory do not match actual files")

	// truncate after offset 3
	err = log.Truncate(Offset(3))
	assert.NoError(t, err, "Failed to truncate commit log after offset 3")
	assert.Equal(t, Offset(3), log.LastOffset(), "Expected last offset after truncation to be 3")

	// ensure we have the expected index/segment files after truncation
	match, err = expectFiles(
		path,
		[]string{
			"00000000000000000000.index",
			"00000000000000000000.log",
			"00000000000000000002.log",
			"00000000000000000002.index",
		},
	)
	assert.NoError(t, err, "Failed to check expected files in commit log directory after truncation")
	assert.True(t, match, "Expected files in commit log directory after truncation do not match actual files")
}

func TestCommitLog_TruncateAtSegmentBoundaryRemovesSegment(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(52)

	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	// append 7 messages (4 segments)
	{
		for range 7 {
			msg := []byte("12345")
			_, _, err := log.AppendMsg(msg)
			assert.NoError(t, err, "Failed to append message to commit log")
		}
	}

	// ensure we have the expected index/segment files
	match, err := expectFiles(
		path,
		[]string{
			"00000000000000000000.index",
			"00000000000000000000.log",
			"00000000000000000002.log",
			"00000000000000000002.index",
			"00000000000000000004.log",
			"00000000000000000004.index",
			"00000000000000000006.log",
			"00000000000000000006.index",
		},
	)

	assert.NoError(t, err, "Failed to check expected files in commit log directory")
	assert.True(t, match, "Expected files in commit log directory do not match actual files")

	// truncate to offset 2 (should remove 2 messages)
	err = log.Truncate(Offset(2))
	assert.NoError(t, err, "Failed to truncate commit log at segment boundary")
	assert.Equal(t, Offset(2), log.LastOffset(), "Expected last offset after truncation to be 2")

	// ensure we have the expected index/segment files after truncation
	match, err = expectFiles(
		path,
		[]string{
			"00000000000000000000.index",
			"00000000000000000000.log",
			"00000000000000000002.log",
			"00000000000000000002.index",
		},
	)
	assert.NoError(t, err, "Failed to check expected files in commit log directory after truncation")
	assert.True(t, match, "Expected files in commit log directory after truncation do not match actual files")
}

func TestCommitLog_TruncateAfterLastAppendDoesNothing(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(52)

	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	// append 7 messages (4 segments)
	{
		for range 7 {
			msg := []byte("12345")
			_, _, err := log.AppendMsg(msg)
			assert.NoError(t, err, "Failed to append message to commit log")
		}
	}

	// ensure we have the expected index/segment files
	match, err := expectFiles(
		path,
		[]string{
			"00000000000000000000.index",
			"00000000000000000000.log",
			"00000000000000000002.log",
			"00000000000000000002.index",
			"00000000000000000004.log",
			"00000000000000000004.index",
			"00000000000000000006.log",
			"00000000000000000006.index",
		},
	)
	assert.NoError(t, err, "Failed to check expected files in commit log directory")
	assert.True(t, match, "Expected files in commit log directory do not match actual files")

	err = log.Truncate(Offset(7))
	assert.NoError(t, err, "Failed to truncate commit log after last append")

	// ensure we still have the same index/segment files after truncation
	match, err = expectFiles(
		path,
		[]string{
			"00000000000000000000.index",
			"00000000000000000000.log",
			"00000000000000000002.log",
			"00000000000000000002.index",
			"00000000000000000004.log",
			"00000000000000000004.index",
			"00000000000000000006.log",
			"00000000000000000006.index",
		},
	)
	assert.NoError(t, err, "Failed to check expected files in commit log directory after truncation")
	assert.True(t, match, "Expected files in commit log directory after truncation do not match actual files")
}

func TestCommitLog_TrimSegmentsBeforeRemovesSegments(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(52)

	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	// append 7 messages (4 segments)
	{
		for range 7 {
			msg := []byte("12345")
			_, _, err := log.AppendMsg(msg)
			assert.NoError(t, err, "Failed to append message to commit log")
		}
	}

	// ensure we have the expected index/segment files
	match, err := expectFiles(
		path,
		[]string{
			"00000000000000000000.index",
			"00000000000000000000.log",
			"00000000000000000002.log",
			"00000000000000000002.index",
			"00000000000000000004.log",
			"00000000000000000004.index",
			"00000000000000000006.log",
			"00000000000000000006.index",
		},
	)

	assert.NoError(t, err, "Failed to check expected files in commit log directory")
	assert.True(t, match, "Expected files in commit log directory do not match actual files")

	err = log.TrimSegmentsBefore(Offset(3))
	assert.NoError(t, err, "Failed to trim segments before offset 3")
	assert.Equal(t, Offset(6), log.LastOffset(), "Expected last offset after trimming to be 6")
}

func TestCommitLog_TrimSegmentsBeforeRemovesSegmentAtBoundary(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(52)

	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	// append 7 messages (4 segments)
	{
		for range 7 {
			msg := []byte("12345")
			_, _, err := log.AppendMsg(msg)
			assert.NoError(t, err, "Failed to append message to commit log")
		}
	}

	// ensure we have the expected index/segment files
	match, err := expectFiles(
		path,
		[]string{
			"00000000000000000000.index",
			"00000000000000000000.log",
			"00000000000000000002.log",
			"00000000000000000002.index",
			"00000000000000000004.log",
			"00000000000000000004.index",
			"00000000000000000006.log",
			"00000000000000000006.index",
		},
	)

	assert.NoError(t, err, "Failed to check expected files in commit log directory")
	assert.True(t, match, "Expected files in commit log directory do not match actual files")

	err = log.TrimSegmentsBefore(Offset(4))
	assert.NoError(t, err, "Failed to trim segments before offset 4")
	assert.Equal(t, Offset(6), log.LastOffset(), "Expected last offset after trimming to be 6")

	// ensure we have the expected index/segment files after trimming
	match, err = expectFiles(
		path,
		[]string{
			"00000000000000000004.index",
			"00000000000000000004.log",
			"00000000000000000006.log",
			"00000000000000000006.index",
		},
	)
	assert.NoError(t, err, "Failed to check expected files in commit log directory after trimming")
	assert.True(t, match, "Expected files in commit log directory after trimming do not match actual files")

	// make sure the messages are really gone
	msgBuf, err := log.Read(0, DefaultReadLimit())
	assert.NoError(t, err, "Failed to read messages after trimming segments")
	msgs := msgBuf.Messages()
	assert.Len(t, msgs, 2, "Expected only 2 messages after trimming segments")
	assert.Equal(t, uint64(4), msgs[0].Offset(), "Expected first message offset to be 4 after trimming")
}

func TestCommitLog_TrimStartLogicCheck(t *testing.T) {
	totalMessages := uint64(20)
	testedTrimStart := totalMessages + 1

	for trimOff := range testedTrimStart {
		path := t.TempDir()

		opts := DefaultLogOptions(path)
		opts.SetIndexMaxItems(20)
		opts.SetSegmentMaxBytes(52)

		log, err := NewCommitLog(opts)
		assert.NoError(t, err, "Failed to create new commit log")

		// append messages
		for i := uint64(0); i < totalMessages; i++ {
			msg := []byte("12345")
			_, _, err := log.AppendMsg(msg)
			assert.NoError(t, err, "Failed to append message to commit log")
		}

		err = log.TrimSegmentsBefore(Offset(trimOff))
		assert.NoError(t, err, "Failed to trim segments before offset")

		// make sure the messages are really gone
		msgBuf, err := log.Read(0, DefaultReadLimit())
		assert.NoError(t, err, "Failed to read messages after trimming segments")

		msgs := msgBuf.Messages()
		assert.True(t, len(msgs) > 0, "Expected some messages after trimming segments")
		startOffset := msgs[0].Offset()
		assert.LessOrEqual(t, startOffset, trimOff, "Expected start offset after trimming to be less than or equal to trimOff")
	}
}

func TestCommitLog_MultipleTrimStartCalls(t *testing.T) {
	totalMessages := uint64(20)

	path := t.TempDir()
	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(52)

	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	// append messages
	for i := uint64(0); i < totalMessages; i++ {
		msg := []byte("12345")
		_, _, err := log.AppendMsg(msg)
		assert.NoError(t, err, "Failed to append message to commit log")
	}

	err = log.TrimSegmentsBefore(Offset(2))
	assert.NoError(t, err, "Failed to trim segments before offset 2")

	{
		msgBuf, err := log.Read(0, DefaultReadLimit())
		assert.NoError(t, err, "Failed to read messages after first trim")

		msgs := msgBuf.Messages()
		assert.True(t, len(msgs) > 0, "Expected some messages after first trim")
		startOffset := msgs[0].Offset()
		assert.LessOrEqual(t, startOffset, Offset(2), "Expected start offset after first trim to be less than or equal to 2")
	}

	err = log.TrimSegmentsBefore(Offset(10))
	assert.NoError(t, err, "Failed to trim segments before offset 10")

	{
		msgBuf, err := log.Read(0, DefaultReadLimit())
		assert.NoError(t, err, "Failed to read messages after second trim")
		msgs := msgBuf.Messages()
		assert.True(t, len(msgs) > 0, "Expected some messages after second trim")
		startOffset := msgs[0].Offset()
		assert.LessOrEqual(t, startOffset, Offset(10), "Expected start offset after second trim to be less than or equal to 10")
	}
}

func TestCommitLog_TrimInactiveLogicCheck(t *testing.T) {
	totalMessages := uint64(20)

	path := t.TempDir()
	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(52)

	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	// append messages
	{
		for i := uint64(0); i < totalMessages; i++ {
			msg := []byte("12345")
			_, _, err := log.AppendMsg(msg)
			assert.NoError(t, err, "Failed to append message to commit log")
		}
	}

	err = log.TrimInactiveSegments()
	assert.NoError(t, err, "Failed to trim inactive segments")
	assert.Equal(t, Offset(totalMessages-1), log.LastOffset(), "Expected last offset after trimming inactive segments to be %d", totalMessages-1)

	// make sure the messages are really gone
	msgBuf, err := log.Read(0, DefaultReadLimit())
	assert.NoError(t, err, "Failed to read messages after trimming inactive segments")
	msgs := msgBuf.Messages()
	assert.True(t, len(msgs) > 0, "Expected some messages after trimming inactive segments")
	startOffset := msgs[0].Offset()
	assert.Equal(t, startOffset, uint64(16), "Expected start offset after trimming inactive segments to be 16, got %d", startOffset)
}

func TestCommitLog_TrimInactiveLogicCheckZeroMessages(t *testing.T) {
	path := t.TempDir()
	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(20)
	opts.SetSegmentMaxBytes(52)

	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	err = log.TrimInactiveSegments()
	assert.NoError(t, err, "Failed to trim inactive segments when no messages are present")
	assert.Equal(t, Offset(0), log.LastOffset(), "Expected last offset after trimming inactive segments with no messages to be 0")

	// append the messages
	_, _, err = log.AppendMsg([]byte("12345"))
	assert.NoError(t, err, "Failed to append message to commit log after trimming inactive segments")

	// make sure the messages are really gone
	msgBuf, err := log.Read(0, DefaultReadLimit())
	assert.NoError(t, err, "Failed to read messages after appending after trimming inactive segments")
	msgs := msgBuf.Messages()
	assert.Len(t, msgs, 1, "Expected one message after appending after trimming inactive segments")
	startOffset := msgs[0].Offset()
	assert.Equal(t, startOffset, uint64(0), "Expected start offset after appending after trimming inactive segments to be 0, got %d", startOffset)
}

func TestCommitLog_ReadLogsFromRust(t *testing.T) {
	expected := []string{
		"hello world",
		"second message",
	}

	targetPath := "./test-data/.log_rust"
	log, err := NewCommitLog(DefaultLogOptions(targetPath))
	assert.NoError(t, err, "Failed to create new commit log from Rust data")

	msgBuf, err := log.Read(0, DefaultReadLimit())
	assert.NoError(t, err, "Failed to read messages from Rust log")

	messages := msgBuf.Messages()
	assert.Len(t, messages, len(expected), "Expected number of messages to match")
	for i, msg := range messages {
		assert.Equal(t, expected[i], string(msg.Payload()), "Expected message data to match at index %d", i)
	}
}

func TestCommitLog_ReopenFileSetWithAllFullIndexFiles(t *testing.T) {
	path := t.TempDir()

	opts := DefaultLogOptions(path)
	opts.SetIndexMaxItems(2)    // each index can hold 2 messages
	opts.SetSegmentMaxBytes(52) // each segment can hold 2 messages

	log, err := NewCommitLog(opts)
	assert.NoError(t, err, "Failed to create new commit log")

	// append 6 messages (cause 3 full segments)
	{
		for range 6 {
			msg := []byte("12345")
			_, _, err := log.AppendMsg(msg)
			assert.NoError(t, err, "Failed to append message to commit log")
		}
	}

	// ensure we have the expected index/segment files
	match, err := expectFiles(
		path,
		[]string{
			"00000000000000000000.index",
			"00000000000000000000.log",
			"00000000000000000002.log",
			"00000000000000000002.index",
			"00000000000000000004.log",
			"00000000000000000004.index",
		},
	)

	assert.NoError(t, err, "Failed to check expected files in commit log directory")
	assert.True(t, match, "Expected files in commit log directory do not match actual files")

	// next offset should be 6
	expectedNextOffset := Offset(6)
	assert.Equal(t, expectedNextOffset, log.NextOffset(), "Expected next offset to be %d", expectedNextOffset)

	// flush to ensure all data is written
	err = log.Flush()
	assert.NoError(t, err, "Failed to flush commit log before reopening")

	// now all existing index files are full, reopen the log
	log, err = NewCommitLog(opts)
	assert.NoError(t, err, "Failed to reopen commit log with all full index files")

	// next offset should be 6 (next offset should be sequential after last max index)
	assert.Equal(t, expectedNextOffset, log.NextOffset(), "Expected next offset after reopening to be 6")

	// new created active index is still appendable
	_, _, err = log.AppendMsg([]byte("12345"))
	assert.NoError(t, err, "Failed to append message to commit log to fill last index file")
}
