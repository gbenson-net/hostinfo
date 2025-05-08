package hostinfo

import (
	_ "embed"
	"testing"

	"gbenson.net/go/invoker"
	"gbenson.net/go/logger"
	"gotest.tools/v3/assert"
)

func TestGatherOSRelease_live(t *testing.T) {
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
		logger.Ctx(testctx(t)).Warn().
			Str("name", r.OS["name"]).
			Str("id", r.OS["id"]).
			Msg("Unhandled OS")
	}
}

func TestGatherOSRelease_fail(t *testing.T) {
	want := &InvalidLineError{"os_test", "x"}
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("cat", "/etc/os-release").Returns(nil, want)

	r, err := invoke(t, mock, gatherOSRelease)
	assert.Equal(t, err, want)
	assert.Equal(t, len(r.OS), 0)
}

func TestGatherOSRelease_empty(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("cat", "/etc/os-release").Returns([]byte{}, nil)

	r := assertMock(t, gatherOSRelease, mock)
	assert.Equal(t, len(r.OS), 0)
}

//go:embed resources/os-release.fedora
var fedoraOSRelease []byte

func TestGatherOSRelease_fedora(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("cat", "/etc/os-release").Returns(fedoraOSRelease, nil)

	r := assertMock(t, gatherOSRelease, mock)
	assert.Equal(t, len(r.OS), 23)

	assert.Equal(t, r.OS["name"], "Fedora Linux")
	assert.Equal(t, r.OS["id"], "fedora")
	assert.Equal(t, r.OS["version"], "42 (Container Image Prerelease)")
	assert.Equal(t, r.OS["version_id"], "42")

	assertNotHasKey(t, r.OS, "id_like")
}

//go:embed resources/os-release.ubuntu
var ubuntuOSRelease []byte

func TestGatherOSRelease_ubuntu(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("cat", "/etc/os-release").Returns(ubuntuOSRelease, nil)

	r := assertMock(t, gatherOSRelease, mock)
	assert.Equal(t, len(r.OS), 12)

	assert.Equal(t, r.OS["name"], "Ubuntu")
	assert.Equal(t, r.OS["id"], "ubuntu")
	assert.Equal(t, r.OS["id_like"], "debian")
	assert.Equal(t, r.OS["version"], "22.04.5 LTS (Jammy Jellyfish)")
	assert.Equal(t, r.OS["version_id"], "22.04")
	assert.Equal(t, r.OS["pretty_name"], "Ubuntu 22.04.5 LTS")
}
