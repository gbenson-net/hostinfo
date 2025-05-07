package hostinfo

import (
	_ "embed"
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
	assert.Equal(t, len(r.CPUInfo), 4)

	assertNotHasKey(t, r.CPUInfo, "Serial")
	assert.Equal(t, r.CPUInfo["serial"], "100000006b11cc9f")

	assert.DeepEqual(t, r.CPUs[2]["features"], []string{
		"asimd",
		"cpuid",
		"crc32",
		"evtstrm",
		"fp",
	})
}

//go:embed resources/cpuinfo.x86
var x86CPUInfo []byte

func TestGatherCPUInfo_X86(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("cat", "/proc/cpuinfo").Returns(x86CPUInfo, nil)

	r := assertMock(t, gatherCPUInfo, mock)
	assert.Equal(t, len(r.CPUs), 12)
	assert.Equal(t, len(r.CPUInfo), 0)

	cpu := r.CPUs[7]
	assertNotHasKey(t, r.CPUInfo, "processor")

	// integer
	assert.Equal(t, cpu["model"], 154)

	// integer, value converted from hex
	assert.Equal(t, cpu["microcode"], 0x436)

	// string, key was converted to snake case
	assertNotHasKey(t, cpu, "cache size")
	assert.Equal(t, cpu["cache_size"], "12288 KB")

	// string, key was converted to snake case
	assertNotHasKey(t, cpu, "cpu MHz") // input
	assertNotHasKey(t, cpu, "cpu_MHz")
	assertNotHasKey(t, cpu, "cpu_m_hz") // strcase fail
	assert.Equal(t, cpu["cpu_mhz"], "400.000")

	// boolean, key was already snake case
	assert.Check(t, cpu["fpuException"])
}
