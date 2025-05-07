package hostinfo

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGatherOSRelease(t *testing.T) {
	r := assertExec(t, gatherOSRelease)
	assert.Assert(t, len(r.OS) > 0)

	switch r.OS["id"] {
	case "debian":
		assert.Check(t, r.OS["name"] == "Debian GNU/Linux")
		assertNotHasKey(t, r.OS, "id_like")
	case "ubuntu":
		assert.Check(t, r.OS["name"] == "Ubuntu")
		assert.Check(t, r.OS["id_like"] == "debian")
	default:
		t.Errorf("unhandled: %q", r.OS["name"])
	}
}
