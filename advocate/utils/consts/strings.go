package consts

import (
	"path/filepath"
)

const (
	ItExitedWithTheFollowingCode                   = "It exited with the following code: "
	TheBugIsLikelyAFalsePositive                   = "The bug is likely a false positive"
	TheAnalyzerHasTriedToRewriteTheTraceInSuchAWay = "The analyzer has tried to rewrite the trace in such a way"

	ModeRecord     = "record"
	ModeReplay     = "replay"
	ModeFuzzing    = "fuzzing"
	ModeAnalysis   = "analysis"
	ModeGui        = "gui"
	ModeShowTrace  = "show"
	ModeExperiment = "experiment"

	ConfirmedTheBug = "confirmed the bug"

	RewrittenConsts = "rewrittenConstraints"

	PosSep = "#"

	Sep = string(filepath.Separator)
)
