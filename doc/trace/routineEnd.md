# End routine

When a routine is ended, an element is added to the trace.
The element is not added if the routine is ended because of a panic in another routine or because the main routine terminated.

# Trace element

The basic form of the trace element is

```
E,[t]
```

where `E` identifies the element as a routine end element. The
fields are

- [t] $\in\mathbb N: This is the value of the global counter when the routine ended

# Implementation

The call to record the routine is set in runtime/proc.go in the [goexit1](../../goPatch/src/runtime/proc.go#L4327) function using the [GoscRoutineExit](../../goPatch/src/runtime/gosc_trace_routine.go#L67) function.

goexit1 is executed by the [GoExit](../../goPatch/src/runtime/panic.go#L624) function,
which is called every time a routine terminates, and is also responsible
for executing all defers.
