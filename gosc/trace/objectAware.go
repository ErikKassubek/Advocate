// Copyright (c) 2024 
//
// File: /gosc/trace/objectAware.go
// Brief: Object awareness for blocking bug detection
//
//
// License: BSD-3-Clause

package trace

import (
	"strconv"
	"strings"
)

func (this *Trace) AddTraceObjectAware(routine int, objects string) error {
	obj := strings.Split(objects, "-")
	if len(obj) == 0 {
		return nil
	}

	for _, o := range obj {
		n, err := strconv.Atoi(o)
		if err != nil {
			return err
		}

		this.routines[routine].addResource(NewResource(n, nil))
	}

	return nil
}
