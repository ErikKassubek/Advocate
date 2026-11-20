// Copyright (c) 2024 Erik Kassubek
//
// File: explanation.go
// Brief: Create an explanation file for a found bug
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package explanation

import (
	"advocate/utils/consts"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/results/buginfo"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type bug struct {
	bugType buginfo.ResultType
	pos     string
}

type bugObj struct {
	pos   string
	routs []int
	t     string
	code  string
}

type bugKeys int

const (
	crit bugKeys = iota
	bugType
	name
	explanation
	file
	importLine
	trace
	replaySuc
	exitCodeDesc
	exitCode
	desc
	class
	arg1
	arg2
)

var printedExplanations = make(map[bug]struct{})

func writeBug(bugType buginfo.ResultType, obj map[int][]bugObj) bool {
	posSet := make(map[string]struct{})
	for _, p := range obj {
		for _, l := range p {
			posSet[l.pos] = struct{}{}
		}
	}

	unique := make([]string, 0, len(posSet))
	for s := range posSet {
		unique = append(unique, s)
	}
	sort.Strings(unique)

	// Join into single string
	pos := strings.Join(unique, ",")

	b := bug{
		bugType: bugType,
		pos:     pos,
	}

	if _, ok := printedExplanations[b]; ok {
		return false
	}
	printedExplanations[b] = struct{}{}
	return true
}

// create an overview over an analyzed, and if possible replayed
// bug. It is mostly meant to give an explanation of a found
// bug to people, who are not used to the internal structure an
// representation of the analyzer.

// It creates one file. This file has the following element:
// - The type of bug found
// - The test/program, where the bug was found
// - if possible, the command to run the program
// - if possible, the command to replay the bug
// - position of the bug elements
// - code of the bug elements in the trace (+- 10 lines)
// - info about replay (was it possible or not)

// CreateOverview creates an overview over a bug found by the analyzer.
// It reads the results of the analysis, the code of the bug elements and the replay info.
// It then writes all this information into a file.
//
// Parameter:
//   - index: the index of the bug in the results
//   - traceID: id of the trace
//   - ignoreDouble: if true, only write one bug report for each bug
//
// Returns:
//   - int: number of results
//   - error: if an error occurred
func CreateOverview(ignoreDouble bool, traceID, fuzzing int) (int, error) {
	// get the code info (main file, test name, commands)

	// remove possible partner information if not searched for
	buildBugCodes()

	_, blockedByReplay := getOutputInfo()

	progInfo, err := readProgInfo()
	if err != nil {
		log.Error("Error reading prog info: ", err)
	}
	progInfo[trace] = fmt.Sprintf("advocateTrace_%d", traceID)

	resultsMachine, err := filepath.Glob(filepath.Join(paths.ResultTraces, "results_machine_*.log"))
	if err != nil {
		log.Error(err.Error())
	}

	resultsMachine = append(resultsMachine, filepath.Join(paths.ResultOut, paths.NameResultMachine))

	var numberResults int
	for _, result := range resultsMachine {
		f, err := os.ReadFile(result)
		if err != nil {
			log.Error(err)
		}

		numberResults = len(strings.Split(string(f), "\n")) - 1

		// timeoutFound := false

		for index := 1; index <= numberResults; index++ {
			id := ""
			if strings.HasSuffix(result, paths.NameResultMachine) {
				id += strconv.Itoa(index)
			} else {
				elem := strings.Split(strings.Split(result, ".log")[0], "_")
				id += elem[len(elem)-1] + "_" + strconv.Itoa(index)
			}

			bugType, bugObj, falsePositive, err := readAnalysisResults(result, index, progInfo[file])
			if err != nil {
				log.Error("Could not read analysis result: ", err.Error())
				continue
			}

			if bugType == buginfo.Empty {
				break
			}

			// get the bug type description
			if bugType.IsLeak() && blockedByReplay {
				bugType = buginfo.RReplayFail
			}
			bugTypeDescription := getBugTypeDescription(bugType)

			// get the code of the bug elements
			err = getBugPositions(bugObj)
			if err != nil {
				log.Error("Error getting bug positions: ", err)
			}

			// get the replay info
			// replay := getRewriteInfo(bugType, replayCodes, id)

			// if ignoreDouble && replay[exitCode] == "double" {
			// 	continue
			// }

			if !writeBug(bugType, bugObj) {
				continue
			}

			err = writeFile(paths.CurrentResult, id, bugTypeDescription, bugObj, progInfo, fuzzing, falsePositive)
		}
	}

	return numberResults, err

}

