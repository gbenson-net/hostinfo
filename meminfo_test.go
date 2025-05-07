package hostinfo

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"gbenson.net/go/invoker"
	"gotest.tools/v3/assert"
)

func TestGatherMemInfo_live(t *testing.T) {
	r := assertExec(t, gatherMemInfo)
	assertMemInfo(t, r.Memory, nil)
}

//go:embed resources/meminfo
var testMemInfo []byte

func TestGatherMemInfo_mock(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("cat", "/proc/meminfo").Returns(testMemInfo, nil)
	r := assertMock(t, gatherMemInfo, mock)
	got := r.Memory

	assert.Equal(t, len(got), 55)
	assertMemInfo(t, got, map[string]any{
		"mem_total_kb":  40_743_392,
		"swap_total_kb": 2_002_940,
	})

	assertNotHasKey(t, got, "HugePages_Total")
	assertNotHasKey(t, got, "hugepages_total")
	assertNotHasKey(t, got, "HugePages_Total_kb")
	assertNotHasKey(t, got, "hugepages_total_kb")
	assertNotHasKey(t, got, "huge_pages_total_kb")
	assert.Equal(t, got["huge_pages_total"], 5)
}

var snakeCaseRx = regexp.MustCompile(`^[0-9a-z]+(?:_[0-9a-z]+)*$`)

func TestGatherMemInfo_strcase_fails(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("cat", "/proc/meminfo").Returns(testMemInfo, nil)
	r := assertMock(t, gatherMemInfo, mock)

	// Detect problems
	for k, _ := range r.Memory {
		if snakeCaseRx.FindString(k) == k {
			continue
		}
		t.Error("malformed key", k)
	}

	// Validate fixes
	for k, want := range map[string]int{
		"active_anon_kb":   2_091_416,
		"active_file_kb":   1_292_544,
		"inactive_anon_kb": 0,
		"inactive_file_kb": 1_474_820,
	} {
		got, found := r.Memory[k]
		if !found {
			t.Error("missing:", k)
			continue
		}
		assert.Equal(t, got, want, k)
	}
}

func assertMemInfo(t *testing.T, got, want map[string]any) {
	t.Helper()

	assertMemInfoTotal(t, got, want, "Mem")
	assertMemInfoTotal(t, got, want, "Swap")
}

func assertMemInfoTotal(t *testing.T, got, want map[string]any, titleKey string) {
	t.Helper()

	lowerKey := strings.ToLower(titleKey)
	assert.Assert(t, lowerKey != titleKey)

	assertNotHasKey(t, got, titleKey+"Total")  // MemTotal, SwapTotal
	assertNotHasKey(t, got, lowerKey+"_total") // mem_total, swap_total

	key := lowerKey + "_total_kb"

	gotValue, ok := got[key].(int)
	assert.Assert(t, ok, key)
	assert.Check(t, gotValue >= 0, key)

	wantValue, found := want[key]
	if !found {
		return
	}

	assert.Equal(t, gotValue, wantValue, fmt.Sprintf("%s: %v", key, gotValue))
}
