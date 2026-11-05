# Atomic

The recording of atomics records atomic operations, both on normal types and on atomic types. This includes Add, CompareAndSwap, Swap, Load and Store operations.

## Trace element:

The basic form of the trace element is

```
A,[tPost],[id],[opA],[pos]
```

where `A` identifies the element as an atomic operation.
The other fields are set as follows:

- [tPost]: This field shows the value of the internal counter when the operation is executed.
- [id]: This field shows a number representing the variable. It is not possible to give every variable its own unique, consecutive id. For this reason, this id is equal to the memory position of the variable.
- [opA]: This field shows the type of operation. Those can be
  - `L`: Load
  - `S`: Store
  - `A`: Add
  - `W`: Swap
  - `C`: CompareAndSwap
  - `N`: And
  - `O`: Or
  - `U`: unknown (should not appear)
- [pos]: The last field show the position in the code, where the atomic operation was executed

## Implementation

While for most operations it is possible to directly add the code for
the recording (and the replay) in the implementation of the operation, this
was not possible for the atomics, since they are partially implemented in
go-assembly. We therefore needed to add a layer between the call of a
atomic operation and its assembly execution.
The main signatures for the functions are defined in [sync/atomic/doc.go](../../goPatch/src/sync/atomic/doc.go) and the implementations are in [sync/atomic/asm.s](../../goPatch/src/sync/atomic/asm.s) (in practice the function in the asm.s file only jump to the actual architecture specific implementations in runtime/atomic/, but for here this is not relevant). To add the recording and replay code, we rename all functions in
the [doc](../../goPatch/src/sync/atomic/doc.go) and [asm](../../goPatch/src/sync/atomic/asm.s) files to [oldName]advocatego. The same is also done in the
[doc_32](../../goPatch/src/sync/atomic/doc_32.go) and [doc_64](../../goPatch/src/sync/atomic/doc_64.go) files. We add two new files [advocate_atomic.go](../../goPatch/src/sync/atomic/advocate_atomic.go) and [advocate_atomic_type.go](../../goPatch/src/sync/atomic/advocate_atomic_type.go). In them, we now implement a function for each of the
atomic functions with the original function name. Those functions
contain the code for the replay and the [recording](../../goPatch/src/runtime/advocate_trace_atomic.go) and then a call
to the renamed original function for the operations. For example, the
SwapInt32 function is now implemented as

```go
func SwapInt32(addr *int32, new int32) (old int32) {
	// replay
	chWait, chAck := runtime.WaitForReplay(runtime.OperationAtomicSwap, 2, true)
	if chWait != nil {
		<-chWait
	}

	// recording
	runtime.AdvocateAtomic(addr, runtime.SwapOp, 2)

	if chAck != nil {
		chAck <- struct{}{}
	}()

	// original function
	return SwapInt32Advocate(addr, new)
}
```

The two files contain almost the same code. The only difference is that
while the [advocate_atomic.go](../../goPatch/src/sync/atomic/advocate_atomic.go)
file contains the functions with the original atomic function names as
shown in the example,
in [advocate_atomic_type.go](../../goPatch/src/sync/atomic/advocate_atomic_type.go),
the names are all of the form [functionName]AdvocateType. Additionally the
skip value used to determine the position of the function call (in the example 2)
is increased by one. While the functions in [advocate_atomic.go](../../goPatch/src/sync/atomic/advocate_atomic.go) are meant to be used directly by the user,
the function in the [advocate_atomic_type.go](../../goPatch/src/sync/atomic/advocate_atomic_type.go) file are used for the implementation on the
atomic types defined in [types.go](../../goPatch/src/sync/atomic/types.go).
This additional function call makes it necessary to change the skip value.

To get this to work, we need to [disable the sync/atomic aliases](../../goPatch/src/cmd/compile/internal/ssagen/intrinsics.go#L1571) in the
[SSA](https://go.dev/src/cmd/compile/internal/ssa/README) (Static Single Assignment) implementation. Additionally, we needed to add the atomic functions
using Uintptr to the [blockedLinknames](../../goPatch/src/cmd/link/internal/loader/loader.go#L2335) map in the link loader, since they otherwise break the linkname system.
