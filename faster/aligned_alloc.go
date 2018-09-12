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

import "unsafe"

func getFirstAddress(buf []byte) unsafe.Pointer {
	if buf != nil && len(buf) != 0 {
		return unsafe.Pointer(&buf[0])
	}
	return nil
}

func alignedAlloc(alignment, size int) (origin []byte, start uintptr) {
	if alignment <= 0 || size <= 0 {
		return nil, 0
	}

	origin = make([]byte, size+alignment-1)
	start = uintptr(unsafe.Pointer(&origin[0]))
	if uint64(start)%uint64(alignment) != 0 {
		offset := alignment - int(uint64(start)%uint64(alignment))
		start = start + uintptr(offset)
	}
	return origin, start
}
