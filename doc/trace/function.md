# Function

Both the start of a function and the return of a function are recorded in the trace.
As an example, see the following.

```go
func f(){
    ...
}

func main() {  // routine 1, line 1
    f()
}
```

Both the call of function `f` and the return of function `f` are recorded.

## Trace element

This will create 2 trace elements.

The call of a function is recorded as

```
F,[t],[name],[posDef],[posCall]
```

where `F` identifies the element as a function call.

- [t] $\in \mathbb N$: This is the time. It is replaced by the int value of the global counter at the moment of the routines creation.
- [posCal]: Position of the function call (f())
- [name]: Name of the function in form "pkgName.funcName" or "pkgName.receiverName.funcName"
- [posDef]: Position of the function declaration
- [posCall]: Position of the function call

The return of a funciton is recorded as 

```
R,[t]
```

where `R` identifies the element as a function call.

- [t] $\in \mathbb N$: This is the time. It is replaced by the int value of the global counter at the moment of the routines creation.

The element does not contain a position. The function corresponding to the return is identified by the last preceding call element that has not been returned (FILO).

## Implementation

The recording of those events cannot simply be implemented in the runtime, 
since Go does not provide function entry and exit hooks.

Instead, we implement it via the compiler. 

During the building of the SSA during the compilation, we add in calls to 
the recording functions implemented in the runtime. 

This is mainly done in [src/cmd/compile/internal/ssagen/ssa.go](../goPatch/src/cmd/compile/internal/ssagen/ssa.go) in the buildssa function.

For the recording of a function call, we implement a function `goscFunctionCall` and insert it into the SSA using

```go
if fn != nil &&
		fn.Sym().Pkg != nil &&
		fn.Sym().Pkg.Path != "runtime" &&
		fn.Pragma&ir.Nosplit == 0 &&
		!fn.Wrapper() &&
		fn.Sym().Name != "goscFunctionCall" &&
        fn.Sym().Name != "goscFunctionReturn" {

		s.rtcall(
			ir.Syms.GoscFunctionCall,
			true,
			nil,
		)
	}
```

We insert it into the compiler using by adding `GoscFunctionCall *obj.LSym` into `symsStruct` in [src/cmd/compile/internal/ir/symtab.go](./goPatch/src/cmd/compile/internal/ir/symtab.go) and initialize it in [src/cmd/compile/internal/ssagen/ssa.go](../goPatch/src/cmd/compile/internal/ssagen/ssa.go) as `ir.Syms.GoscFunctionCall = typecheck.LookupRuntimeFunc("goscFunctionCall")`.
We classify the recording function as as a runtimeDecl in [src/cmd/compile/internal/typecheck/builtin.go](./goPatch/src/cmd/compile/internal/typecheck/builtin.go).

For the return of a function we create the corresponding elements for the goscFunctionReturn function and add it into the `exit` function in the SSA creation.