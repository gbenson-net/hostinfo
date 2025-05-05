package hostinfo

import (
	"encoding/json"
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestGatherMachineID(t *testing.T) {
	g := gatherer{&ExecInvoker{}}
	r := HostInfo{}
	assert.NilError(t, g.gatherMachineID(t.Context(), &r))
	assert.Check(t, len(r.MachineID) == 32)
}

func TestGatherOSRelease(t *testing.T) {
	g := gatherer{&ExecInvoker{}}
	r := HostInfo{}
	assert.NilError(t, g.gatherOSRelease(t.Context(), &r))
	assert.Assert(t, len(r.OS) > 0)

	switch r.OS["id"] {
	case "debian":
		assert.Check(t, r.OS["name"] == "Debian GNU/Linux")
		_, found := r.OS["id_like"]
		assert.Check(t, !found)
	case "ubuntu":
		assert.Check(t, r.OS["name"] == "Ubuntu")
		assert.Check(t, r.OS["id_like"] == "debian")
	default:
		t.Errorf("unhandled: %q", r.OS["name"])
	}
}

// TestGather flexes the expected use case.
func TestGather(t *testing.T) {
	ctx := t.Context()
	hostinfo, err := Gather(ctx, &ExecInvoker{})
	assert.NilError(t, err)

	data, err := json.Marshal(hostinfo)
	assert.NilError(t, err)

	const filename = "localhost.json"
	assert.NilError(t, os.WriteFile(filename, data, 0666))
	t.Log("Wrote", filename)
}
