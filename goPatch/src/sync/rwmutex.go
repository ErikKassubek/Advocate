// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
	"internal/race"
	"sync/atomic"
	"unsafe"

	// ADVOCATE-START
	"runtime"
	// ADVOCATE-END
)

// There is a modified copy of this file in runtime/rwmutex.go.
// If you make any changes here, see if you should make them there.

// A RWMutex is a reader/writer mutual exclusion lock.
// The lock can be held by an arbitrary number of readers or a single writer.
// The zero value for a RWMutex is an unlocked mutex.
//
// A RWMutex must not be copied after first use.
//
// If any goroutine calls [RWMutex.Lock] while the lock is already held by
// one or more readers, concurrent calls to [RWMutex.RLock] will block until
// the writer has acquired (and released) the lock, to ensure that
// the lock eventually becomes available to the writer.
// Note that this prohibits recursive read-locking.
// A [RWMutex.RLock] cannot be upgraded into a [RWMutex.Lock],
// nor can a [RWMutex.Lock] be downgraded into a [RWMutex.RLock].
//
// In the terminology of [the Go memory model],
// the n'th call to [RWMutex.Unlock] “synchronizes before” the m'th call to Lock
// for any n < m, just as for [Mutex].
// For any call to RLock, there exists an n such that
// the n'th call to Unlock “synchronizes before” that call to RLock,
// and the corresponding call to [RWMutex.RUnlock] “synchronizes before”
// the n+1'th call to Lock.
//
// [the Go memory model]: https://go.dev/ref/mem
type RWMutex struct {
	w           Mutex        // held if there are pending writers
	writerSem   uint32       // semaphore for writers to wait for completing readers
	readerSem   uint32       // semaphore for readers to wait for completing writers
	readerCount atomic.Int32 // number of pending readers
	readerWait  atomic.Int32 // number of departing readers

	// ADVOCATE-START
	id uint64 // id for the mutex
	// ADVOCATE-END
}

const rwmutexMaxReaders = 1 << 30

// Happens-before relationships are indicated to the race detector via:
// - Unlock  -> Lock:  readerSem
// - Unlock  -> RLock: readerSem
// - RUnlock -> Lock:  writerSem
//
// The methods below temporarily disable handling of race synchronization
// events in order to provide the more precise model above to the race
// detector.
//
// For example, atomic.AddInt32 in RLock should not appear to provide
// acquire-release semantics, which would incorrectly synchronize racing
// readers, thus potentially missing races.

// RLock locks rw for reading.
//
// It should not be used for recursive read locking; a blocked Lock
// call excludes new readers from acquiring the lock. See the
// documentation on the [RWMutex] type.
func (rw *RWMutex) RLock() {
	// ADVOCATE-START
	chWait, chAck := runtime.WaitForReplay(runtime.OperationRWMutexRLock, runtime.CallerSkipMutex)
	if chWait != nil {
		<-chWait
	}

	// AdvocateMutexReq records, that a routine tries to lock a mutex.
	// AdvocatePost is called, if the mutex was locked successfully.
	// In this case, the Lock event in the trace is updated to include
	// this information. advocateIndex is used for AdvocatePost to find the
	// pre event.
	advocateIndex := runtime.AdvocateMutexReq(unsafe.Pointer(rw), rw.id, runtime.OperationRWMutexRLock)
	// ADVOCATE-END

	if race.Enabled {
		race.Read(unsafe.Pointer(&rw.w))
		race.Disable()
	}

	// ADVOCATE-START
	if chAck != nil {
		close(chAck)
	}
	// ADVOCATE-END

	if rw.readerCount.Add(1) < 0 {
		// A writer is pending, wait for it.
		runtime_SemacquireRWMutexR(&rw.readerSem, false, 0)
	}
	if race.Enabled {
		race.Enable()
		race.Acquire(unsafe.Pointer(&rw.readerSem))
	}

	//ADVOCATE-START
	runtime.AdvocateMutexCom(advocateIndex, true)
	// ADVOCATE-END
}

// TryRLock tries to lock rw for reading and reports whether it succeeded.
//
// Note that while correct uses of TryRLock do exist, they are rare,
// and use of TryRLock is often a sign of a deeper problem
// in a particular use of mutexes.
func (rw *RWMutex) TryRLock() bool {
	// ADVOCATE-START
	chWait, chAck := runtime.WaitForReplay(runtime.OperationRWMutexTryRLock, runtime.CallerSkipMutex)
	if chWait != nil {
		<-chWait
	}

	// AdvocateMutexReq records, that a routine tries to lock a mutex.
	// advocateIndex is used for AdvocateMutexCom to find the pre event.
	advocateIndex := runtime.AdvocateMutexReq(unsafe.Pointer(rw), rw.id, runtime.OperationRWMutexTryRLock)
	// ADVOCATE-END

	if race.Enabled {
		race.Read(unsafe.Pointer(&rw.w))
		race.Disable()
	}

	// ADVOCATE-START
	if chAck != nil {
		close(chAck)
	}
	// ADVOCATE-END

	for {
		c := rw.readerCount.Load()
		if c < 0 {
			if race.Enabled {
				race.Enable()
			}

			// ADVOCATE-START
			// If the mutex was not locked successfully, AdvocateMutexCom is called
			// to update the trace.
			runtime.AdvocateMutexCom(advocateIndex, false)
			// ADVOCATE-END
			return false
		}
		if rw.readerCount.CompareAndSwap(c, c+1) {
			if race.Enabled {
				race.Enable()
				race.Acquire(unsafe.Pointer(&rw.readerSem))
			}

			// ADVOCATE-START
			// If the mutex was locked successfully, AdvocateMutexCom is called
			// to update the trace.
			runtime.AdvocateMutexCom(advocateIndex, true)
			// ADVOCATE-END
			return true
		}
	}
}

