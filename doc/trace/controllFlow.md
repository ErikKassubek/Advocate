# Controll Flow 

A controll flow statement records `if` and `switches`.

## Trace element

The basic form of the trace element is

```
I,[t],[ct],[numCases],[chosenCase],[pos]
```

where `I` identifies the element as a conditional element (for if).
The other fields are set as follows:

- [t] $\in \mathbb N$: Timestamp
- [ct]: This field shows the operation of the element. Those can be
  - [ct] = `I`: If
  - [ct] = `S`: Switch
  - [ct] = `L`: Loop
- [numCases] $\in \mathbb N$: Number of cases in the if or switch
- [chosenCases] $\in \mathbb N$: Chosen case number in the if or switch (0 based)
- [pos]: The last field show the position in the code, where the mutex operation
  was executed. It consists of the file and line number separated by a colon (:)

## Implementation

Like for function, there is no direct place in the runtime, where we could hook in our recording. We therefore add recording functions into the ir during the compilation. 
The implementation can be mostly found [here](../../goPatch/src/cmd/compile/internal/walk/gosc.go).