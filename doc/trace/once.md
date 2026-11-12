# Once

The Do of an Once is recorded the in trace, when where the operation occures.

# Trace element

The basic form of the trace element is

```
O,[tPre],[tPost],[id],[suc],[pos]
```

where `O` identifies the element as a wait group element. The following
fields are

- [tPre] $\in\mathbb N$: This is the value of the global counter when the operation starts
  the execution of the lock or unlock function
- [tPost] $\in\mathbb N$: This is the value of the global counter when the operation has finished
  its operation.
- [id] $\in\mathbb N$: This is the unique id identifying this once
- [suc] $\in \{t, f\}$ records, whether the function in the once was
  executed (`t`) or not (`f`). Exactly on trace element per once must be `t`.
- [pos]: The last field show the position in the code, where the mutex operation
  was executed. It consists of the file and line number separated by a colon (:)

## Implementation

The recording of the operations is done in the `goPatch/src/sync/once.go` file in the [Do](../../goPatch/src/sync/once.go#L60) function. The recording is done with the [GoscOncePre](../../goPatch/src/runtime/gosc_trace_once.go#40) and [GoscOncePost](../../goPatch/src/runtime/gosc_trace_once.go#68) functions.
