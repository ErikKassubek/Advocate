// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_replay.gotrIt
// Brief: Advocate Replay
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package advocatego

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	_ "unsafe"
)

const (
	activeFile = "replay_constratins.log"
	spawnFile  = "replay_forktree.log"
	posSep     = "#"
)

var timeout = false
var tracePathRewritten = ""

var initReplay bool

// InitReplay reads the trace from the trace folder.
// The function reads all files in the trace folder and adds the trace to the runtime.
// The trace is added to the runtime by calling the AddReplayTrace function.
//
// Parameter:
//   - tracePath string: The path to the rewritten trace, relative the the dir where advocateResult is placed
//   - timeout int: Timeout in seconds, 0: no timeout
//   - atomic bool: if true, replay includes atomic
//   - init bool: true if called from init, false for main
//
//go:linkname InitReplay runtime.AdvocateInitReplay
func InitReplay(tracePath string, timeout int, init bool) {
	if initReplay { // called by main but alredy run by init
		return
	}
	initReplay = true

	FinishFunc = FinishReplay

	tracePathRewritten = tracePath

	startReplay(timeout)
}

// startReplay reads the trace from the files. It also start the timeout for
// the replay
//
// Parameter:
//   - timeout int: timeout in seconds
func startReplay(timeout int) {
	// if printDebug {
	// 	println("Reading constraints from " + tracePathRewritten)
	// }

	// if trace folder does not exist, panic
	if _, err := os.Stat(tracePathRewritten); os.IsNotExist(err) {
		panic("Trace folder '" + tracePathRewritten + "' does not exist.")
	}

	// check for and if exists, read the rewrite_constratins.log file
	readReplayConstraint(tracePathRewritten)
	readSpawnTree(tracePathRewritten)

	if timeout > 0 {
		// start time timeout
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			runtime.ExitReplayTimeout()
			os.Exit(10)
		}()
	}

	go runtime.Timeout()

	// DetectBlockingGC(0)

	runtime.EnableReplay()
}

// readReplayConstraint checks if a rewrite_constratins.log file exists in the path
// if not, it just returns
// If the file exists, it is read and the replay is set to partial replay
//
// Parameter
//   - tracePathRewritten string: the path to the trace folder
//
// Returns
//   - int: number of constraint operations
func readReplayConstraint(tracePathRewritten string) int {
	path := filepath.Join(tracePathRewritten, activeFile)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		panic(err)
	}
	defer file.Close()

	runtime.NewConstraint()

	numberConstraints := 0

	firstLine := true

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		elem := scanner.Text()
		if elem == "" {
			continue
		}

		if firstLine {
			firstLine = false
			runtime.AddConstraintInfo(elem)
			continue
		}

		runtime.AddConstraintElemFromString(elem)
		numberConstraints++
	}

	return numberConstraints
}

// readReplayConstraint checks if a rewrite_spawntree.log file exists in the path
// if not, it just returns
// If the file exists, it is read and the spawn is set to correct values
//
// Parameter
//   - tracePathRewritten string: the path to the trace folder
func readSpawnTree(tracePathRewritten string) {
	path := filepath.Join(tracePathRewritten, spawnFile)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			println("Could not read spawn tree file")
			return
		}
		panic(err)
	}
	defer file.Close()

	spawns := make(map[int][]int)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		elem := scanner.Text()
		if elem == "" {
			continue
		}

		fields := strings.Split(elem, "#")
		if len(fields) != 2 {
			continue
		}

		parent, err := strconv.Atoi(fields[0])
		if err != nil {
			println("Could not convert %d to int", fields[0])
		}
		spawns[parent] = make([]int, 0)

		children := strings.Split(fields[1], ",")
		for _, child := range children {
			childInt, err := strconv.Atoi(child)
			if err != nil {
				println("Could not convert %d to int", fields[0])
			}
			spawns[parent] = append(spawns[parent], childInt)
		}
	}

	runtime.SetSpawns(spawns)
}

// Import the trace.
// The function creates the replay data structure, that is used to replay the trace.
// We only store the information that is needed to replay the trace.
// This includes operations on
//   - spawn
//   - channels
//   - mutexes
//   - once
//   - waitgroups
//   - select
//
// We only record the relevant information for each operation.
// func readTraceFile(fileName string,
// 	replayData *runtime.AdvocateReplayTrace, spawns *map[int][]int, selects *map[string][]runtime.ReplayElement) {
// 	// get the routine id from the file name
// 	routineID, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(fileName, tracePathRewritten+"/trace_"), ".log"))
// 	if err != nil {GetReplayA
// 		panic(err)
// 	}

