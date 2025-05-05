// Package hostinfo gathers information about a host.
package hostinfo

import (
	"context"
	"regexp"
	"strings"
)

// A HostInfo describes a host and is returned by [Gather].
type HostInfo struct {
	// MachineID is the contents of "/etc/machine-id".
	MachineID string `json:"machine_id,omitempty"`
}

type gatherer struct {
	invoker Invoker
}

// Gather returns a [HostInfo] describing a host.
func Gather(ctx context.Context, invoker Invoker) (*HostInfo, error) {
	result := &HostInfo{}

	g := gatherer{invoker}
	for _, op := range []func(context.Context, *HostInfo) error{
		g.gatherMachineID,
	} {
		if err := op(ctx, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// invoke wraps Invoker.Invoke.
func (g *gatherer) invoke(
	ctx context.Context,
	name string,
	arg ...string,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	out, err := g.invoker.Invoke(name, arg...)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

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
func (g *gatherer) gatherMachineID(ctx context.Context, r *HostInfo) error {
	s, err := g.invoke(ctx, "cat", "/etc/machine-id")
	if err != nil {
		return err
	}

	r.MachineID = machineIDrx.FindString(strings.TrimSpace(s))
	return nil
}
