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

import "math"

const CacheLineSize = 64
const MaxNumThreads = 64

type epochEntryValue struct {
	localCurrentEpoch uint64
	reentrant         uint32
}

type epochEntry struct {
	epochEntryValue
	// padding to make sure this entry is not shared between cache lines
	_ [CacheLineSize - 12]byte
}

const (
	epochFree   = math.MaxUint64
	epochLocked = math.MaxUint64 - 1
)

type EpochCallbackFunc func()

type EpochAction struct {
	epoch    uint64
	callback EpochCallbackFunc
}

const (
	InvalidIndex  = uint32(0)
	Unprotected   = uint32(0)
	TableSize     = MaxNumThreads
	DrainListSize = 256
)

type LightEpoch struct {
	currentEpoch       uint64
	safeToReclaimEpoch uint64

	table []epochEntry

	drainCount uint32
	drainList  [DrainListSize]EpochAction
}
