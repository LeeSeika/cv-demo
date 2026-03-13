package commitlog

import (
	"fmt"
	"io"
	"math"
)

const HeaderSize = 20
const MaxMetadataSize = math.MaxUint16

type buf []byte

type MessageSet interface {
	Bytes() []byte
	Len() uint64
	IsEmpty() bool
	VerifyHashes() error
	Messages() []*Message
}

type MessageBuf struct {
	bytes buf
	len   uint64
}

func NewMessageBufFromBytes(bytes []byte) (*MessageBuf, error) {
	msgCount := 0

	for offset := 0; offset < len(bytes); {
		if len(bytes)-offset < HeaderSize {
			return nil, fmt.Errorf("not enough bytes for the full message(header) at offset %d", offset)
		}
		size := readMessageSize(bytes[offset:])
		nextMessageOffset := offset + HeaderSize + int(size)
		if len(bytes) < nextMessageOffset {
			return nil, fmt.Errorf("not enough bytes for the full message(payload) at offset %d", offset)
		}

		// check hash
		payload := bytes[offset+HeaderSize : nextMessageOffset]
		target := readMessageHash(bytes[offset:])
		if target == 0 {
			return nil, fmt.Errorf("message at offset %d has zero hash", offset)
		}
		actual := crc32C(payload)
		if actual != target {
			return nil, fmt.Errorf("message at offset %d failed hash verification: expected %d, got %d", offset, target, actual)
		}

		msgCount++
		offset += (HeaderSize + int(size))
	}

	return &MessageBuf{
		bytes: bytes,
		len:   uint64(msgCount),
	}, nil
}

func DefaultMessageBuf() *MessageBuf {
	return &MessageBuf{
		bytes: make([]byte, 0, 1024), // initial capacity of 1024 bytes
		len:   0,
	}
}

func NewMessageBufWithCapacity(capacity int) *MessageBuf {
	return &MessageBuf{
		bytes: make([]byte, 0, capacity),
		len:   0,
	}
}

func (mb *MessageBuf) Reset() {
	mb.bytes = mb.bytes[:0] // reset the slice to zero length
	mb.len = 0
}

func (mb *MessageBuf) Push(payload []byte) error {
	meta := make([]byte, 0)
	appendedBuf, err := serialize(mb.bytes, mb.len, meta, payload)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}
	mb.bytes = appendedBuf
	mb.len++
	return nil
}

func (mb *MessageBuf) PushWithMetadata(meta []byte, payload []byte) error {
	// blank offset, expect the log to set the offsets
	// empty metadata
	buf, err := serialize(mb.bytes, 0, meta, payload)
	if err != nil {
		return fmt.Errorf("failed to serialize message with metadata: %w", err)
	}
	mb.bytes = buf
	mb.len++

	return nil
}

func (mb *MessageBuf) Bytes() []byte {
	return mb.bytes
}

func (mb *MessageBuf) Len() uint64 {
	return mb.len
}

func (mb *MessageBuf) IsEmpty() bool {
	return mb.len == 0
}

func (mb *MessageBuf) Messages() []*Message {
	if mb.IsEmpty() {
		return nil
	}
	messages := make([]*Message, 0, mb.len)
	for offset := 0; offset < len(mb.bytes); {
		if len(mb.bytes)-offset < HeaderSize {
			break // not enough bytes for a full message header
		}
		size := readMessageSize(mb.bytes[offset:]) + HeaderSize
		if size < HeaderSize || len(mb.bytes)-offset < int(size) {
			break // not enough bytes for the full message
		}
		msg := &Message{
			bytes: mb.bytes[offset : offset+int(size)],
		}

		messages = append(messages, msg)

		offset += int(size)
	}
	return messages
}

func (mb *MessageBuf) VerifyHashes() error {
	if mb.IsEmpty() {
		return nil
	}
	for _, msg := range mb.Messages() {
		if !msg.VerifyHash() {
			return fmt.Errorf("message at offset %d failed hash verification", msg.Offset())
		}
	}
	return nil
}

