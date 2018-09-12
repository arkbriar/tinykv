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
	"context"
	"math"
	"math/rand"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
)

// CacheLineSize defines the common cache line size to 64
const CacheLineSize = 64

// init function checks if epochEntry aligns to cache line size
func init() {
	if unsafe.Sizeof(epochEntry{}) != CacheLineSize {
		logrus.Fatal("Epoch entry is not size of cache line! (sizeof(epochEntry) != 64)")
	}
}

// Phases, used internally by FASTER to keep track of how far along FASTER
// has gotten during checkpoint, gc, and grow actions
const (
	PrepareIndexCheckPhase   uint32 = 0x0
	IndexCheckPhase          uint32 = 0x1
	PreparePhase             uint32 = 0x2
	InProgressPhase          uint32 = 0x3
	WaitPendingPhase         uint32 = 0x4
	WaitFlushPhase           uint32 = 0x5
	RestPhase                uint32 = 0x6
	PersistenceCallbackPhase uint32 = 0x7
	GcIoPendingPhase         uint32 = 0x8
	GcInProgressPhase        uint32 = 0x9
	GrowPreparePhase         uint32 = 0xa
	GrowInProgressPhase      uint32 = 0xb
	InvalidPhase             uint32 = math.MaxUint32
)

// epochEntry is the entry in epoch table
type epochEntry struct {
	localCurrentEpoch   uint64
	reentrant           uint32
	atomicPhaseFinished uint32

	// padding to make sure this entry is not shared between cache lines
	_ [CacheLineSize - 16]byte
}

// Initializes the epochEntry
func (entry *epochEntry) initialize() {
	entry.localCurrentEpoch = 0
	entry.reentrant = 0
	entry.atomicPhaseFinished = RestPhase
}

// States to use in epochAction indicate free and locked
const (
	epochFree   = math.MaxUint64
	epochLocked = math.MaxUint64 - 1
)

// EpochCallbackFunc represents the callback function type
type EpochCallbackFunc func(ctx *context.Context)

type epochAction struct {
	// the epoch field is atomic -- always read it first and write it last
	atomicEpoch uint64
	callback    EpochCallbackFunc
	ctx         *context.Context
}

func (action *epochAction) initialize() {
	action.ctx = nil
	action.callback = nil
	atomic.StoreUint64(&action.atomicEpoch, epochFree)
}

func (action *epochAction) isFree() bool {
	return atomic.LoadUint64(&action.atomicEpoch) == epochFree
}

func (action *epochAction) tryPop(expectedEpoch uint64) bool {
	retVal := atomic.CompareAndSwapUint64(&action.atomicEpoch, expectedEpoch, epochLocked)
	if retVal {
		ctx := action.ctx
		callback := action.callback
		action.callback = nil
		action.ctx = nil
		// release the lock
		atomic.StoreUint64(&action.atomicEpoch, epochFree)
		// perform the action
		callback(ctx)
	}
	return retVal
}

func (action *epochAction) tryPush(priorEpoch uint64, newCallback EpochCallbackFunc, newCtx *context.Context) bool {
	retVal := atomic.CompareAndSwapUint64(&action.atomicEpoch,
		epochFree, epochLocked)
	if retVal {
		action.callback = newCallback
		action.ctx = newCtx
		// release the lock
		atomic.StoreUint64(&action.atomicEpoch, priorEpoch)
	}
	return retVal
}

func (action *epochAction) trySwap(expectedEpoch, priorEpoch uint64,
	newCallback EpochCallbackFunc, newCtx *context.Context) bool {
	retVal := atomic.CompareAndSwapUint64(&action.atomicEpoch, expectedEpoch, epochLocked)
	if retVal {
		existingCallback := action.callback
		existingCtx := action.ctx
		action.callback = newCallback
		action.ctx = newCtx
		// release the lock
		atomic.StoreUint64(&action.atomicEpoch, priorEpoch)
		// perform the action
		existingCallback(existingCtx)
	}
	return retVal
}

const (
	// InvalidIndex is the default invalid page index entry
	InvalidIndex = uint32(0)
	// Unprotected indicates that this thread is not protecting any epoch
	Unprotected = uint64(0)
	// drainListSize is the default drain list size
	drainListSize = 256
)

// LightEpoch implements the epoch protection framework proposed in FASTER paper
type LightEpoch struct {
	// current system epoch (global state)
	atomicCurrentEpoch uint64
	// cached value of epoch that is safe to reclaim
	atomicSafeToReclaimEpoch uint64
	// epoch table
	table []*epochEntry
	// table used slice
	tableUsed []int32
	// number of entries in epoch table
	entryNum uint32
	// count of drain actions
	atomicDrainCount uint32
	// list of action, epoch pairs containing actions to performed when an epoch becomes safe to reclaim
	drainList [drainListSize]epochAction
}

