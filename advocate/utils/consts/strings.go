package consts

import (
	"path/filepath"
)

const (
	ItExitedWithTheFollowingCode                   = "It exited with the following code: "
	TheBugIsLikelyAFalsePositive                   = "The bug is likely a false positive"
	TheAnalyzerHasTriedToRewriteTheTraceInSuchAWay = "The analyzer has tried to rewrite the trace in such a way"

	ConfirmedTheBug = "confirmed the bug"

	PosSep = "#"

	Sep = string(filepath.Separator)
)
