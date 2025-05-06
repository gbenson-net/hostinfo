package hostinfo

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGatherMachineID(t *testing.T) {
	r := assertExec(t, gatherMachineID)
	assert.Check(t, len(r.MachineID) == 32)
}
