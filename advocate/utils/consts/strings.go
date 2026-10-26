package consts

import (
	"path/filepath"
)

const (
	ItExitedWithTheFollowingCode                   = "It exited with the following code: "
	TheBugIsLikelyAFalsePositive                   = "The bug is likely a false positive"
	TheAnalyzerHasTriedToRewriteTheTraceInSuchAWay = "The analyzer has tried to rewrite the trace in such a way"

	Bug        = "Bug"
	Potential  = "Potential"
	Leak       = "Leak"
	Diagnostic = "Diagnostic"

	Actual    = "Actual"
	Possible  = "Possible"
	Confirmed = "Confirmed"
	Runtime   = "Runtime"

	ModeRecord   = "record"
	ModeReplay   = "replay"
	ModeFuzzing  = "fuzzing"
	ModeAnalysis = "analysis"
	ModeGui      = "gui"

	ConfirmedTheBug = "confirmed the bug"

	RewrittenConsts = "rewrittenConstraints"

	PosSep = "#"

	Sep = string(filepath.Separator)
)