// 	file, err := os.Open(fileName)
// 	if err != nil {
// 		panic(err)
// 	}

// 	scanner := bufio.NewScanner(file)

// 	for scanner.Scan() {
// 		elem := scanner.Text()
// 		if elem == "" {
// 			continue
// 		}

// 		fields := strings.Split(elem, ",")

// 		tPre, err := strconv.Atoi(fields[1])
// 		if err != nil {
// 			panic("Invalid tPre " + fields[0] + " in line " + elem + " in file " + fileName + ".")
// 		}

// 		var time int
// 		var op runtime.Operation
// 		var file string
// 		var line int
// 		var blocked = false
// 		var suc = true
// 		var index int

// 		// for some elements, the tPre and tPost are the same
// 		// we here set the time value as the tPre as a default. If the element
// 		// has a separate tPost, this will be set later
// 		time = tPre

// 		switch fields[0] {
// 		case "X": // disable replay
// 			op = runtime.OperationReplayEnd
// 			line, _ = strconv.Atoi(fields[2]) // misuse the line for the exit code
// 			runtime.SetExpectedExitCode(line)
// 		case "G":
// 			op = runtime.OperationSpawn
// 			// time, _ = strconv.Atoi(fields[1])
// 			pos := strings.Split(fields[3], posSep)
// 			file = pos[0]
// 			line, _ = strconv.Atoi(pos[1])
// 			index, _ = strconv.Atoi(fields[2])
// 			if _, ok := (*spawns)[routineID]; !ok {
// 				(*spawns)[routineID] = make([]int, 0)
// 			}
// 			(*spawns)[routineID] = append((*spawns)[routineID], index)
// 			println("ADD SPAWN ", routineID, " ", index)
// 		case "C":
// 			switch fields[4] {
// 			case "S":
// 				op = runtime.OperationChannelSend
// 			case "R":
// 				op = runtime.OperationChannelRecv
// 			case "C":
// 				op = runtime.OperationChannelClose
// 			default:
// 				panic("Unknown channel operation " + fields[4] + " in line " + elem + " in file " + fileName + ".")
// 			}
// 			time, _ = strconv.Atoi(fields[2])
// 			if time == 0 {
// 				blocked = true
// 			}
// 			pos := strings.Split(fields[9], posSep)
// 			file = pos[0]
// 			line, _ = strconv.Atoi(pos[1])
// 		case "M":
// 			rw := false
// 			if fields[4] == "R" {
// 				rw = true
// 			}
// 			time, _ = strconv.Atoi(fields[2])
// 			if fields[6] == "f" {
// 				suc = false
// 			}
// 			pos := strings.Split(fields[7], posSep)
// 			file = pos[0]
// 			line, _ = strconv.Atoi(pos[1])
// 			switch fields[5] {
// 			case "L":
// 				if rw {
// 					op = runtime.OperationRWMutexLock
// 				} else {
// 					op = runtime.OperationMutexLock
// 				}
// 			case "U":
// 				if rw {
// 					op = runtime.OperationRWMutexUnlock
// 				} else {
// 					op = runtime.OperationMutexUnlock
// 				}
// 			case "T":
// 				if rw {
// 					op = runtime.OperationRWMutexTryLock
// 				} else {
// 					op = runtime.OperationMutexTryLock
// 				}
// 			case "R":
// 				op = runtime.OperationRWMutexRLock
// 			case "N":
// 				op = runtime.OperationRWMutexRUnlock
// 			case "Y":
// 				op = runtime.OperationRWMutexTryRLock
// 			default:
// 				panic("Unknown mutex operation")
// 			}
// 			if fields[2] == "0" {
// 				blocked = true
// 			}

