// GOSC-FILE-START

// Copyright (c) 2026 
//
// File: gosc_util.go
// Brief: Helper functions
//
//
// License: BSD-3-Clause

package runtime

const (
	CallerSkipAtomic           = 2
	CallerSkipAtomicType       = 3
	CallerSkipChanSendRecv     = 3
	CallerSkipChanClose        = 2
	CallerSkipCond             = 2
	CallerSkipMutex            = 2
	CallerSkipSelect           = 3
	CallerSkipSelectRepl       = 2
	CallerSkipSelectOneDef     = 2
	CallerSkipWaitGroupAddWait = 2
	CallerSkipWaitGroupDone    = 3
	CallerSkipOne              = 2
)

// Get a string representation of an uint64
//
// Parameter:
//   - n: int to convert
//
// Returns:
//   - string representation of the int
func uint64ToString(n uint64) string {
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return uint64ToString(n/10) + string(rune(n%10+'0'))
	}
}

// Get a string representation of an int64
//
// Parameter:
//   - n int64: int64 to convert
//
// Returns:
//   - string: string representation of the int64
func int64ToString(n int64) string {
	if n < 0 {
		return "-" + int64ToString(-n)
	}

	if n < 10 {
		return string(rune(n + '0'))
	}

	return int64ToString(n/10) + string(rune(n%10+'0'))
}

// Get a string representation of an int32
//
// Parameter:
//   - n int32: int32 to convert
//
// Returns:
//   - string: string representation of the int32
func int32ToString(n int32) string {
	if n < 0 {
		return "-" + int32ToString(-n)
	}
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return int32ToString(n/10) + string(rune(n%10+'0'))
	}
}

// Get a string representation of an uint32
//
// Parameter:
//   - n uint32: uint32 to convert
//
// Returns:
//   - string representation of the uint32
func uint32ToString(n uint32) string {
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return uint32ToString(n/10) + string(rune(n%10+'0'))
	}
}

// Get a string representation of an int
//
// Parameter:
//   - n int : int to convert
//
// Returns:
//   - string representation of the int
func intToString(n int) string {
	if n < 0 {
		return "-" + intToString(-n)
	}
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return intToString(n/10) + string(rune(n%10+'0'))

	}
}

// Convert a string to an integer. If not possible, this panics
//
// Parameter:
//   - s string: the string to convert
//
// Returns:
//   - int: the int representation
func stringToInt(s string) int {
	var result int
	sign := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '-' && i == 0 {
			sign = -1
		} else if s[i] >= '0' && s[i] <= '9' {
			result = result*10 + int(s[i]-'0')
		} else {
			panic("Invalid input")
		}
	}
	return result * sign
}

// Convert a string to an uint32. If not possible, this panics
//
// Parameter:
//   - s string: the string to convert
//
// Returns:
//   - uint32: the uint32 representation
func stringToUint32(s string) uint32 {
	return uint32(stringToInt(s))
}

// Convert a string to an uint64. If not possible, this panics
//
// Parameter:
//   - s string: the string to convert
//
// Returns:
//   - int: the uint64 representation
func stringToUint64(s string) uint64 {
	return uint64(stringToInt(s))
}

// Get a string representation of a bool
//
// Parameter:
//   - b: bool to convert
//
// Returns:
//   - string representation of the bool (true: "t", false: "f")
func boolToString(b bool) string {
	if b {
		return "t"
	}
	return "f"
}

// Given a value of a number or bool, convert it into its string representation
//
// Parameter:
//   - val any: the value to convert
//
// Returns:
//   - string: the string representation of the value or "" if it was not able to convert
func convToString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return intToString(v)
	case uint:
		return uint64ToString(uint64(v))
	case int32:
		return int32ToString(v)
	case int64:
		return int64ToString(v)
	case uint32:
		return uint32ToString(v)
	case uint64:
		return uint64ToString(v)
	case bool:
		if v {
			return "t"
		}
		return "f"
	}
	return ""
}

// Given a file and line, return a position
//
// Parameter:
//   - file string: the file
//   - line int: the line number
//
// Returns:
//   - string: [file]:[line]
func posToString(file string, line int) string {
	return file + posSep + intToString(line)
}

// Get a pos from a caller
//
// Parameter:
//   - skip int: the skip value as if the Caller was called at the position of the posFromCaller
//
// Returns:
//   - string: the position in the form [file]:[line]
func posFromCaller(skip int) string {
	_, file, line, _ := Caller(skip + 1)
	return file + posSep + intToString(line)
}

// Check if a list contains an element
//
// Parameter:
//   - list []T: list of values
//   - elem T: element to check
//
// Returns:
//   - true if the list contains the element, false otherwise
func IsInSlice[T comparable](list []T, elem T) bool {
	for _, e := range list {
		if e == elem {
			return true
		}
	}
	return false
}

// Check if a string s contains a substring sub
//
// Parameter:
//   - s: the long string
//   - sub: the sub string
//
// Returns:
//   - bool: true if s contains sub as a substring, false otherwise
func containsStr(s, sub string) bool {
	// Get the lengths of both the main string and the substring
	lenS := len(s)
	lenSub := len(sub)

	// If the substring is longer than the string, it can't be a substring
	if lenSub > lenS {
		return false
	}

	// Iterate over the main string `s`
	for i := 0; i <= lenS-lenSub; i++ {
		// Check if substring matches
		match := true
		for j := 0; j < lenSub; j++ {
			if s[i+j] != sub[j] {
				match = false
				break
			}
		}
		// If we found a match, return true
		if match {
			return true
		}
	}

	// No match found, return false
	return false
}

// Check if a string s has a prefix
//
// Parameter:
//   - s: the string
//   - sub: the prefix
//
// Returns:
//   - bool: true if prefix is a prefix of s, false otherwise
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// Check if a string s has a suffix
//
// Parameter:
//   - s: the string
//   - sub: the suffix
//
// Returns:
//   - bool: true if suffix is a suffix of s, false otherwise
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// Given a string, split it by a separator
//
// Parameter:
//   - s string: the string to split
//   - sep rune: the separator
//
// Returns:
//   - []string: the list of strings
func split(s string, sep rune) []string {
	var res []string
	var current string

	for _, char := range s {
		if char == sep {
			res = append(res, current)
			current = ""
		} else {
			current += string(char)
		}
	}

	res = append(res, current)
	return res
}

// printAllGoroutines prints the stack traces of all routines
func printAllGoroutines() {
	buf := make([]byte, 1<<20) // 1 MB buffer
	n := Stack(buf, true)
	println(string(buf[:n]))
}

func mapToSlice[T comparable](m map[T]struct{}) []T {
	res := make([]T, len(m))

	i := 0
	for k := range m {
		res[i] = k
		i++
	}

	return res
}
