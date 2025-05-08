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

//go:embed resources/luksdump
var luksdump []byte

func TestGatherDiskAttrs(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("/sbin/blkid").
		Returns(blkidOutput, nil)
	mock.ExpectInvoke("cryptsetup", "luksDump", "/dev/nvme0n1p3").
		Returns(luksdump, nil)

	r := assertMock(t, gatherDiskAttrs, mock)
	devices := r.Disks

	deviceNames := slices.Sorted(maps.Keys(devices))
	wantDeviceNames := []string{
		"/dev/mapper/nvme0n1p3_crypt",
		"/dev/mapper/vgubuntu-root",
		"/dev/mapper/vgubuntu-swap_1",
		"/dev/nvme0n1p1",
		"/dev/nvme0n1p2",
		"/dev/nvme0n1p3",
	}
	assert.DeepEqual(t, deviceNames, wantDeviceNames)

	assert.Equal(t, devices["/dev/nvme0n1p1"]["uuid"], "8B92-BD41") // string
	assert.Equal(t, devices["/dev/nvme0n1p2"]["block_size"], 4096)  // integer

	const luksDevice = "/dev/nvme0n1p3"
	for name, device := range devices {
		if name == luksDevice {
			continue
		}
		assertNotHasKey(t, device, "luks")
	}

	luks, ok := devices[luksDevice]["luks"].(map[string]any)
	assert.Assert(t, ok)
	keys := slices.Sorted(maps.Keys(luks))
	assert.DeepEqual(t, keys, []string{
		"data_segments",
		"digests",
		"epoch",
		"keyslots",
		"keyslots_area_bytes",
		"metadata_area_bytes",
		"tokens",
		"uuid",
		"version",
	})

	// integer
	assert.Equal(t, luks["version"], 2)

	// string
	assert.Equal(t, luks["uuid"], "242e637a-461b-d087-c66e-384d35525691")

	// dimensioned integer
	assert.Equal(t, luks["metadata_area_bytes"], 16384)

	// slice with omission
	dataSegments, ok := luks["data_segments"].([]map[string]any)
	assert.Assert(t, ok)
	assert.Equal(t, len(dataSegments), 1)
	ds := dataSegments[0]
	assert.Equal(t, ds["cipher"], "aes-xts-plain64")
	assert.Equal(t, ds["offset_bytes"], 1<<24)
	assertNotHasKey(t, ds, "length")

	// slice with hex-encoded data
	keySlots, ok := luks["keyslots"].([]map[string]any)
	assert.Assert(t, ok)
	assert.Equal(t, len(keySlots), 1)
	ks := keySlots[0]
	assert.Equal(t, ks["type"], "luks2")            // string added by slice parser
	assert.Equal(t, ks["priority"], "normal")       // string added by map parser
	assert.Equal(t, ks["time_cost"], 7)             // dimensionless integer
	assert.Equal(t, ks["key_bits"], 512)            // dimensioned integer
	assert.Equal(t, ks["area_offset_bytes"], 32768) // dimensioned integer
	assert.Equal(t, ks["digest_id"], 0)             // (checking on strcase)
	assert.Equal(t, ks["salt"], "e9 44 e4 64 "+     // salt
		"39 38 41 52 8b ca 8a d4 61 5d 2a 37 b6 01 "+
		"c4 76 52 ac e4 5c 7d 77 f3 9b 9f 25 2a de")
}
