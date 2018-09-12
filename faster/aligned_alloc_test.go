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
	"encoding/binary"
	"log"
	"testing"
	"unsafe"
)

func TestAlignedAlloc(t *testing.T) {
	var alignment, size uint64 = 64, 64 * 20
	origin, ptr := alignedAlloc(alignment, size)
	if uint64(ptr)%uint64(alignment) != 0 {
		t.Errorf("allocated ptr is not aligned to %d, ptr is %d", alignment, ptr)
	}

	originPtr := uintptr(getFirstAddress(origin))
	log.Printf("origin ptr is %x, ptr is %x", originPtr, ptr)
	if ptr < originPtr || ptr+uintptr(size) >= originPtr+uintptr(size+alignment)-1 {
		t.Errorf("allocated ptr overflow, origin ptr is %x, ptr is %x", originPtr, ptr)
	}
}

func TestAlignedAllocEpochEntry(t *testing.T) {
	var alignment, size uint64 = 64, 64 * 20
	origin, ptr := alignedAlloc(alignment, size)

	// convert ptr to *epochEntry
	entry := (*epochEntry)(unsafe.Pointer(ptr))
	entry.initialize()

	// assume it is little endian
	offset := unsafe.Offsetof(epochEntry{}.atomicPhaseFinished)
	if binary.LittleEndian.Uint32(origin[offset:]) != RestPhase {
		t.Errorf("entry is not initialized correctly")
	}
}
