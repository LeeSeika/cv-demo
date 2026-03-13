package commitlog

import "os"

type (
	LogSliceReader interface {
		ReadFrom(file *os.File, filePosition uint32, size uint64) (*MessageBuf, error)
	}

	MessageBufReader struct{}
)

func (m *MessageBufReader) ReadFrom(file *os.File, filePosition uint32, size uint64) (*MessageBuf, error) {
	buf := make([]byte, size)
	_, err := file.ReadAt(buf, int64(filePosition))
	if err != nil {
		return nil, err
	}
	return NewMessageBufFromBytes(buf)
}