// NewLightEpoch creates a new LightEpoch and initializes it
func NewLightEpoch(size uint32) *LightEpoch {
	epoch := &LightEpoch{
		atomicCurrentEpoch:       1,
		atomicSafeToReclaimEpoch: 0,
		table:                    make([]*epochEntry, size+2),
		tableUsed:                make([]int32, size+2),
		entryNum:                 size,
		atomicDrainCount:         0,
	}

	// to avoid gc, do not declare with _, start
	tableBuf, start := alignedAlloc(CacheLineSize, int((size+2)*CacheLineSize))

	for i := uint32(0); i < size+2; i++ {
		// setup and initialize table
		epoch.table[i] = (*epochEntry)(unsafe.Pointer(start + uintptr(CacheLineSize)))
		epoch.table[i].initialize()
	}
	for i := uint32(0); i < drainListSize; i++ {
		epoch.drainList[i].initialize()
	}

	// here we have epoch.table tracing memory allocated to tableBuf, it's safe to delete tableBuf
	if len(tableBuf) > 0 {
		tableBuf = nil
	}

	return epoch
}

// FindAndReserveEmptyEntry finds an empty entry and returns its index
func (epoch *LightEpoch) FindAndReserveEmptyEntry() int {
	startIdx := rand.Int()
	iteration := 0
	for {
		for i := 0; i < int(epoch.entryNum); i++ {
			idx := 1 + (startIdx+i)%int(epoch.entryNum)
			if atomic.CompareAndSwapInt32(&epoch.tableUsed[idx], 0, 1) {
				return idx
			}
			iteration++
		}

		if iteration > int(3*epoch.entryNum) {
			panic("Unable to find an empty entry, try increasing the epoch table size")
		}
	}
}

// ReleaseEntry releases entry to light epoch
func (epoch *LightEpoch) ReleaseEntry(epochIdx int) {
	epoch.table[epochIdx].initialize()
	atomic.StoreInt32(&epoch.tableUsed[epochIdx], 0)
}

// Protect enters the thread into the protected code region
func (epoch *LightEpoch) Protect(entryIdx int) uint64 {
	entry := epoch.table[entryIdx]
	entry.localCurrentEpoch = atomic.LoadUint64(&epoch.atomicCurrentEpoch)
	return entry.localCurrentEpoch
}

// ProtectAndDrain enters the thread into the protected code region and
// processes entries in drain list if possible
func (epoch *LightEpoch) ProtectAndDrain(entryIdx int) uint64 {
	entry := epoch.table[entryIdx]
	entry.localCurrentEpoch = atomic.LoadUint64(&epoch.atomicCurrentEpoch)
	if atomic.LoadUint32(&epoch.atomicDrainCount) > 0 {
		epoch.Drain(entry.localCurrentEpoch)
	}
	return entry.localCurrentEpoch
}

func (epoch *LightEpoch) ReentrantProtect(entryIdx int) uint64 {
	entry := epoch.table[entryIdx]
	entry.reentrant++

	if Unprotected != entry.localCurrentEpoch {
		return entry.localCurrentEpoch
	}

	entry.localCurrentEpoch = atomic.LoadUint64(&epoch.atomicCurrentEpoch)
	return entry.localCurrentEpoch
}

// IsProtected checks if thread is in protected code region
func (epoch *LightEpoch) IsProtected(entryIdx int) bool {
	return epoch.table[entryIdx].localCurrentEpoch != Unprotected
}

// Unprotect exits the thread from the protected code region
func (epoch *LightEpoch) Unprotect(entryIdx int) {
	epoch.table[entryIdx].localCurrentEpoch = Unprotected
}

func (epoch *LightEpoch) ReentrantUnprotect(entryIdx int) {
	entry := epoch.table[entryIdx]
	entry.reentrant--
	if entry.reentrant == 0 {
		entry.localCurrentEpoch = Unprotected
	}
}

func (epoch *LightEpoch) Drain(nextEpoch uint64) {
	epoch.ComputeNewSafeToReclaimEpoch(nextEpoch)
	for i := uint32(0); i < drainListSize; i++ {
		triggerEpoch := atomic.LoadUint64(&epoch.drainList[i].atomicEpoch)
		if triggerEpoch <= atomic.LoadUint64(&epoch.atomicSafeToReclaimEpoch) {
			if epoch.drainList[i].tryPop(triggerEpoch) {
				// atomic decrement epoch.atomicDrainCount by 1
				if atomic.AddUint32(&epoch.atomicDrainCount, ^uint32(0)) == 0 {
					break
				}
			}
		}
	}
}

