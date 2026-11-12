// Copyright (c) 2025 
//
// File: gosc_atomic.go
// Brief: Copies to use from the sync/atomic/type.go
// 	They are identical to the others except for the WaitForReplay skip counter
//
// Created: 2024-02-16
//
// License: BSD-3-Clause

package atomic

import "runtime"

// Copies to use from the sync/atomic/type.go
// They are identical to the others except for the WaitForReplay skip counter

func SwapInt32GoscType(addr *int32, new int32) (old int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt32Gosc(addr, new)
}

func SwapInt64GoscType(addr *int64, new int64) (old int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt64Gosc(addr, new)
}

func SwapUint32GoscType(addr *uint32, new uint32) (old uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint32Gosc(addr, new)
}

func SwapUint64GoscType(addr *uint64, new uint64) (old uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint64Gosc(addr, new)
}

func SwapUintptrGoscType(addr *uintptr, new uintptr) (old uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUintptrGosc(addr, new)
}

func CompareAndSwapInt32GoscType(addr *int32, old, new int32) (swapped bool) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt32Gosc(addr, old, new)
}

func CompareAndSwapInt64GoscType(addr *int64, old, new int64) (swapped bool) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt64Gosc(addr, old, new)
}

func CompareAndSwapUint32GoscType(addr *uint32, old, new uint32) (swapped bool) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint32Gosc(addr, old, new)
}

func CompareAndSwapUint64GoscType(addr *uint64, old, new uint64) (swapped bool) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint64Gosc(addr, old, new)
}

func CompareAndSwapUintptrGoscType(addr *uintptr, old, new uintptr) (swapped bool) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUintptrGosc(addr, old, new)
}

func AddInt32GoscType(addr *int32, delta int32) (new int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt32Gosc(addr, delta)
}

func AddUint32GoscType(addr *uint32, delta uint32) (new uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint32Gosc(addr, delta)
}

func AddInt64GoscType(addr *int64, delta int64) (new int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt64Gosc(addr, delta)
}

func AddUint64GoscType(addr *uint64, delta uint64) (new uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint64Gosc(addr, delta)
}

func AddUintptrGoscType(addr *uintptr, delta uintptr) (new uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUintptrGosc(addr, delta)
}

func LoadInt32GoscType(addr *int32) (val int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt32Gosc(addr)
}

func LoadInt64GoscType(addr *int64) (val int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt64Gosc(addr)
}

func LoadUint32GoscType(addr *uint32) (val uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint32Gosc(addr)
}

func LoadUint64GoscType(addr *uint64) (val uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint64Gosc(addr)
}

func LoadUintptrGoscType(addr *uintptr) (val uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUintptrGosc(addr)
}

func StoreInt32GoscType(addr *int32, val int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt32Gosc(addr, val)
}

func StoreInt64GoscType(addr *int64, val int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt64Gosc(addr, val)
}

func StoreUint32GoscType(addr *uint32, val uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint32Gosc(addr, val)
}

func StoreUint64GoscType(addr *uint64, val uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint64Gosc(addr, val)
}

func StoreUintptrGoscType(addr *uintptr, val uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUintptrGosc(addr, val)
}

func AndInt64GoscType(addr *int64, mask int64) (old int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt64Gosc(addr, mask)
}

func AndUint64GoscType(addr *uint64, mask uint64) (old uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint64Gosc(addr, mask)
}

func AndInt32GoscType(addr *int32, mask int32) (old int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt32Gosc(addr, mask)
}

func AndUint32GoscType(addr *uint32, mask uint32) (old uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint32Gosc(addr, mask)
}

func AndUintptrGoscType(addr *uintptr, mask uintptr) (old uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUintptrGosc(addr, mask)
}

func OrInt64GoscType(addr *int64, mask int64) (old int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt64Gosc(addr, mask)
}

func OrUint64GoscType(addr *uint64, mask uint64) (old uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint64Gosc(addr, mask)
}

func OrInt32GoscType(addr *int32, mask int32) (old int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt32Gosc(addr, mask)
}

func OrUint32GoscType(addr *uint32, mask uint32) (old uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint32Gosc(addr, mask)
}

func OrUintptrGoscType(addr *uintptr, mask uintptr) (old uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUintptrGosc(addr, mask)
}