// 		case "O":
// 			op = runtime.OperationOnceDo
// 			// time, _ = strconv.Atoi(fields[1]) // read tpre to prevent false order
// 			if time == 0 {
// 				blocked = true
// 			}
// 			if fields[4] == "f" {
// 				suc = false
// 			}
// 			pos := strings.Split(fields[5], posSep)
// 			file = pos[0]
// 			line, _ = strconv.Atoi(pos[1])
// 		case "W":
// 			switch fields[4] {
// 			case "W":
// 				op = runtime.OperationWaitgroupWait
// 			case "A":
// 				op = runtime.OperationWaitgroupAddDone
// 			default:
// 				panic("Unknown waitgroup operation")
// 			}
// 			time, _ = strconv.Atoi(fields[2])
// 			if time == 0 {
// 				blocked = true
// 			}
// 			pos := strings.Split(fields[7], posSep)
// 			file = pos[0]
// 			line, _ = strconv.Atoi(pos[1])
// 		case "S":
// 			cases := strings.Split(fields[4], "~")
// 			if cases[len(cases)-1] == "D" {
// 				op = runtime.OperationSelectDefault
// 			} else {
// 				op = runtime.OperationSelectCase
// 			}
// 			time, _ = strconv.Atoi(fields[2])
// 			if time == 0 {
// 				blocked = true
// 			}
// 			index, _ = strconv.Atoi(fields[5])
// 			pos := strings.Split(fields[6], posSep)
// 			file = pos[0]
// 			line, _ = strconv.Atoi(pos[1])
// 		case "D":
// 			switch fields[4] {
// 			case "W":
// 				op = runtime.OperationCondWait
// 			case "S":
// 				op = runtime.OperationCondSignal
// 			case "B":
// 				op = runtime.OperationCondBroadcast
// 			default:
// 				panic("Unknown cond operation: " + fields[4])
// 			}
// 			pos := strings.Split(fields[5], posSep)
// 			file = pos[0]
// 			line, _ = strconv.Atoi(pos[1])
// 			if fields[2] == "0" {
// 			GetReplayA	blocked = true
// 			}
// 		case "A":
// 			if !runtime.GetReplayAtomic() {
// 				continue
// 			}
// 			switch fields[3] {
// 			case "L":
// 				op = runtime.OperationAtomicLoad
// 			case "S":
// 				op = runtime.OperationAtomicStore
// 			case "A":
// 				op = runtime.OperationAtomicAdd
// 			case "W":
// 				op = runtime.OperationAtomicSwap
// 			case "C":
// 				op = runtime.OperationAtomicCompareAndSwap
// 			case "N":
// 				op = runtime.OperationAtomicAnd
// 			case "O":
// 				op = runtime.OperationAtomicOr
// 			}
// 			pos := strings.Split(fields[4], posSep)
// 			if len(pos) < 2 {
// 				runtime.SetReplayAtomic(false)
// 				file = ""
// 				line = 0
// 			} else {
// 				file = pos[0]
// 				line, _ = strconv.Atoi(pos[1])
// 			}
// 		default:
// 			continue
// 		}
// 		if blocked || time == 0 {
// 			time = math.MaxInt
// 		}

// 		// do not add element to the trace if it is not active
// 		// if activeTime != nil && activeStart != -1 && time > activeStart {
// 		// 	if _, ok := activeTime[time]; !ok {
// 		// 		continue
// 		// 	}
// 		// }

// 		if op != runtime.OperationNone && !runtime.AdvocateIgnoreReplay(op, file) {
// 			newElem := runtime.ReplayElement{
// 				Op: op, Routine: routineID, Time: time, TimePre: tPre, File: file, Line: line,
// 				Blocked: blocked, Suc: suc,
// 				Index: index}
// 			*replayData = append(*replayData, newElem)

// 			if op == runtime.OperationSelect || op == runtime.OperationSelectCase || op == runtime.OperationSelectDefault {
// 				key := runtime.BuildReplayKey(routineID, file, line)
// 				(*selects)[key] = append((*selects)[key], newElem)
// 			}
// 		}
// 	}

// 	if err := scanner.Err(); err != nil {
// 		panic(err)
// 	}
// }

// FinishReplay waits for the replay to finish.
//
//go:linkname FinishReplay runtime.AdvocateFinishReplay
func FinishReplay() {
	if r := recover(); r != nil {
		println("Replay failed.")
	}

	runtime.WaitForReplayFinish()

	// DetectBlockingGC()

	time.Sleep(time.Second)

	runtime.ExitReplayWithCode(runtime.ExitCodeDefault, "")
}
