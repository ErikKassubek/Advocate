// Copyright (c) 2025 
//
// File: gosc_atomic.go
// Brief: Atomic functions with replay and recording
//
// Created: 2024-02-16
//
// License: BSD-3-Clause

package atomic

import (
	"runtime"
	"unsafe"
)

// SwapInt32 atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Int32.Swap] instead.
func SwapInt32(addr *int32, new int32) (old int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomic)
	return SwapInt32Gosc(addr, new)
}

// SwapInt64 atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Int64.Swap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func SwapInt64(addr *int64, new int64) (old int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomic)
	return SwapInt64Gosc(addr, new)
}

// SwapUint32 atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Uint32.Swap] instead.
func SwapUint32(addr *uint32, new uint32) (old uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomic)
	return SwapUint32Gosc(addr, new)
}

// SwapUint64 atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Uint64.Swap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func SwapUint64(addr *uint64, new uint64) (old uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomic)
	return SwapUint64Gosc(addr, new)
}

// SwapUintptr atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Uintptr.Swap] instead.
func SwapUintptr(addr *uintptr, new uintptr) (old uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomic)
	return SwapUintptrGosc(addr, new)
}

// SwapPointer atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Pointer.Swap] instead.
// func SwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) (old unsafe.Pointer) {
// 	return SwapPointerGosc(addr, new)
// }

// CompareAndSwapInt32 executes the compare-and-swap operation for an int32 value.
// Consider using the more ergonomic and less error-prone [Int32.CompareAndSwap] instead.
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomic)
	return CompareAndSwapInt32Gosc(addr, old, new)
}

// CompareAndSwapInt64 executes the compare-and-swap operation for an int64 value.
// Consider using the more ergonomic and less error-prone [Int64.CompareAndSwap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func CompareAndSwapInt64(addr *int64, old, new int64) (swapped bool) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomic)
	return CompareAndSwapInt64Gosc(addr, old, new)
}

// CompareAndSwapUint32 executes the compare-and-swap operation for a uint32 value.
// Consider using the more ergonomic and less error-prone [Uint32.CompareAndSwap] instead.
func CompareAndSwapUint32(addr *uint32, old, new uint32) (swapped bool) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomic)
	return CompareAndSwapUint32Gosc(addr, old, new)
}

// CompareAndSwapUint64 executes the compare-and-swap operation for a uint64 value.
// Consider using the more ergonomic and less error-prone [Uint64.CompareAndSwap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomic)
	return CompareAndSwapUint64Gosc(addr, old, new)
}

// CompareAndSwapUintptr executes the compare-and-swap operation for a uintptr value.
// Consider using the more ergonomic and less error-prone [Uintptr.CompareAndSwap] instead.
func CompareAndSwapUintptr(addr *uintptr, old, new uintptr) (swapped bool) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomic)
	return CompareAndSwapUintptrGosc(addr, old, new)
}

// // CompareAndSwapPointer executes the compare-and-swap operation for a unsafe.Pointer value.
// // Consider using the more ergonomic and less error-prone [Pointer.CompareAndSwap] instead.
// func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) (swapped bool) {
// 	return CompareAndSwapPointerGosc(addr, old, new)
// }

// AddInt32 atomically adds delta to *addr and returns the new value.
// Consider using the more ergonomic and less error-prone [Int32.Add] instead.
func AddInt32(addr *int32, delta int32) (new int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomic)
	return AddInt32Gosc(addr, delta)
}

// AddUint32 atomically adds delta to *addr and returns the new value.
// To subtract a signed positive constant value c from x, do AddUint32(&x, ^uint32(c-1)).
// In particular, to decrement x, do AddUint32(&x, ^uint32(0)).
// Consider using the more ergonomic and less error-prone [Uint32.Add] instead.
func AddUint32(addr *uint32, delta uint32) (new uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomic)
	return AddUint32Gosc(addr, delta)
}

// AddInt64 atomically adds delta to *addr and returns the new value.
// Consider using the more ergonomic and less error-prone [Int64.Add] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func AddInt64(addr *int64, delta int64) (new int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomic)
	return AddInt64Gosc(addr, delta)
}

