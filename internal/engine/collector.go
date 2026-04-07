package engine

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VerifyResult holds the parsed output of a verify run.
type VerifyResult struct {
	L1Build    bool
	L2Passed   int
	L2Total    int
	L3Issues   int
	L4Passed   int
	L4Total    int
	RawOutput  string
}

// benchResultRe matches "BENCH_RESULT: L1=... L2=... L3=... L4=..."
var benchResultRe = regexp.MustCompile(
	`BENCH_RESULT:\s*L1=(\w+)\s+L2=(\d+)/(\d+)\s+L3=(\d+)\s+L4=(\d+)/(\d+)`,
)

// ParseVerifyOutput extracts L1-L4 results from verify script output.
// The expected format is: BENCH_RESULT: L1=PASS L2=8/8 L3=0 L4=5/5
//
// @implements REQ-COLLECTOR (parse verify output into structured L1-L4 results)
func ParseVerifyOutput(output string) (*VerifyResult, error) {
	result := &VerifyResult{RawOutput: output}

	// Check for L1=FAIL first (build failure).
	if strings.Contains(output, "BENCH_RESULT: L1=FAIL") {
		result.L1Build = false
		return result, nil
	}

	matches := benchResultRe.FindStringSubmatch(output)
	if matches == nil {
		return nil, fmt.Errorf("no BENCH_RESULT line found in verify output")
	}

	result.L1Build = matches[1] == "PASS"
	result.L2Passed, _ = strconv.Atoi(matches[2])
	result.L2Total, _ = strconv.Atoi(matches[3])
	result.L3Issues, _ = strconv.Atoi(matches[4])
	result.L4Passed, _ = strconv.Atoi(matches[5])
	result.L4Total, _ = strconv.Atoi(matches[6])

	return result, nil
}