// RUnlock undoes a single [RWMutex.RLock] call;
// it does not affect other simultaneous readers.
// It is a run-time error if rw is not locked for reading
// on entry to RUnlock.
func (rw *RWMutex) RUnlock() {
	// ADVOCATE-START
	chWait, chAck := runtime.WaitForReplay(runtime.OperationRWMutexRUnlock, runtime.CallerSkipMutex)
	if chWait != nil {
		<-chWait
	}

	// AdvocateMutexReq is used to record the unlocking of a mutex.
	// AdvocatePost records the successful unlocking of a mutex.
	advocateIndex := runtime.AdvocateMutexReq(unsafe.Pointer(rw), rw.id, runtime.OperationRWMutexRUnlock)
	// ADVOCATE-END

	if race.Enabled {
		race.Read(unsafe.Pointer(&rw.w))
		race.ReleaseMerge(unsafe.Pointer(&rw.writerSem))
		race.Disable()
	}

	// ADVOCATE-START
	if chAck != nil {
		close(chAck)
	}
	// ADVOCATE-END

	if r := rw.readerCount.Add(-1); r < 0 {
		// Outlined slow-path to allow the fast-path to be inlined
		rw.rUnlockSlow(r)
	}
	if race.Enabled {
		race.Enable()
	}

	// ADVOCATE-START
	runtime.AdvocateMutexCom(advocateIndex, true)
	// ADVOCATE-END
}

func (rw *RWMutex) rUnlockSlow(r int32) {
	if r+1 == 0 || r+1 == -rwmutexMaxReaders {
		race.Enable()
		fatal("sync: RUnlock of unlocked RWMutex")
	}
	// A writer is pending.
	if rw.readerWait.Add(-1) == 0 {
		// The last reader unblocks the writer.
		runtime_Semrelease(&rw.writerSem, false, 1)
	}
}

// Lock locks rw for writing.
// If the lock is already locked for reading or writing,
// Lock blocks until the lock is available.
func (rw *RWMutex) Lock() {
	// ADVOCATE-START
	chWait, chAck := runtime.WaitForReplay(runtime.OperationRWMutexLock, runtime.CallerSkipMutex)
	if chWait != nil {
		<-chWait
	}

	// AdvocateMutexReq records, that a routine tries to lock a mutex.
	// AdvocatePost is called, if the mutex was locked successfully.
	// In this case, the Lock event in the trace is updated to include
	// this information. advocateIndex is used for AdvocatePost to find the
	// pre event.
	advocateIndex := runtime.AdvocateMutexReq(unsafe.Pointer(rw), rw.id, runtime.OperationRWMutexLock)
	// ADVOCATE-END

	if race.Enabled {
		race.Read(unsafe.Pointer(&rw.w))
		race.Disable()
	}

	// ADVOCATE-START
	if chAck != nil {
		close(chAck)
	}
	// ADVOCATE-END

	// First, resolve competition with other writers.
	rw.w.Lock()
	// Announce to readers there is a pending writer.
	r := rw.readerCount.Add(-rwmutexMaxReaders) + rwmutexMaxReaders
	// Wait for active readers.
	if r != 0 && rw.readerWait.Add(r) != 0 {
		runtime_SemacquireRWMutex(&rw.writerSem, false, 0)
	}
	if race.Enabled {
		race.Enable()
		race.Acquire(unsafe.Pointer(&rw.readerSem))
		race.Acquire(unsafe.Pointer(&rw.writerSem))
	}

	// ADVOCATE-START
	runtime.AdvocateMutexCom(advocateIndex, true)
	// ADVOCATE-END
}