// AddUint64 atomically adds delta to *addr and returns the new value.
// To subtract a signed positive constant value c from x, do AddUint64(&x, ^uint64(c-1)).
// In particular, to decrement x, do AddUint64(&x, ^uint64(0)).
// Consider using the more ergonomic and less error-prone [Uint64.Add] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func AddUint64(addr *uint64, delta uint64) (new uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomic)
	return AddUint64Gosc(addr, delta)
}

// AddUintptr atomically adds delta to *addr and returns the new value.
// Consider using the more ergonomic and less error-prone [Uintptr.Add] instead.
func AddUintptr(addr *uintptr, delta uintptr) (new uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomic)
	return AddUintptrGosc(addr, delta)
}

// LoadInt32 atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Int32.Load] instead.
func LoadInt32(addr *int32) (val int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	return LoadInt32Gosc(addr)
}

// LoadInt64 atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Int64.Load] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func LoadInt64(addr *int64) (val int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	return LoadInt64Gosc(addr)
}

// LoadUint32 atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Uint32.Load] instead.
func LoadUint32(addr *uint32) (val uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	return LoadUint32Gosc(addr)
}

// LoadUint64 atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Uint64.Load] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func LoadUint64(addr *uint64) (val uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	return LoadUint64Gosc(addr)
}

// LoadUintptr atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Uintptr.Load] instead.
func LoadUintptr(addr *uintptr) (val uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	return LoadUintptrGosc(addr)
}

// LoadPointer atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Pointer.Load] instead.
func LoadPointer(addr *unsafe.Pointer) (val unsafe.Pointer) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomic)
	return LoadPointerGosc(addr)
}

// StoreInt32 atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Int32.Store] instead.
func StoreInt32(addr *int32, val int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
	StoreInt32Gosc(addr, val)
}

// StoreInt64 atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Int64.Store] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func StoreInt64(addr *int64, val int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
	StoreInt64Gosc(addr, val)
}

// StoreUint32 atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Uint32.Store] instead.
func StoreUint32(addr *uint32, val uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
	StoreUint32Gosc(addr, val)
}

// StoreUint64 atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Uint64.Store] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func StoreUint64(addr *uint64, val uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
	StoreUint64Gosc(addr, val)
}

// StoreUintptr atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Uintptr.Store] instead.
func StoreUintptr(addr *uintptr, val uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
	StoreUintptrGosc(addr, val)
}

// StorePointer atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Pointer.Store] instead.
// func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer) {
// 	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
// 	if chWait != nil {
// 		<-chWait
// runtime.AckReplay(key)
// 	}
// 	runtime.GoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomic)
// 	StorePointerGosc(addr, val)
// }

func AndInt64(addr *int64, mask int64) (old int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomic)
	return AndInt64Gosc(addr, mask)
}

func AndUint64(addr *uint64, mask uint64) (old uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomic)
	return AndUint64Gosc(addr, mask)
}

func AndInt32(addr *int32, mask int32) (old int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomic)
	return AndInt32Gosc(addr, mask)
}

func AndUint32(addr *uint32, mask uint32) (old uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomic)
	return AndUint32Gosc(addr, mask)
}

func AndUintptr(addr *uintptr, mask uintptr) (old uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomic)
	return AndUintptrGosc(addr, mask)
}

func OrInt64(addr *int64, mask int64) (old int64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomic)
	return OrInt64Gosc(addr, mask)
}

func OrUint64(addr *uint64, mask uint64) (old uint64) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomic)
	return OrUint64Gosc(addr, mask)
}

func OrInt32(addr *int32, mask int32) (old int32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomic)
	return OrInt32Gosc(addr, mask)
}

func OrUint32(addr *uint32, mask uint32) (old uint32) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomic)
	return OrUint32Gosc(addr, mask)
}

func OrUintptr(addr *uintptr, mask uintptr) (old uintptr) {
	key, chWait := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomic)
	if chWait != nil {
		<-chWait
		runtime.AckReplay(key)
	}
	runtime.GoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomic)
	return OrUintptrGosc(addr, mask)
}
