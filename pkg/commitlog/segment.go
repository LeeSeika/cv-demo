package commitlog

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	SegmentFileNameLen uint64 = 20
	SegmentFileNameExt string = ".log"
)

var ErrSegmentFull = errors.New("segment is full")

var VersionOneMagic = []byte{0xff, 0xff}

type Segment struct {
	file *os.File

	path string

	baseOffset uint64

	writePos uint64

	maxBytes uint64
}

func NewSegment(logDir string, baseOffset uint64, maxBytes uint64) (*Segment, error) {
	segPath := filepath.Join(logDir, fmt.Sprintf("%020d", baseOffset)+SegmentFileNameExt)
	file, err := os.OpenFile(
		segPath,
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

	// write magic number
	if _, err := file.Write(VersionOneMagic); err != nil {
		return nil, fmt.Errorf("failed to write magic number: %w", err)
	}

	shouldClose = false

	return &Segment{
		file:       file,
		path:       segPath,
		baseOffset: baseOffset,
		writePos:   2, // magic number is 2 bytes
		maxBytes:   maxBytes,
	}, nil
}

func OpenSegment(segPath string, maxBytes uint64) (*Segment, error) {
	file, err := os.OpenFile(
		segPath,
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
	filename := filepath.Base(segPath)
	if uint64(len(filename)) < SegmentFileNameLen {
		return nil, errors.New("segment file name is invalid")
	}
	baseOffsetStr := filename[:SegmentFileNameLen]
	baseOffset, err := strconv.ParseInt(baseOffsetStr, 10, 64)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()

	// check the magic
	magicBytes := make([]byte, 2)
	n, err := file.ReadAt(magicBytes, 0)
	if err != nil {
		return nil, err
	}
	if n < 2 {
		return nil, fmt.Errorf("invalid magic number length: %d", n)
	}

	if !bytes.Equal(magicBytes, VersionOneMagic) {
		return nil, fmt.Errorf("segment file %s does not contains version 1 magic", filename)
	}

	log.Info().
		Str("segment_filename", filename).
		Int64("base_offset", baseOffset).
		Int64("file_size", fileSize).
		Uint64("max_bytes", maxBytes).
		Msgf("opened segment: %s", filename)

	shouldClose = false

	return &Segment{
		file:       file,
		path:       segPath,
		baseOffset: uint64(baseOffset),
		writePos:   uint64(fileSize),
		maxBytes:   maxBytes,
	}, nil
}

func (s *Segment) StartingOffset() uint64 {
	return s.baseOffset
}

func (s *Segment) Size() uint64 {
	return s.writePos
}

func (s *Segment) ReadSlice(reader LogSliceReader, filePos uint32, size uint32) (*MessageBuf, error) {
	return reader.ReadFrom(s.file, filePos, uint64(size))
}

func (s *Segment) Append(payload MessageSet) (uint64, error) {
	// ensure we have the capacity
	payloadLen := len(payload.Bytes())
	if uint64(payloadLen)+s.writePos > s.maxBytes {
		return 0, ErrSegmentFull
	}

	startPos := s.writePos

	_, err := s.file.Write(payload.Bytes())
	if err != nil {
		return 0, fmt.Errorf("failed to write to segment: %w", err)
	}
	s.writePos += uint64(payloadLen)

	return startPos, nil
}

func (s *Segment) Flush() error {
	err := s.file.Sync()
	if err != nil {
		return fmt.Errorf("failed to flush segment at %s: %w", s.path, err)
	}

	// do not close the file here, as it is still in use

	return nil
}

func (s *Segment) Remove() error {
	if err := s.file.Close(); err != nil {
		return fmt.Errorf("failed to close segment file %s: %w", s.path, err)
	}
	if err := os.Remove(s.path); err != nil {
		return fmt.Errorf("failed to remove segment file %s: %w", s.path, err)
	}
	log.Info().
		Str("segment_path", s.path).
		Msgf("removed segment: %s", s.path)

	return nil
}

func (s *Segment) Truncate(len uint32) error {
	err := s.file.Truncate(int64(len))
	if err != nil {
		return fmt.Errorf("failed to truncate segment %s: %w", s.path, err)
	}
	s.writePos = uint64(len)
	log.Info().
		Str("segment_path", s.path).
		Uint32("length", len).
		Msgf("truncated segment: %s to length %d", s.path, len)
	return nil
}
