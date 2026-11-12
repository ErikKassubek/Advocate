# Not Executed

We also collect some diagnostic information, namely never executed operations
or select cases. This information is useful for debugging and optimization,
especially if the analyzer is used on tests.

## Usage
This is run automatically by the toolchain.

It creates a folder called `GoscNotExecuted` in the
results directory. This folder contains two files, on containing the
files and lines for all never executed operations (only for operations
we could have recorded) and one for all never executed select cases.

## Result format
The file `GoscNotExecutedOperations.txt` contains the never executed
operations. It consists of one line for each file, followed by the line number(s)
in square brackets, containing the never executed operations. If no operations
where ever executed, the line numbers are replaced by the line `No element in file was executed`.
If all operations in the program where executed, the file contains the line
`All program elements were executed`.

The file `GoscNotExecutedSelectCases.txt` contains the never executed select
cases of executed selects. It consists of one line for each select
statement, that was executed at least once, but where at least one case was never
executed. The line contains the file name and line number of the select statement
and the indices of the never executed cases.
The indices are the internal indices of the go runtime. They have the
following form: First all send cases are numbered in the order they appear in the
select statement, then all receive cases are numbered in the order they appear in
the select statement, and finally the default case. The following example illustrates
this:
```go
select {
  case <-a:     // index 2
  case b <- 1:  // index 0
  case <-c:     // index 3
  case d <- 2:  // index 1
  case <-e:     // index 4
  default:      // index 5
}
```


## Implementation

The detection of never executed operations is done by analyzing the AST of the
program. The AST contains all elements of the program and it is possible to
determine the types of the variables from it. Based on this, all relevant
operations in the AST can be found. The AST also contains the corresponding
file and line number for all tokens. Additionally, the program goes through all
recorded traces and reads all recorded elements, i.e., all elements that have
ever been executed. Then it iterates over all nodes of the AST and checks if
the node represents a potentially recorded operation and if it is present in at
least one of the traces. If it is not, the operation is marked as not executed.

The select analysis is done directly on the traces. For each
select in the program (identified by file name and line) with n cases, a list of
n bools is created, one for each case (default cases are simply considered as
another case). The program iterates over all selects in the traces and sets the
bool corresponding to the executed case to true. At the end, it checks if there
are lists that still contain false. These indicate that a case has never been
executed.
