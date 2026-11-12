# New Routine

The spawning of a new routine is recorded in the trace. The following is
and example of code where such an trace element is recorded.

```go
func main() {  // routine 1, line 1
    go func() {  // routine 2, line 2
        ...
    }
}
```

In the main routine (routine 1) a new routine (routine 2) is spawned using the
`go` keyword.
This is recorded in the trace of routine 1.

## Trace element

This will create 2 trace elements.

In the routine, where the new routine is created, the following element is added.

```
G,[tPost],[id],[pos]
```

where `G` identifies the element as an routine creation element.

- [tPost] $\in \mathbb N$: This is the time. It is replaced by the int value of the global counter at the moment of the routines creation.
- [id] $\in \mathbb N$: This is the id of the newly created routine. This integer id corresponds with
  the line number, where the trace of this new routine is saved in the trace.
- [pos]: Position in the program, where the spawn was created.

The function called when creating the new routine can be read from the first element in the new routine, which is a [function](./function.md) element.

## Implementation

The element is recorded in the [newproc](../../goPatch/src/runtime/proc.go#L5059) function in the `goPatch/src/runtime/proc.go` file. Unfortunately we cannot use the normal `runtime.Caller` function to determine the code position of
the `go func` command, because the compiler turns a `go` statement into a call of `newproc`, which looses this information.\
But we are still able to get the location using following construction:

```go
pc := sys.GetCallerPC()
f := findfunc(pc)
tracepc := pc
if pc > f.entry() {
    tracepc -= sys.PCQuantum
}
file, line := funcline(f, tracepc)
```

This code looks up function metadata for a PC. With this it is able to get the
file and line information. The `PCQuantum` is the minimum value for a
program counter (1 on x86, 4 on most other systems).

The creation is then recorded in the old routine with [GoscSpawnCaller](../../goPatch/src/runtime/gosc_trace_routine.go#L44).

Here we also create the `goscRoutineInfo` used to store the trace for the
new routine using the [newGoscRoutine](../../goPatch/src/runtime/gosc_routine.go#L44) function.
