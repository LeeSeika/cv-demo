package commitlog

import (
	"bytes"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestMessage_Construction(t *testing.T) {
	msgBuf := DefaultMessageBuf()
	err := msgBuf.Push([]byte("123456"))
	assert.NoError(t, err, "Failed to push data into message buffer")
	err = msgBuf.Push([]byte("000000000"))
	assert.NoError(t, err, "Failed to push data into message buffer")

	setMessagesOffsets(msgBuf, 100)

	msgs := msgBuf.Messages()
	assert.Equal(t, 2, len(msgs), "Expected 2 messages in the buffer")

	firstMsg := msgs[0]
	assert.Equal(t, uint64(100), firstMsg.Offset(), "Expected first message offset to be 100")
	assert.Equal(t, []byte("123456"), firstMsg.Payload(), "Expected first message payload to be '123456'")
	assert.Equal(t, uint32(6), firstMsg.Size(), "Expected first message size to be 6 bytes")
	assert.True(t, firstMsg.VerifyHash(), "Expected first message to verify its hash")

	secondMsg := msgs[1]
	assert.Equal(t, uint64(101), secondMsg.Offset(), "Expected second message offset to be 101")
	assert.Equal(t, []byte("000000000"), secondMsg.Payload(), "Expected second message payload to be '000000000'")
	assert.Equal(t, uint32(9), secondMsg.Size(), "Expected second message size to be 9 bytes")
	assert.True(t, secondMsg.VerifyHash(), "Expected second message to verify its hash")
}

func TestMessage_Read(t *testing.T) {
	buf, err := serialize(buf{}, 120, []byte(""), []byte("123456789"))
	assert.NoError(t, err, "Failed to serialize message")

	bufReader := bytes.NewReader(buf)

	reader := DefaultMessageBuf()
	err = reader.Read(bufReader)
	assert.NoError(t, err, "Failed to read message from buffer")

	readMsgs := reader.Messages()
	assert.Equal(t, 1, len(readMsgs), "Expected 1 message in the buffer after reading")
	firstMsg := readMsgs[0]
	assert.Equal(t, uint64(120), firstMsg.Offset(), "Expected message offset to be 120")
	assert.Equal(t, []byte("123456789"), firstMsg.Payload(), "Expected message payload to be '123456789'")
	assert.Equal(t, uint32(9), firstMsg.Size(), "Expected message size to be 9 bytes")
}

func TestMessage_ConstructionWithMetadata(t *testing.T) {
	buf, err := serialize(buf{}, 120, []byte("123"), []byte("456789"))
	assert.NoError(t, err, "Failed to serialize message with metadata")
	msgBuf, err := NewMessageBufFromBytes(buf)
	assert.NoError(t, err, "Failed to create message buffer from bytes")

	msgs := msgBuf.Messages()
	assert.Equal(t, 1, len(msgs), "Expected 1 message in the buffer")
	firstMsg := msgs[0]
	assert.Equal(t, uint64(120), firstMsg.Offset(), "Expected message offset to be 120")
	assert.Equal(t, []byte("123"), firstMsg.Metadata(), "Expected message metadata to be '123'")
	assert.Equal(t, []byte("456789"), firstMsg.Payload(), "Expected message payload to be '456789'")
}

func TestMessage_PushWithMetadata(t *testing.T) {
	msgBuf := DefaultMessageBuf()
	err := msgBuf.PushWithMetadata([]byte("123"), []byte("456789"))
	assert.NoError(t, err, "Failed to push message with metadata into buffer")

	msgs := msgBuf.Messages()
	assert.Equal(t, 1, len(msgs), "Expected 1 message in the buffer after pushing with metadata")
	firstMsg := msgs[0]
	assert.Equal(t, []byte("123"), firstMsg.Metadata(), "Expected message metadata to be '123'")
	assert.Equal(t, []byte("456789"), firstMsg.Payload(), "Expected message payload to be '456789'")
}

func TestMessage_ReadInvalidHash(t *testing.T) {
	buf, err := serialize(buf{}, 120, []byte(""), []byte("123456789"))
	assert.NoError(t, err, "Failed to serialize message with invalid hash")

	// mess with the payload such that the hash does not match
	lastIndex := len(buf) - 1
	buf[lastIndex] ^= buf[lastIndex] + 1

	bufReader := bytes.NewReader(buf)

	reader := DefaultMessageBuf()
	err = reader.Read(bufReader)
	assert.Error(t, err, "Expected error when reading message with invalid hash")
	assert.Contains(t, err.Error(), "hash mismatch", "Expected error to contain 'hash mismatch'")
}

func TestMessage_ReadInvalidPayloadLength(t *testing.T) {
	buf, err := serialize(buf{}, 120, []byte(""), []byte("123456789"))
	assert.NoError(t, err, "Failed to serialize message with invalid payload length")

	// pop the last byte
	buf = buf[:len(buf)-1]

	bufReader := bytes.NewReader(buf)
	reader := DefaultMessageBuf()
	err = reader.Read(bufReader)
	assert.Error(t, err, "Expected error when reading message with invalid payload length")
	// will result in a hash mismatch since the payload length is incorrect
	assert.Contains(t, err.Error(), "hash mismatch", "Expected error to contain 'hash mismatch'")
}

func TestMessage_Deserialize(t *testing.T) {
	msgBuf := DefaultMessageBuf()
	err := msgBuf.Push([]byte("foo"))
	assert.NoError(t, err, "Failed to push data into message buffer")
	err = msgBuf.Push([]byte("bar"))
	assert.NoError(t, err, "Failed to push data into message buffer")
	err = msgBuf.Push([]byte("baz"))
	assert.NoError(t, err, "Failed to push data into message buffer")

	setMessagesOffsets(msgBuf, 10)

	bytes := msgBuf.Bytes()

	copiedBytes := make([]byte, len(bytes))
	copy(copiedBytes, bytes)

	deserializedBuf, err := NewMessageBufFromBytes(copiedBytes)
	assert.NoError(t, err, "Failed to deserialize message buffer from bytes")
	assert.Equal(t, msgBuf.Len(), deserializedBuf.Len(), "Expected deserialized message buffer length to match original")

	deserializedMsgs := deserializedBuf.Messages()
	assert.Equal(t, 3, len(deserializedMsgs), "Expected 3 messages in the deserialized buffer")

	firstMsg := deserializedMsgs[0]
	assert.Equal(t, uint64(10), firstMsg.Offset(), "Expected first message offset to be 10")
	assert.Equal(t, []byte("foo"), firstMsg.Payload(), "Expected first message payload to be 'foo'")

	secondMsg := deserializedMsgs[1]
	assert.Equal(t, uint64(11), secondMsg.Offset(), "Expected second message offset to be 11")
	assert.Equal(t, []byte("bar"), secondMsg.Payload(), "Expected second message payload to be 'bar'")

	thirdMsg := deserializedMsgs[2]
	assert.Equal(t, uint64(12), thirdMsg.Offset(), "Expected third message offset to be 12")
	assert.Equal(t, []byte("baz"), thirdMsg.Payload(), "Expected third message payload to be 'baz'")
}
