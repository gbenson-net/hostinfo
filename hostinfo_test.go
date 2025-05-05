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
