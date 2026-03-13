package commitlog

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"

	"github.com/google/btree"
)

type bTreeNode = interface {
	*entryNode | *indexNode | *segmentNode
}

type Ordering int8

const (
	OrderingLess    = -1
	OrderingEqual   = 0
	OrderingGreater = 1
)

type ReadLimit struct {
	limit uint64
}

func NewReadLimitWithMaxBytes(maxBytes uint64) *ReadLimit {
	return &ReadLimit{
		limit: maxBytes,
	}
}

func DefaultReadLimit() *ReadLimit {
	return NewReadLimitWithMaxBytes(8 * 1024)
}

func binarySearchIndex(indexes []byte, compare func(relOffset uint32, fileOffset uint32) Ordering) (uint32, error) {
	if len(indexes)%int(IndexEntryBytes) != 0 {
		return 0, errors.New("invalid index length")
	}

	i := uint32(0)
	j := uint32(len(indexes)/int(IndexEntryBytes) - 1)

	for i < j {
		// grab midpoint
		mid := i + ((j - i) / 2)
		// read the relative offset at the midpoint
		relMid := mid * uint32(IndexEntryBytes)

		relOffset := binary.LittleEndian.Uint32(indexes[relMid : relMid+4])
		filePos := binary.LittleEndian.Uint32(indexes[relMid+4 : relMid+8])

		result := compare(relOffset, filePos)
		switch result {
		case OrderingEqual:
			return mid, nil
		case OrderingLess:
			i = mid + 1
		case OrderingGreater:
			j = mid
		}
	}

	return i, nil
}

// findNextBack finds the next node in the BTree that is less than the given node.
func findNextBack[T bTreeNode](t *btree.BTreeG[T], target T) T {
	var nextBack T

	t.DescendLessOrEqual(target, func(item T) bool {
		nextBack = item
		return false
	})

	return nextBack
}

// allNodes extracts all items from a tree in order as a slice.
func allNodes[T bTreeNode](t *btree.BTreeG[T]) (out []T) {
	t.Ascend(func(a T) bool {
		out = append(out, a)
		return true
	})
	return
}

// splitOff splits the original BTree at the given splitAt node, returning the left part and right part of the tree.
func splitOff[T bTreeNode](original *btree.BTreeG[T], splitAt T, less btree.LessFunc[T]) (*btree.BTreeG[T], *btree.BTreeG[T]) {
	// split such that original contains [..splitAt), suffix=[splitAt,...]
	right := btree.NewG(BTreeDegree, less)

	original.AscendGreaterOrEqual(splitAt, func(item T) bool {
		right.ReplaceOrInsert(item)
		return true
	})
	left := btree.NewG(BTreeDegree, less)

	original.AscendLessThan(splitAt, func(item T) bool {
		left.ReplaceOrInsert(item)
		return true
	})

	return left, right
}

func readMessageOffset(buf []byte) uint64 {
	if len(buf) < 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(buf[0:8])
}

func readMessageSize(buf []byte) uint32 {
	if len(buf) < 12 {
		return 0
	}
	return binary.LittleEndian.Uint32(buf[8:12])
}

func readMessageHash(buf []byte) uint32 {
	if len(buf) < 16 {
		return 0
	}
	return binary.LittleEndian.Uint32(buf[12:16])
}

func readMessageMetadataSize(buf []byte) uint16 {
	if len(buf) < 20 {
		return 0
	}
	return binary.LittleEndian.Uint16(buf[18:20])
}

func writeMessageOffset(buf []byte, offset uint64) {
	if len(buf) < 8 {
		return
	}
	binary.LittleEndian.PutUint64(buf[0:8], offset)
}

func writeMessageSize(buf []byte, size uint32) {
	if len(buf) < 12 {
		return
	}
	binary.LittleEndian.PutUint32(buf[8:12], size)
}

func writeMessageHash(buf []byte, hash uint32) {
	if len(buf) < 16 {
		return
	}
	binary.LittleEndian.PutUint32(buf[12:16], hash)
}

func writeMessageMetadataSize(buf []byte, size uint16) {
	if len(buf) < 20 {
		return
	}
	binary.LittleEndian.PutUint16(buf[18:20], size)
}

func readN(reader io.Reader, buf []byte, size uint64) (int, error) {
	return reader.Read(buf)
}

func crc32C(data []byte) uint32 {
	return crc32.Checksum(data, crc32.MakeTable(crc32.Castagnoli))
}
