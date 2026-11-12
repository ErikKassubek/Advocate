# Select

Channel operations including select statements are recorded in the trace where the select statement is located.

## Info

Select statements with only one case are not recorded as select statements. If the case is an default case, it is not recorded at all. If the case is a send or receive, the select statement is equivalent to just the send/receive statement and therefor recorded as such.

## Trace element

The basic form of the trace element is

```
S,[tPre],[tPost],[id],[cases],[selIndex],[pos]
```

where `S` identifies the element as a select element.
The other fields are set as follows:

- [tPre] $\in \mathbb N$: This is the value of the global counter when the routine gets to the select statement.
- [tPost] $\in \mathbb N$: This is the value of the global counter when the select has finished, meaning if either on of the select cases has successfully send or received or the default case has been executed
- [id]: This field contains the unique id of the select statement
- [cases]: This field shows a list of the available cases in the select. The
  cases are separated by a ~ symbol. The elements are equal to an equivalent
  channel operation without the `tPre`, `tPost` and `pos` field, except that the fields are separated
  by a decimapl point (.) instead of a comma. The operations are ordered in the following way: First all send
  operations, ordered as in the order of the select cases, then all receive operations,
  again ordered as written in the select cases. If a case is on a nil channel, the channel ID is set to \*.
  By checking the the `tPre` of those channel operations are
  all the same as the `tPre` of the select. The `tPost` can be used to find out,
  which case was selected. For the selected and executed case, `tPost` is equal
  to the `tPost` of the select statement. For all other cases it is zero. The channel
  element also includes the operation id `oId`, to connect the sending and
  receiving operations. If the
  select contains a default case, it is denoted with a single `d` as the last
  element in the cases list. If the default value was chosen it is capitalized (`D`).
- [selIndex]: The internal index of the selected case.
- [pos]: The last field show the position in the code, where the select is implemented. It consists of the file and line number separated by a colon (:). It the select only contains one case, the line number is
  equal to the line of this case.

A select statement is only recorded as such, if it contains at least two non-default cases. Otherwise the Go compiler rewrites it
internally as a normal channel operation and is therefore recorded as such. A select case with one non-default and a default case is
only recorded, if the non-default case was chosen.

## Implementation

There are two different select implementations. One is for selects with exactly one case and a default case ([send case](../../goPatch/src/runtime/chan.go#L992) and [recv case](../../goPatch/src/runtime/chan.go#L1059)). The other one is for [all other selects](../../goPatch/src/runtime/select.go#L123).

The implementation for the select with more than one non-default case has been split in the implementation for the [original select](../../goPatch/src/runtime/select.go#L629) and the [select with preferred case](../../goPatch/src/runtime/select.go#L151) (for more info see [here](../replay.md#select)). The entry point for select in the [selectgo](../../goPatch/src/runtime/select.go#L123) function now calls one of the two select versions. The recording is done in the two selects separately, since we need to know the lock order of the cases, but in terms of recording, the two versions are basically identical.

We implement a [Pre](../../goPatch/src/runtime/gosc_trace_select.go#L189) and a [Post](../../goPatch/src/runtime/gosc_trace_select.go#L249) function for the select case with exactly one non default case and separate [Pre](../../goPatch/src/runtime/gosc_trace_select.go#L51) and [Post](../../goPatch/src/runtime/gosc_trace_select.go#L141) functions for all other selects.

The case with only one select case is the simple one. Here the Pre function records the involved channel and whether the case is a send or receive. The post function simple adds the information about whether the default or the non-default case was chosen. Since they basically have the form

```go
if selectnbsend(c, v) {
	... foo
} else {
	... bar
}
```

and

```go
if selected, ok = selectnbrecv(&v, c); selected {
	... foo
} else {
	... bar
}
```

we can easily add the recording function into the implementation and record all relevant information.

For selects with more than one non-default case, the implementation is more complicated: The main problem lies in connecting the internal representation of the cases to the actual cases.
The internal representation works as follows. First, the select representation a slice of scases `cases`. From them, we can get the information about the channel involved in each select case.
In this slice, the cases are ordered in the following way: first all sends, then all receives.
Additionally, the select has the slice `lockOrder`. It contains the indixes of the cases in `cases` with a non nil channel in a pseudo random but consistent order. When iterating over the cases, the select
implementation will always iterate over the cases in the order given by the lock order, meaning as

```go
for _, casei := range lockorder {
	casi := int(casei)
	cas := (*cases)[casi]
	c := cas.c
```

where cas is the representation of the case, and c is the channel involved in the case.
Since the number of sending cases is given as `nsend`, we can determine for a case whether
it is a send or a receive by checking

```go
chanOp := OperationChannelRecv
if casi < nsends {
	chanOp = OperationChannelSend
}
```

To now record the involved channels in the [Pre](../../goPatch/src/runtime/gosc_trace_select.go#L51)
function, we create a map `caseElementMap` from the `casi` to an `GoscTraceChannel`, which is our
internal representation for a channel event, and therefore for on of the select cases.
We now iterate over the lockorder as shown above, and store the id of the
involved channel and the send/recv information in the `GoscTraceChannel`.
We then iterate over the `casi` from 0 to the total number if cases (`ncases`). We
then check if a case with this `casi` is in `caseElementMap`.
If yes, we add the `GoscTraceChannel` to a slice `caseElements`.
If not, we know that it was a case with a nil channel. We than add
a corresponding `GoscTraceChannel` about this nil channel to `caseElements`.

We now have a slice containing the information about each case in the select,
sorted by the `casi`. This is also the same order in which we write
the select channels in the final trace.\
Additionally, we use the Pre function to get the tPre, the select id,
whether it has a default case and the select position in the code.

For the [Post](../../goPatch/src/runtime/gosc_trace_select.go#L141)
function we now get the `selIndex`. This index is either `-1`, if the default
case has been chosen or equal to the `casi` of the selected case.

We now set the tPost and the selIndex value of the select to the correct values.

If a non default case has been selected, we additionally need to update
the representation of the chosen case:\
Since our list of select cases is now sorted by the `casi`, the chosen
case is the case in `caseElements[selIndex]`. We can therefore
set the tPost value of this channel element to the current time,
set the `oId` for the chosen case, and update the `numberSend` or `numberRecv`
value for the involved channel.

We now have all required information an can therefore build the corresponding
trace element.