// BumpCurrentEpoch increments the current epoch (global system state)
func (epoch *LightEpoch) BumpCurrentEpoch() uint64 {
	nextEpoch := atomic.AddUint64(&epoch.atomicCurrentEpoch, 1)
	if atomic.LoadUint32(&epoch.atomicDrainCount) > 0 {
		epoch.Drain(nextEpoch)
	}
	return nextEpoch
}

// BumpCurrentEpochWithCallback increments the current epoch (global system state) and register
// a trigger action for when older epoch becomes safe to reclaim
func (epoch *LightEpoch) BumpCurrentEpochWithCallback(callback EpochCallbackFunc, ctx *context.Context) uint64 {
	priorEpoch := epoch.BumpCurrentEpoch() - 1
	i, j := 0, 0
	for {
		triggerEpoch := atomic.LoadUint64(&epoch.drainList[i].atomicEpoch)
		if triggerEpoch == epochFree {
			if epoch.drainList[i].tryPush(priorEpoch, callback, ctx) {
				atomic.AddUint32(&epoch.atomicDrainCount, 1)
				break
			}
		} else if triggerEpoch <= atomic.LoadUint64(&epoch.atomicSafeToReclaimEpoch) {
			if epoch.drainList[i].trySwap(triggerEpoch, priorEpoch, callback, ctx) {
				break
			}
		}

		i++
		if i == drainListSize {
			i = 0
			j++
			if j == 500 {
				j = 0
				time.Sleep(time.Second)
				logrus.Errorf("Slowdown: Unable to add trigger to epoch")
			}
		}
	}
	return priorEpoch + 1
}

// ComputeNewSafeToReclaimEpoch computes latest epoch that is safe to reclaim, by scanning the epoch table
func (epoch *LightEpoch) ComputeNewSafeToReclaimEpoch(currentEpoch uint64) uint64 {
	oldestOngoingCall := currentEpoch
	for i := uint32(1); i <= epoch.entryNum; i++ {
		entryEpoch := epoch.table[i].localCurrentEpoch
		if entryEpoch != Unprotected && entryEpoch < oldestOngoingCall {
			oldestOngoingCall = entryEpoch
		}
	}
	atomic.StoreUint64(&epoch.atomicSafeToReclaimEpoch, oldestOngoingCall-1)
	return atomic.LoadUint64(&epoch.atomicSafeToReclaimEpoch)
}

// SpinWaitForSafeToReclaim spin until safe-to-reclaim-epoch is larger than expected
func (epoch *LightEpoch) SpinWaitForSafeToReclaim(currentEpoch, safeToReclaimEpoch uint64) {
	for {
		if safeToReclaimEpoch <= epoch.ComputeNewSafeToReclaimEpoch(currentEpoch) {
			break
		}
	}
}

// IsSafeToReclaim checks if this epoch is safe to reclaim
func (epoch *LightEpoch) IsSafeToReclaim(expectedEpoch uint64) bool {
	return expectedEpoch <= atomic.LoadUint64(&epoch.atomicSafeToReclaimEpoch)
}

// ResetPhaseFinished is a CPR checkpoint function
func (epoch *LightEpoch) ResetPhaseFinished() {
	for i := uint32(1); i < epoch.entryNum; i++ {
		atomic.StoreUint32(&epoch.table[i].atomicPhaseFinished, RestPhase)
	}
}

// MarkAndCheckIfFinished complete the specified phase of this thread and checks if it is finished by all threads
func (epoch *LightEpoch) MarkAndCheckIfFinished(entryIdx, phase uint32) bool {
	entry := epoch.table[entryIdx]
	atomic.StoreUint32(&entry.atomicPhaseFinished, phase)
	// check if other threads have reported complete
	for i := uint32(1); i < epoch.entryNum; i++ {
		entryPhase := atomic.LoadUint32(&epoch.table[i].atomicPhaseFinished)
		entryEpoch := epoch.table[i].localCurrentEpoch
		if entryEpoch != 0 && entryPhase != phase {
			return false
		}
	}
	return true
}

// HasThreadFinishedPhase returns true if this thread completed the specified phase (i.e., is it waiting for
// other threads to finish the specified phase, before it can advance the global phase)
func (epoch *LightEpoch) HasThreadFinishedPhase(entryIdx, phase uint32) bool {
	return atomic.LoadUint32(&epoch.table[entryIdx].atomicPhaseFinished) == phase
}
