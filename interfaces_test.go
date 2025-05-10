package hostinfo

import (
	_ "embed"
	"maps"
	"slices"
	"testing"

	"gbenson.net/go/invoker"
	"gotest.tools/v3/assert"
)

func TestGatherInterfaces_live(t *testing.T) {
	r := assertExec(t, gatherInterfaces)

	var gotIPv4, gotIPv6 bool

	iface := r.Interfaces["lo"]
	assert.Check(t, iface != nil)

	addrs, _ := iface["addr_info"].([]any)
	assert.Check(t, addrs != nil)

	for _, untyped := range addrs {
		addr, _ := untyped.(map[string]any)
		assert.Check(t, addr != nil)

		family, ok := addr["family"].(string)
		assert.Check(t, ok)
		switch family {
		case "inet":
			local, _ := addr["local"].(string)
			assert.Equal(t, local, "127.0.0.1")
			gotIPv4 = true
		case "inet6":
			local, _ := addr["local"].(string)
			assert.Equal(t, local, "::1")
			gotIPv6 = true
		}
	}

	assert.Check(t, gotIPv4 || gotIPv6)
}

//go:embed resources/ip-addr-j.out
var testInterfaces []byte

func TestGatherInterfaces_mock(t *testing.T) {
	mock := invoker.NewMock(t)
	mock.ExpectInvoke("ip", "--json", "address", "show").
		Returns(testInterfaces, nil)

	r := assertMock(t, gatherInterfaces, mock)
	ifs := r.Interfaces

	gotNames := slices.Sorted(maps.Keys(ifs))
	wantNames := []string{
		"br-121144d4a5cf",
		"br-adeb0277c970",
		"docker0",
		"eth0",
		"lo",
	}
	assert.DeepEqual(t, gotNames, wantNames)

	iface := r.Interfaces["eth0"]
	assert.Check(t, iface != nil)
	assert.Equal(t, iface["link_type"], "ether") // string
	assert.Equal(t, iface["mtu"], float64(1500)) // "integer"

	addrs, _ := iface["addr_info"].([]any)
	assert.Equal(t, len(addrs), 2)

	ipv4, _ := addrs[0].(map[string]any)
	assert.Equal(t, ipv4["family"], "inet")
	assert.Equal(t, ipv4["local"], "100.115.92.201")

	ipv6, _ := addrs[1].(map[string]any)
	assert.Equal(t, ipv6["family"], "inet6")
	assert.Equal(t, ipv6["local"], "fe80::216:3eff:feba:ab67")
}