// Read an result machine file file and get one result
//
// Parameter:
//   - path string: path to the result file
//   - index int: index of the relevant bug in the file
//   - fileWithHeader string: file that contains the header for the recording
//   - headerLine int: line number of the first line of the header
//
// Returns:
//   - buginfo.ResultType: bug type
//   - map[int][]string: bug elements
//   - bool: true if false positive
//   - error
func readAnalysisResults(path string, index int, fileWithHeader string) (buginfo.ResultType, map[int][]bugObj, bool, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", nil, false, err
	}

	lines := strings.Split(string(file), "\n")

	index-- // the index is 1-based

	if index >= len(lines) {
		return "", nil, false, errors.New("index out of range")
	}

	bugStr := string(lines[index])
	bugFields := strings.Split(bugStr, ",")
	bugType := buginfo.ResultTypeFromString(bugFields[0])
	falsePositive := (bugFields[1] == "fp")

	obj := make(map[int][]bugObj)

	posAlreadyKnown := make(map[string]int, 0)

	for i := 2; i < len(bugFields); i++ {
		bugElems := strings.Split(bugFields[i], ";")
		if len(bugElems) == 0 {
			continue
		}

		obj[i] = make([]bugObj, 0)

		for _, elem := range bugElems {
			fields := strings.Split(elem, consts.PosSep)

			if fields[0] != "T" {
				continue
			}

			rout, err := strconv.Atoi(fields[1])
			if err != nil {
				log.Error(err)
				continue
			}

			file := fields[5]
			line := fields[6]

			pos := file + consts.PosSep + line

			if o, ok := posAlreadyKnown[pos]; ok {
				obj[i][o].routs = append(obj[i][o].routs, rout)
				continue
			}
			posAlreadyKnown[pos] = len(obj[i])

			o := bugObj{
				pos:   pos,
				routs: []int{rout},
				t:     getBugElementType(fields[4]),
			}

			obj[i] = append(obj[i], o)
		}
	}

	return bugType, obj, falsePositive, nil
}

// writeFile(path, id, bugTypeDescription, bugPos, bugElemType, code,
// 	replay, progInfo, fuzzing)

