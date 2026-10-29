// Copyright (c) 2024 Erik Kassubek
//
// File: io.go
// Brief: Functions to read ans write the fuzzing files
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_base

import (
	"advocate/fuzzing/f_constraints"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/io"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/settings"
	"advocate/utils/timer"
	"fmt"
	"os"
	"path/filepath"
)

// WriteMutConstraint writes a constratin based mutation the mutation to file and add it to the queue
//
// Parameter:
//   - tr *trace.Trace: the trace the constraint is build on
//   - mut Constratin: the constratin to write
//   - first bool: set to true, if it is the first mutation for a given test/prog
//
// Returns:
//   - bool: true if max fuzzing number is reached
//   - error
func WriteMutConstraint(tr *trace.Trace, mut *f_constraints.Constraint, first bool) (bool, error) {
	if MaxNumberRuns != -1 && NumberWrittenMutations > MaxNumberRuns {
		return true, nil
	}
	NumberWrittenMutations++

	if first {
		AddFuzzingTraceFolder(paths.FuzzingTraces)
	}

	fuzzingTracePath := filepath.Join(paths.FuzzingTraces, fmt.Sprintf("fuzzingTrace_%d", NumberWrittenMutations))
	ChainFiles[NumberWrittenMutations] = *mut

	err := io.CreateTraceFolder(fuzzingTracePath)
	if err != nil {
		return false, err
	}

	closure := f_constraints.CalcuateStableClosure(mut)

	// write the active map to a "replay_constratins.log"
	if flags.FuzzingMode == GoPie || settings.WithoutReplay {
		WriteConstraint(fuzzingTracePath, tr, mut, closure)
	} else {
		WriteConstraint(fuzzingTracePath, tr, mut, closure)
	}

	muta := Mutation{MutType: MutPiType, MutPie: NumberWrittenMutations}

	AddMutToQueue(muta, false)

	return false, nil
}

// Create the folder for the fuzzing traces
//
// Parameter:
//   - path string: path to the folder
func AddFuzzingTraceFolder(path string) {
	os.RemoveAll(path)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Error("Could not create fuzzing folder")
	}
}

// WriteMutationToFile writes a given mutation to a mutation file.
// These files are used to run the mutation
//
// Parameter:
//   - pathToFolder string: path to where the mutation should be created
//   - mut mutation: the mutation to write
//
// Returns:
//   - error
func WriteMutationToFile(pathToFolder string, mut Mutation) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	// write for mut and mut type, for goPie it is already written
	if mut.MutType == MutSelType || mut.MutType == MutFlowType {
		fileName := filepath.Join(pathToFolder, paths.NameFuzzingData)
		sep := "@"

		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close() // Ensure the file is closed when we're done

		// Write the string to the file
		for pos, selects := range mut.MutSel {
			content := fmt.Sprintf("%s;", pos)

			for i, sel := range selects {
				if i != 0 {
					content += ","
				}
				content += fmt.Sprintf("%d", sel.ChosenCase)
			}

			content += "\n"

			_, err = file.WriteString(content)
			if err != nil {
				return err
			}
		}

		_, err = file.WriteString(sep + "\n")

		for pos, count := range mut.MutFlow {
			content := fmt.Sprintf("%s;%d\n", pos, count)
			_, err = file.WriteString(content)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetPath returns a path for a path
// Given a path, if it a dir, return the path, otherwise return the
// path to the dir the file is in
//
// Parameter:
//   - path string: path
//
// Returns:
//   - string: path to the dir
func GetPath(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return path
	}

	if info.IsDir() {
		return path
	}

	return filepath.Dir(path)
}

// WriteConstraints writes the element in the chain into a rewriteActive.log
// file for use in GoPie and Guided
//
// Parameter
//   - fuzzingTracePath string: path to the trace folder
//   - mut *chain: chain to write
//   - mutLen int: number of elements in mutation (without closure)
//   - closureLen int: number of elements in closure
func writeConstraintOrder(fuzzingTracePath string, tr *trace.Trace, mut *f_constraints.Constraint, mutLen, closureLen int) error {
	constraintPath := filepath.Join(fuzzingTracePath, paths.NameReplayConstraints)

	f, err := os.Create(constraintPath)
	if err != nil {
		log.Error("Failed to create constraint file: ", err.Error())
		return err
	}

	defer f.Close()

	f.WriteString(fmt.Sprintf("%d:%d\n", mutLen, closureLen))

	// find the counter for all elements in the mut
	mutCounter := make(map[trace.Element]int)
	posCounter := make(map[string]int)
	for _, elem := range mut.Elems {
		mutCounter[elem] = 0
	}

	traceIter := tr.AsIterator()

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		// if alloc is not recorded, it has the same position as the first operation
		// At the moment we do not use alloc in replay, we can therefore just skip it
		if _, ok := elem.(*trace.ElementAlloc); ok {
			continue
		}
		routPos := getRoutPos(elem)
		posCounter[routPos]++
		if _, ok := mutCounter[elem]; ok { // is in chain
			mutCounter[elem] = posCounter[routPos]
		}
	}

	for _, elem := range mut.Elems {
		info := "0"
		switch e := elem.(type) {
		case *trace.ElementSelect:
			info = fmt.Sprintf("%d", e.GetChosenIndex())
		}

		if elem.Pos().File() == "" {
			continue
		}

		// key := fmt.Sprintf("%d:%s,%d,%d\n", elem.GetRoutine(), elem.GetPos(), mutTPre[traceID], mutCounter[traceID])
		key := fmt.Sprintf("%d:%s:%d:%s\n", elem.RoutineID(), elem.Pos(), mutCounter[elem], info)
		f.WriteString(key)
	}

	return nil
}

func writeForkMap(fuzzingTracePath string, tr *trace.Trace) error {
	tree := tr.ForkTree()

	activePath := filepath.Join(fuzzingTracePath, paths.NameReplayForkTree)

	f, err := os.Create(activePath)
	if err != nil {
		log.Error("Failed to create fuzzing trace folder: ", err.Error())
		return err
	}

	defer f.Close()

	for old, new := range tree {
		newStr := ""
		for i, n := range new {
			if i != 0 {
				newStr += ","
			}
			newStr += fmt.Sprintf("%d", n)
		}
		key := fmt.Sprintf("%d#%s\n", old, newStr)
		f.WriteString(key)
	}

	return nil
}

func WriteConstraint(tracePath string, tr *trace.Trace, constr, closure *f_constraints.Constraint) error {
	mutLen := constr.Len()
	closureLen := 0
	var fullConstr *f_constraints.Constraint

	if closure != nil {
		closure.Len()
		fullConstr = f_constraints.MergeClosureConstraint(closure, constr)

	}

	err := os.MkdirAll(tracePath, 0755)
	if err != nil {
		log.Error("Failed to create constraint folder: ", err.Error())
		return err
	}

	err = writeConstraintOrder(tracePath, tr, fullConstr, mutLen, closureLen)
	if err != nil {
		return err
	}
	err = writeForkMap(tracePath, tr)
	return err
}

func getRoutPos(elem trace.Element) string {
	return fmt.Sprintf("%d:%s", elem.RoutineID(), elem.Pos())
}
