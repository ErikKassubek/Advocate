// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_atomic.go
// Brief: Copies to use from the sync/atomic/type.go
// 	They are identical to the others except for the WaitForReplay skip counter
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package atomic

import "runtime"

// Copies to use from the sync/atomic/type.go
// They are identical to the others except for the WaitForReplay skip counter

func SwapInt32AdvocateType(addr *int32, new int32) (old int32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt32Advocate(addr, new)
}

func SwapInt64AdvocateType(addr *int64, new int64) (old int64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt64Advocate(addr, new)
}

func SwapUint32AdvocateType(addr *uint32, new uint32) (old uint32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint32Advocate(addr, new)
}

func SwapUint64AdvocateType(addr *uint64, new uint64) (old uint64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint64Advocate(addr, new)
}

func SwapUintptrAdvocateType(addr *uintptr, new uintptr) (old uintptr) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUintptrAdvocate(addr, new)
}

func CompareAndSwapInt32AdvocateType(addr *int32, old, new int32) (swapped bool) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt32Advocate(addr, old, new)
}

func CompareAndSwapInt64AdvocateType(addr *int64, old, new int64) (swapped bool) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt64Advocate(addr, old, new)
}

func CompareAndSwapUint32AdvocateType(addr *uint32, old, new uint32) (swapped bool) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint32Advocate(addr, old, new)
}

func CompareAndSwapUint64AdvocateType(addr *uint64, old, new uint64) (swapped bool) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint64Advocate(addr, old, new)
}

func CompareAndSwapUintptrAdvocateType(addr *uintptr, old, new uintptr) (swapped bool) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUintptrAdvocate(addr, old, new)
}

func AddInt32AdvocateType(addr *int32, delta int32) (new int32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt32Advocate(addr, delta)
}

func AddUint32AdvocateType(addr *uint32, delta uint32) (new uint32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint32Advocate(addr, delta)
}

func AddInt64AdvocateType(addr *int64, delta int64) (new int64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt64Advocate(addr, delta)
}

func AddUint64AdvocateType(addr *uint64, delta uint64) (new uint64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint64Advocate(addr, delta)
}

func AddUintptrAdvocateType(addr *uintptr, delta uintptr) (new uintptr) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUintptrAdvocate(addr, delta)
}

func LoadInt32AdvocateType(addr *int32) (val int32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt32Advocate(addr)
}

func LoadInt64AdvocateType(addr *int64) (val int64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt64Advocate(addr)
}

func LoadUint32AdvocateType(addr *uint32) (val uint32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint32Advocate(addr)
}

func LoadUint64AdvocateType(addr *uint64) (val uint64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint64Advocate(addr)
}

func LoadUintptrAdvocateType(addr *uintptr) (val uintptr) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUintptrAdvocate(addr)
}

func StoreInt32AdvocateType(addr *int32, val int32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt32Advocate(addr, val)
}

func StoreInt64AdvocateType(addr *int64, val int64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt64Advocate(addr, val)
}

func StoreUint32AdvocateType(addr *uint32, val uint32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint32Advocate(addr, val)
}

func StoreUint64AdvocateType(addr *uint64, val uint64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint64Advocate(addr, val)
}

func StoreUintptrAdvocateType(addr *uintptr, val uintptr) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUintptrAdvocate(addr, val)
}

func AndInt64AdvocateType(addr *int64, mask int64) (old int64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt64Advocate(addr, mask)
}

func AndUint64AdvocateType(addr *uint64, mask uint64) (old uint64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint64Advocate(addr, mask)
}

func AndInt32AdvocateType(addr *int32, mask int32) (old int32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt32Advocate(addr, mask)
}

func AndUint32AdvocateType(addr *uint32, mask uint32) (old uint32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint32Advocate(addr, mask)
}

func AndUintptrAdvocateType(addr *uintptr, mask uintptr) (old uintptr) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUintptrAdvocate(addr, mask)
}

func OrInt64AdvocateType(addr *int64, mask int64) (old int64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt64Advocate(addr, mask)
}

func OrUint64AdvocateType(addr *uint64, mask uint64) (old uint64) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint64Advocate(addr, mask)
}

func OrInt32AdvocateType(addr *int32, mask int32) (old int32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt32Advocate(addr, mask)
}

func OrUint32AdvocateType(addr *uint32, mask uint32) (old uint32) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint32Advocate(addr, mask)
}

func OrUintptrAdvocateType(addr *uintptr, mask uintptr) (old uintptr) {
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		close(chAck)
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUintptrAdvocate(addr, mask)
}
