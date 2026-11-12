# Mutex

The locking and unlocking of sync.(rw)-mutexes is recorded in the trace where it occurred.

## Trace element

The basic form of the trace element is

```
M,[tPre],[tPost],[id],[rw],[opM],[suc],[pos]
```

where `M` identifies the element as a mutex element.
The other fields are set as follows:

- [tPre] $\in \mathbb N$: This is the value of the global counter when the operation starts
  the execution of the lock or unlock function
- [tPost] $\in \mathbb N$: This is the value of the global counter when the mutex has finished its operation. For lock operations this can be either if the lock was successfully acquired or if the routines continues its execution without
  acquiring the lock in case of a trylock.
- [id] $\in \mathbb N$: This is the unique id identifying this mutex
- [rw]: This field records, whether the mutex is an rw mutex ([rw] = `t`) or not
  ([rw] = `f`)
- [opM]: This field shows the operation of the element. Those can be
  - [opM] = `L`: Lock
  - [opM] = `R`: RLock
  - [opM] = `T`: TryLock
  - [opM] = `Y`: TryRLock
  - [opM] = `U`: Unlock
  - [opM] = `N`: RUnlock
- [suc]: This field is used to determine if an Try(R)Lock was successful ([suc] = `t`)
  or not ([suc] = `f`) in acquiring the mutex. For all other operation it is always
  set to `t`.
- [pos]: The last field show the position in the code, where the mutex operation
  was executed. It consists of the file and line number separated by a colon (:)

## Implementation

The recording of the mutex operations is implemented in the `goPatch/src/sync/mutex.go` and `goPatch/src/sync/rwmutex.go` files in the implementation of the
Lock, RLock, TryLock, TryRLock, Unlock and RUnlock function.\
To save the id of the mutex, a field for the id is added to the `Mutex` and
`RWMutex` structs.\
The recording consist of
two function calls, one at the beginning and one at the end of each function.
The first function call is called before the Operation tries to executed
and records the id ([id]) and type ([rw]) of the involved mutex, the called operation (opM), the position of the operation in the program ([pos]) and the counter at the beginning of the operation ([tPre]).\
The second function call records the success of the operation. This includes
the counter at the end of the operation ([tPost]), the information that the
operation finished ([exec]) and the success of try lock operations ([suc]).

The implementation of those function calls can be found in
`goPatch/src/runtime/gosc_trace.go` in the functions [GoscMutexPre](../../goPatch/src/runtime/gosc_trace_mutex.go#L46) and [GoscMutexPost](../../goPatch/src/runtime/gosc_trace_mutex.go#L77)
