package hostinfo

import (
	"strings"

	"github.com/acobaugh/osrelease"
)

// gatherOSRelease gathers the content of `/etc/os-release`.
func gatherOSRelease(gi *gatherInvoker, r *HostInfo) error {
	s, err := gi.ReadFile("/etc/os-release")
	if err != nil {
		return err
	}

	m, err := osrelease.ReadString(s)
	if err != nil || len(m) == 0 {
		return err
	}

	r.OS = make(map[string]string)
	for k, v := range m {
		r.OS[strings.ToLower(k)] = v
	}

	return nil
}