// TryLock tries to lock rw for writing and reports whether it succeeded.
//
// Note that while correct uses of TryLock do exist, they are rare,
// and use of TryLock is often a sign of a deeper problem
// in a particular use of mutexes.
func (rw *RWMutex) TryLock() bool {
	// ADVOCATE-START
	chWait, chAck := runtime.WaitForReplay(runtime.OperationRWMutexTryLock, runtime.CallerSkipMutex)
	if chWait != nil {
		<-chWait
	}

	// AdvocateMutexReq records, that a routine tries to lock a mutex.
	// advocateIndex is used for AdvocateMutexCom to find the pre event.
	advocateIndex := runtime.AdvocateMutexReq(unsafe.Pointer(rw), rw.id, runtime.OperationRWMutexTryLock)
	// ADVOCATE-END

	if race.Enabled {
		race.Read(unsafe.Pointer(&rw.w))
		race.Disable()
	}

	// ADVOCATE-START
	if chAck != nil {
		close(chAck)
	}
	// ADVOCATE-END

	if !rw.w.TryLock() {
		if race.Enabled {
			race.Enable()
		}

		// ADVOCATE-START
		// If the mutex was not locked successfully, AdvocateMutexCom is called
		// to update the trace.
		runtime.AdvocateMutexCom(advocateIndex, false)
		// ADVOCATE-END

		return false
	}
	if !rw.readerCount.CompareAndSwap(0, -rwmutexMaxReaders) {
		rw.w.Unlock()
		if race.Enabled {
			race.Enable()
		}
		// ADVOCATE-START
		// If the mutex was not locked successfully, AdvocateMutexCom is called
		// to update the trace.
		runtime.AdvocateMutexCom(advocateIndex, false)
		// ADVOCATE-END

		return false
	}
	if race.Enabled {
		race.Enable()
		race.Acquire(unsafe.Pointer(&rw.readerSem))
		race.Acquire(unsafe.Pointer(&rw.writerSem))
	}

	// ADVOCATE-START
	// If the mutex was locked successfully, AdvocateMutexCom is called
	// to update the trace.
	runtime.AdvocateMutexCom(advocateIndex, true)
	// ADVOCATE-END

	return true
}

// Unlock unlocks rw for writing. It is a run-time error if rw is
// not locked for writing on entry to Unlock.
//
// As with Mutexes, a locked [RWMutex] is not associated with a particular
// goroutine. One goroutine may [RWMutex.RLock] ([RWMutex.Lock]) a RWMutex and then
// arrange for another goroutine to [RWMutex.RUnlock] ([RWMutex.Unlock]) it.
func (rw *RWMutex) Unlock() {
	// ADVOCATE-START
	chWait, chAck := runtime.WaitForReplay(runtime.OperationRWMutexUnlock, runtime.CallerSkipMutex)
	if chWait != nil {
		<-chWait
	}
	// AdvocateMutexReq is used to record the unlocking of a mutex.
	// AdvocatePost records the successful unlocking of a mutex.
	// For non rw mutexe, the unlock cannot fail. Therefore it is not
	// strictly necessary to record the post for the unlocking of a mutex.
	advocateIndex := runtime.AdvocateMutexReq(unsafe.Pointer(rw), rw.id, runtime.OperationMutexUnlock)
	// ADVOCATE-END

	if race.Enabled {
		race.Read(unsafe.Pointer(&rw.w))
		race.Release(unsafe.Pointer(&rw.readerSem))
		race.Disable()
	}

	// ADVOCATE-START
	if chAck != nil {
		close(chAck)
	}
	// ADVOCATE-END

	// Announce to readers there is no active writer.
	r := rw.readerCount.Add(rwmutexMaxReaders)
	if r >= rwmutexMaxReaders {
		race.Enable()
		// ADVOCATE-START
		panic("sync: Unlock of unlocked RWMutex")
		// ADVOCATE-END
	}
	// Unblock blocked readers, if any.
	for i := 0; i < int(r); i++ {
		runtime_Semrelease(&rw.readerSem, false, 0)
	}
	// Allow other writers to proceed.
	rw.w.Unlock()
	if race.Enabled {
		race.Enable()
	}

	// ADVOCATE-START
	runtime.AdvocateMutexCom(advocateIndex, true)
	// ADVOCATE-END
}

// syscall_hasWaitingReaders reports whether any goroutine is waiting
// to acquire a read lock on rw. This exists because syscall.ForkLock
// is an RWMutex, and we can't change that without breaking compatibility.
// We don't need or want RWMutex semantics for ForkLock, and we use
// this private API to avoid having to change the type of ForkLock.
// For more details see the syscall package.
//
//go:linkname syscall_hasWaitingReaders syscall.hasWaitingReaders
func syscall_hasWaitingReaders(rw *RWMutex) bool {
	r := rw.readerCount.Load()
	return r < 0 && r+rwmutexMaxReaders > 0
}

// RLocker returns a [Locker] interface that implements
// the [Locker.Lock] and [Locker.Unlock] methods by calling rw.RLock and rw.RUnlock.
func (rw *RWMutex) RLocker() Locker {
	return (*rlocker)(rw)
}

type rlocker RWMutex

func (r *rlocker) Lock()   { (*RWMutex)(r).RLock() }
func (r *rlocker) Unlock() { (*RWMutex)(r).RUnlock() }
