// Copyright (c) 2025 
//
// File: falsePos.go
// Brief: Check for false positives
//
//
// License: BSD-3-Clause

package benign

import (
	"gosc/utils/log"
	"gosc/utils/results/buginfo"
	"fmt"
)

// IsBenign checks if the given bug is likely a false positive based on
// the program code and gc based leak detection
func IsBenign(resultType buginfo.ResultType, fileName string, line int, blocked map[string]map[int]struct{}, contextCancel map[int]struct{}, contextDone map[string]map[int]int) (bool, error) {
	// is confirmed dead by GC
	if _, ok := blocked[fileName][line]; ok {
		return false, nil
	}

	fset, file, info, err := parseFile(fileName)
	if err != nil {
		return false, err
	}

	parents := buildParentMap(file)

	node := findNodeAtLine(fset, file, line)
	if node == nil {
		return false, fmt.Errorf("Could not find node: %s:%d", fileName, line)
	}

	endlessLoop := isInsideEndlessLoop(node, parents)

	if endlessLoop {
		return true, nil
	}

	rangeOverChan, err := isRangeOverChan(node, parents, info)
	if err != nil {
		log.Errorf("Could not check for endless loop")
	}
	if rangeOverChan {
		return true, nil
	}

	contextWithCancel := isContextDoneWithCancel(node, info, fileName, line, contextCancel, contextDone)
	if contextWithCancel {
		return true, nil
	}

	return false, nil
}