// Write an bug explanation file
//
// Parameter:
//   - path string: path where the explanation file should be created
//   - index int: index of the bug file (name is e.g. bug_actual_[index])
//   - description map[string]string: description of the bug, containing e.g. bug/diagnostics, name, explanation
//   - obj map[int][]bugObj: bug objects
//   - progInfo map[string]sting: Info about the prog, e.g. prog/test name
//   - fuzzing int: Fuzzing run number
//   - falsePositive bool: true if it is likely a false positive
//
// Returns:
//   - error
func writeFile(path string, index string, description map[bugKeys]string,
	obj map[int][]bugObj, progInfo map[bugKeys]string, fuzzing int, falsePositive bool) error {

	// write the bug type description
	res := "# " + description[crit] + ": " + description[bugType] + " - " + description[name] + "\n\n"
	res += description[explanation] + "\n\n"

	// write the positions of the bug
	res += "## Test/Program\n"
	res += "The bug was found in the following test/program:\n\n"
	if progInfo[name] != "" {
		res += "- Test/Prog: " + progInfo[name] + "\n"
	} else {
		res += "- Test: unknown" + "\n"
	}

	if progInfo[file] != "" {
		res += "- File: " + progInfo[file] + "\n"
	} else {
		res += "- File: unknown" + "\n"
	}

	if progInfo[trace] != "" {
		res += "- Trace: " + progInfo[trace] + "\n\n"
	} else {
		res += "- Trace: unknown" + "\n\n"
	}

	if falsePositive {
		res += "The bug is likely a false positive\n\n"
	}

	// write the code of the bug elements
	if len(obj) > 0 {
		res += "## Bug Elements\n\n"
		res += "The elements involved in the found "
		res += strings.ToLower(description[crit])
		res += " are located at the following positions:\n\n"

		for key := range obj {
			if len(obj[key]) == 0 {
				continue
			}

			for _, o := range obj[key] {

				res += "###  "

				closeBracket := false
				switch key {
				case 2:
					res += description[arg1] + " ("
					closeBracket = true
				case 3:
					res += description[arg2] + " ("
					closeBracket = true
				}

				res += o.t

				if len(o.routs) > 1 {
					res += fmt.Sprintf(", %dx", len(o.routs))
				}

				if closeBracket {
					res += ")"
				}

				res += "\n\n"

				if o.pos == ":-1" {
					return nil
				}
				code := o.code

				routsName := "Routines"
				if len(o.routs) < 2 {
					routsName = "Routine"
				}
				routs := ""
				slices.Sort(o.routs)
				for i, r := range o.routs {
					if i != 0 {
						routs += ", "
					}
					routs += fmt.Sprintf("%d", r)
				}

				res += fmt.Sprintf("-> %s\n-> %s: %s\n\n", o.pos, routsName, routs)
				res += code + "\n\n"
			}
		}
	}

	// Replay now in own file

	// confirmed := false

	// if description[class] == consts.Possible { // replay only for possible bugs
	// 	// write the info about the replay, if possible including the command to read the bug
	// 	replayPossible := replay[replaySuc] != "was not possible" && replay[replaySuc] != "was not run"
	// 	replayDouble := replay[exitCode] == "double"

	// 	res += "## Replay\n"
	// 	if replayPossible && !replayDouble {
	// 		res += replay[desc] + "\n\n"
	// 	}

	// 	if replayDouble {
	// 		res += "The replay was not performed, because the same bug had been found before."
	// 	} else {
	// 		res += "**Replaying " + replay[replaySuc] + "**.\n\n"
	// 		if replayPossible {
	// 			// res += "The replayed trace can be found in: "
	// 			// res += paths.NameRewrittenConsts + "_" + index + "\n\n"
	// 			if replay[replaySuc] == "panicked" {
	// 				res += "It panicked with the following message:\n\n"
	// 				res += replay[replaySuc] + "\n\n"
	// 			} else if replay[exitCode] == "fail" {
	// 				res += replay[exitCodeDesc] + "\n\n"
	// 			} else {
	// 				res += consts.ItExitedWithTheFollowingCode
	// 				res += replay[exitCode] + "\n\n"
	// 				res += replay[exitCodeDesc] + "\n\n"
	// 			}
	// 		}
	// 	}

	// 	if description[crit] == "Bug" {
	// 		if replayDouble || replay[replaySuc] == "confirmed the bug" ||
	// 			strings.HasPrefix(description[name], consts.Actual) {
	// 			confirmed = true
	// 		}
	// 	}

	// }

	confirmed := !strings.HasPrefix(description[bugType], "P")
	id := progInfo[file] + "#" + progInfo[name]
	log.Resultf(true, confirmed, id, "Found %s.", description[name])

	// dep := strings.TrimPrefix(description[name], consts.Possible)
	// if replay[replaySuc] == "was not run" {
	// 	log.Resultf(true, confirmed, id, "Found %s.", description[name])
	// } else if replay[replaySuc] == "confirmed the bug" {
	// 	log.Resultf(true, confirmed, id, "Found %s.", dep)
	// } else if !confirmed {
	// 	return nil
	// }

	// if in path, the folder "bugs" does not exist, create it
	if _, err := os.Stat(path + consts.Sep + "bugs"); os.IsNotExist(err) {
		err := os.Mkdir(path+consts.Sep+"bugs", 0755)
		if err != nil {
			return err
		}
	}

	folderName := filepath.Join(path, "bugs")
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		err := os.Mkdir(folderName, 0755)
		if err != nil {
			return err
		}
	}

	// create the file
	fileName := ""
	if fuzzing == -1 {
		fileName = filepath.Join(folderName, strings.ToLower(description[crit])) + "_" + index + ".md"
	} else {
		fileName = filepath.Join(folderName, fmt.Sprintf("bug_%s_%d_%s.md", strings.ToLower(description[crit]), fuzzing, index))
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(res)
	return err

}
