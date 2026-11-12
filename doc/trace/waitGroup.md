# WaitGroup

The Add, Done and Wait operations of a wait group are recorded in the trace where the operations occurs.

## Trace element

The basic form of the trace element is

```
W,[tPre],[tPost],[id],[opW],[delta],[val],[pos]
```

where `W` identifies the element as a wait group element. The following sets are set as follows:

- [tPre] $\in\mathbb N: This is the value of the global counter when the operation starts
  the execution of the lock or unlock function
- [tPost] $\in\mathbb N: This is the value of the global counter when the operation has finished
  its operation.
- [id] $\in\mathbb N: This is the unique id identifying this wait group
- [opW]: This filed identifies the operation type that was executed on the wait group:
  - [opW] = `A`: change of the internal counter by delta. This is done by Add or Done.
  - [opW] = `W`: wait on the wait group
- [delta]$\in \mathbb Z$ : This field shows the change of the internal value of the wait group.
  For Add this is a positive number. For Done this is `-1`. For Wait this is always
  `0`.
- [val]$\in \mathbb N_0$ : This field shows the new value of the internal counter after the operation
  finished. This value is always greater or equal 0. For Wait, this field must be `0`.
- [pos]: The last field show the position in the code, where the mutex operation
  was executed. It consists of the file and line number separated by a colon (:)

## Implementation

The recording of the operations is done in the `goPatch/src/sync/waitgroup.go` file in the [Add](../../goPatch/src/sync/waitgroup.go#L53) (Add, Done) and [Wait](../../goPatch/src/sync/waitgroup.go#L138) with
the functions being implemented [here](../../goPatch/src/runtime/gosc_trace_waitgroup.go).

We differentiate between add and done by checking the delta value (done -1, add > 0).
