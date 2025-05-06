package hostinfo

import (
	"regexp"
	"strings"
)

var machineIDrx = regexp.MustCompile(`^[0-9a-f]{32}$`)

// gatherMachineID gathers the content of `/etc/machine-id`.
//
// "The `/etc/machine-id` file contains the unique machine ID of the
// local system that is set during installation or boot. The machine
// ID is a single newline-terminated, hexadecimal, 32-character,
// lowercase ID. When decoded from hexadecimal, this corresponds to
// a 16-byte/128-bit value. This ID may not be all zeros." [1]
//
// [1]: https://www.man7.org/linux/man-pages/man5/machine-id.5.html
func gatherMachineID(gi *gatherInvoker, r *HostInfo) error {
	s, err := gi.ReadFile("/etc/machine-id")
	if err != nil {
		return err
	}

	r.MachineID = machineIDrx.FindString(strings.TrimSpace(s))
	return nil
}
