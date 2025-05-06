package hostinfo

import (
	_ "embed"
	"maps"
	"slices"
	"testing"

	"gbenson.net/go/invoker"
	"gotest.tools/v3/assert"
)

//go:embed resources/blkid.out
var blkidOutput []byte

func TestGatherDiskAttrs(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("/sbin/blkid").Returns(blkidOutput, nil)

	r := assertMock(t, gatherDiskAttrs, mock)
	devices := r.Disks

	deviceNames := slices.Sorted(maps.Keys(devices))
	assert.DeepEqual(t, deviceNames, []string{
		"/dev/mapper/nvme0n1p3_crypt",
		"/dev/mapper/vgubuntu-root",
		"/dev/mapper/vgubuntu-swap_1",
		"/dev/nvme0n1p1",
		"/dev/nvme0n1p2",
		"/dev/nvme0n1p3",
	})

	assert.Equal(t, devices["/dev/nvme0n1p1"]["uuid"], "8B92-BD41")
	assert.Equal(t, devices["/dev/nvme0n1p2"]["block_size"], "4096")
}