func (mb *MessageBuf) Read(reader io.Reader) error {
	headerBuf := make([]byte, HeaderSize)
	_, err := readN(reader, headerBuf, HeaderSize)
	if err != nil {
		return fmt.Errorf("failed to read message header: %w", err)
	}

	size := readMessageSize(headerBuf)
	hash := readMessageHash(headerBuf)

	payloadBuf := make([]byte, size)
	_, err = readN(reader, payloadBuf, uint64(size))
	if err != nil {
		return fmt.Errorf("failed to read message payload: %w", err)
	}

	payloadHash := crc32C(payloadBuf)
	if payloadHash != hash {
		return fmt.Errorf("message hash mismatch: expected %d, got %d", hash, payloadHash)
	}

	appendBuf := make([]byte, 0, HeaderSize+size)
	appendBuf = append(appendBuf, headerBuf...)
	appendBuf = append(appendBuf, payloadBuf...)

	mb.bytes = append(mb.bytes, appendBuf...)
	mb.len++

	return nil
}

// Message contains finite-sized binary values with an offset from
// the beginning of the log.
//
// | Bytes       | Encoding          | Value                          |
// | ---------   | ----------------- | ------------------------------ |
// | 0-7         | Little Endian u64 | Offset                         |
// | 8-11        | Little Endian u32 | Payload and Metadata Size      |
// | 12-15       | Little Endian u32 | CRC32C of payload and metadata |
// | 16-17       |                   | Reserved                       |
// | m: 18-19    | Little Endian u16 | Size of metadata               |
// | 20-(20+m-1) |                   | Metadata                       |
// | (20+m)      |                   | Payload                        |
type Message struct {
	bytes []byte
}

func (m *Message) Offset() uint64 {
	return readMessageOffset(m.bytes)
}

func (m *Message) Size() uint32 {
	return readMessageSize(m.bytes)
}

func (m *Message) Hash() uint32 {
	return readMessageHash(m.bytes)
}

func (m *Message) MetadataSize() uint16 {
	return readMessageMetadataSize(m.bytes)
}

func (m *Message) TotalBytes() uint64 {
	return uint64(len(m.bytes))
}

func (m *Message) Payload() []byte {
	if len(m.bytes) < HeaderSize {
		return nil
	}
	payloadSize := m.Size() - uint32(m.MetadataSize())
	if len(m.bytes) < int(HeaderSize+payloadSize) {
		return nil
	}
	return m.bytes[HeaderSize+int(m.MetadataSize()) : HeaderSize+int(m.MetadataSize())+int(payloadSize)]
}

func (m *Message) Metadata() []byte {
	if len(m.bytes) < HeaderSize {
		return nil
	}
	metadataSize := m.MetadataSize()
	if len(m.bytes) < int(HeaderSize+metadataSize) {
		return nil
	}
	return m.bytes[HeaderSize : HeaderSize+metadataSize]
}

func (m *Message) VerifyHash() bool {
	target := m.Hash()
	if target == 0 {
		return false
	}
	payload := m.Payload()
	if payload == nil {
		return false
	}
	actual := crc32C(payload)
	return actual == target
}

func (m *Message) SetOffset(offset uint64) {
	writeMessageOffset(m.bytes, offset)
}

func serialize(
	bytes buf,
	offset uint64,
	meta []byte,
	payload []byte,
) (buf, error) {
	metaSize := uint64(len(meta))
	if metaSize > MaxMetadataSize {
		return nil, fmt.Errorf("metadata size %d exceeds maximum %d", len(meta), MaxMetadataSize)
	}

	payloadSize := uint64(len(payload))
	appendSize := HeaderSize + metaSize + payloadSize

	if bytes.remaining() < uint64(appendSize) {
		return nil, fmt.Errorf("not enough space in the buffer, required %d bytes, remaining %d bytes", appendSize, bytes.remaining())
	}

	toHash := make([]byte, 0, metaSize+payloadSize)
	toHash = append(toHash, meta...)
	toHash = append(toHash, payload...)

	headerBuf := make([]byte, HeaderSize)
	writeMessageOffset(headerBuf, offset)
	writeMessageSize(headerBuf, uint32(metaSize+payloadSize))
	writeMessageHash(headerBuf, crc32C(toHash))
	writeMessageMetadataSize(headerBuf, uint16(metaSize))

	// avoid unnecessary multiple allocations by preallocating the buffer
	buf := make([]byte, 0, appendSize)
	buf = append(buf, headerBuf...)
	buf = append(buf, meta...)
	buf = append(buf, payload...)

	bytes = append(bytes, buf...)

	return bytes, nil
}

func (b *buf) remaining() uint64 {
	var bytes []byte = *b
	return math.MaxInt64 - uint64(len(bytes))
}

func setMessagesOffsets(msgSet MessageSet, startOffset Offset) {
	offset := uint64(startOffset)

	msgs := msgSet.Messages()
	for _, msg := range msgs {
		msg.SetOffset(offset)
		offset++
	}
}
