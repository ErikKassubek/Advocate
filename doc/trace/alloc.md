# New Channel

This element record the creation of a concurreny resource.
For now we only record the creation of the blocking primitives channel, mutex, conditional variable and wait group.

# Trace element

The basic form of the trace element is

```
N,[tPost],[id],[elemType],[num],[pos]
```

where `N` identifies the element as a wait group element. The following
fields are

- [tPost] $\in\mathbb N$: This is the value of the global counter when the channel was created
  its operation.
- [id] $\in\mathbb N$: This is the unique id identifying this channel
- [elemType] = "C": channel, "M": mutex, "D": cond var, "W": wait group
- [qSize] = $\in\mathbb N$: for channel buffer size, otherwise 0
- [pos]: The last field show the position in the code, where the mutex operation
  was executed. It consists of the file and line number separated by a colon (:)

## Implementation

For channels, main implementation of this is done in the [makechan](../../goPatch/src/runtime/chan.go#L200) function by calling the [GoscChanMake](../../goPatch/src/runtime/gosc_trace_new_elem.go#L44) function, that is called if a channel is created with a make. Since a channel can also be created, without a make, the GoscChanMake is also called in the send and receive functions if the id of the channel has not been set yet.


TODO: for others