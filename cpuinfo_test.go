package hostinfo

import (
	_ "embed"
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"testing"

	"gbenson.net/go/invoker"
	"gotest.tools/v3/assert"
)

func TestGatherCPUInfo_live(t *testing.T) {
	r := assertExec(t, gatherCPUInfo)
	assert.Assert(t, len(r.CPUs) > 0)
}

//go:embed resources/cpuinfo.arm
var armCPUInfo []byte

func TestGatherCPUInfo_ARM(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("cat", "/proc/cpuinfo").Returns(armCPUInfo, nil)

	r := assertMock(t, gatherCPUInfo, mock)
	assert.Equal(t, len(r.CPUs), 4)
	assert.Equal(t, len(r.CPUInfo), 11)

	assertNotHasKey(t, r.CPUInfo, "Serial")
	assert.Equal(t, r.CPUInfo["serial"], "100000006b11cc9f")

	assert.DeepEqual(t, r.CPUInfo["features"], []string{
		"asimd",
		"cpuid",
		"crc32",
		"evtstrm",
		"fp",
	})

	// Ensure JSON marshaling preserves empty CPUs.
	b, err := json.Marshal(r)
	assert.NilError(t, err)
	assert.Check(t, strings.Contains(string(b), "\"cpus\":[{},{},{},{}]"))

	// Ensure JSON unmarshaling preserves empty CPUs.
	var u HostInfo
	assert.NilError(t, json.Unmarshal(b, &u))
	assert.Equal(t, len(u.CPUs), 4)
}

//go:embed resources/cpuinfo.x86
var x86CPUInfo []byte

func TestGatherCPUInfo_X86(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("cat", "/proc/cpuinfo").Returns(x86CPUInfo, nil)

	r := assertMock(t, gatherCPUInfo, mock)
	assert.Equal(t, len(r.CPUs), 12)
	assert.Equal(t, len(r.CPUInfo), 22)

	info := r.CPUInfo

	// integer
	assert.Equal(t, info["model"], 154)

	// integer, value converted from hex
	assert.Equal(t, info["microcode"], 0x436)

	// dimensioned integer, key was converted to snake case
	assertNotHasKey(t, info, "cache size")
	assertNotHasKey(t, info, "cache_size")
	assertNotHasKey(t, info, "cache_size_KB")
	assert.Equal(t, info["cache_size_kb"], 12_288)

	// string (non-integer float)
	assert.Equal(t, info["bogomips"], "5222.40")

	// boolean, key was already snake case
	assertNotHasKey(t, info, "fpuException")
	assert.Equal(t, info["fpu_exception"], true)

	wantCPUKeys := []string{
		"apicid",
		"core_id",
		"cpu_mhz",
		"initial_apicid",
	}

	cpu := r.CPUs[7]
	assert.DeepEqual(t, slices.Sorted(maps.Keys(cpu)), wantCPUKeys)
	intMHz, ok := cpu["cpu_mhz"].(int)
	assert.Check(t, ok)
	assert.Equal(t, intMHz, 400) // floaty integer converted to integer

	cpu = r.CPUs[6]
	assert.DeepEqual(t, slices.Sorted(maps.Keys(cpu)), wantCPUKeys)
	floatMHz, ok := cpu["cpu_mhz"].(string)
	assert.Check(t, ok)
	assert.Equal(t, floatMHz, "3443.199") // non-integral float unconverted
}
