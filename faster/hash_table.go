//  Copyright © 2018 The TinyKV Authors.
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy
//  of this software and associated documentation files (the "Software"), to deal
//  in the Software without restriction, including without limitation the rights
//  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//  copies of the Software, and to permit persons to whom the Software is
//  furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in
//  all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
//  THE SOFTWARE.

package faster

import (
	"github.com/sirupsen/logrus"
	"hash"
	"reflect"
	"sync/atomic"
	"unsafe"
)

type AtomicByteSlice atomic.Value

func (s *AtomicByteSlice) Load() []byte {
	v := (*atomic.Value)(s)
	if slice := v.Load(); slice != nil {
		realSlice, _ := slice.([]byte)
		return realSlice
	}
	return nil
}

func (s *AtomicByteSlice) Store(slice []byte) {
	if slice == nil {
		return
	}

	v := (*atomic.Value)(s)
	v.Store(slice)
}

// 64-bit record header with memory layout like
// | 16-bit unused | 48-bit address |
type RecordHeader uint64

func (hdr *RecordHeader) AtomicLoadAddress() uintptr {
	return uintptr(atomic.LoadUint64((*uint64)(hdr)) & (1<<48 - 1))
}

type Record struct {
	header RecordHeader
	key    string
	value  AtomicByteSlice
}

func (r *Record) AtomicNextRecord() *Record {
	nextRecordUintPtr := r.header.AtomicLoadAddress()
	return (*Record)(unsafe.Pointer(nextRecordUintPtr))
}

// 64 bit entry with memory layout like
// | 1-bit tentative | 1-bit reserved | 14-bit tag | 48-bit address |
type HashBucketEntry uint64

func (entry *HashBucketEntry) Equals(other HashBucketEntry) bool {
	return *entry == other
}

func (entry *HashBucketEntry) GetAddress() uint64 {
	return uint64(*entry) & (1<<48 - 1)
}

func (entry *HashBucketEntry) SetAddress(addr uint64) {
	ptr := (*uint64)(entry)
	*ptr = (*ptr & (uint64(0xffff) << 48)) + (addr & (1<<48 - 1))
}

func (entry *HashBucketEntry) GetTag() uint16 {
	ptr := (*uint64)(entry)
	return uint16((*ptr >> 48) & 0x3fff)
}

func (entry *HashBucketEntry) SetTag(tag uint16) {
	ptr := (*uint64)(entry)
	*ptr = (*ptr)&(1<<48-1) + *ptr&(uint64(1)<<63) + uint64(tag&0x3fff)<<48
}

func (entry *HashBucketEntry) IsTentative() bool {
	return (uint64(*entry) >> 63) != 0
}

func (entry *HashBucketEntry) SetTentative() {
	*(*uint64)(entry) |= uint64(1) << 63
}

func (entry *HashBucketEntry) PointerOfUint64() *uint64 {
	return (*uint64)(entry)
}

func (entry *HashBucketEntry) Uint64Value() uint64 {
	return uint64(*entry)
}

type HashBucket struct {
	Entries       [7]HashBucketEntry
	OverflowEntry HashBucketEntry
}

func init() {
	if unsafe.Sizeof(HashBucket{}) != CacheLineSize {
		logrus.Fatal("Hash bucket is not size of cache line! (sizeof(HashBucket) != 64)")
	}
}

type HashTable struct {
	buckets []HashBucket
	size    uint64
	hash    hash.Hash64

	// TODO implement other things like disk, file, checkout and recovery
}

func NewHashTable(size uint64, hash hash.Hash64) *HashTable {
	if size < 0 || !isPowerOfTwo(size) {
		panic("Hash table size must be power of 2!")
	}

	if hash == nil {
		panic("Hash table must have a hash method!")
	}

	table := &HashTable{
		buckets: nil,
		size:    size,
		hash:    hash,
	}

	bucketBuf, ptr := alignedAlloc(CacheLineSize, size*CacheLineSize)

	// setup buckets
	// here buckets should be initialized with all 0 because bucketBuf is all 0
	// so no additional initialization is required
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&table.buckets))
	hdr.Data = ptr
	hdr.Len = int(size)
	hdr.Cap = int(size)

	// assert table.buckets != nil

	// release bucketBuf
	if len(bucketBuf) >= 0 {
		bucketBuf = nil
	}

	return table
}

func (table *HashTable) Size() uint64 {
	return table.size
}

// size must be a power of 2
func bucketIdxForHash(hash uint64, size uint64) uint64 {
	return hash & (size - 1)
}

func (table *HashTable) GetBucket(key string) *HashBucket {
	// Get the hash for key
	table.hash.Reset()
	table.hash.Write([]byte(key))
	code := table.hash.Sum64()

	idx := bucketIdxForHash(code, table.size)
	return &table.buckets[idx]
}
