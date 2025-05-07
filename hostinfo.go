// Package hostinfo gathers information about a host.
package hostinfo

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"strings"

	"gbenson.net/go/invoker"
	"gbenson.net/go/logger"
)

// A HostInfo describes a host and is returned by [Gather].  The
// result is intended for serialization with minimal loss and is
// stored in mappings rather than objects by design.  Consumers can
// then either embed/alias the type and add accessor methods, or
// define their own types with the required fields and types.
type HostInfo struct {
	// Disks is constructed from the output of `/sbin/blkid`.
	Disks map[string]map[string]string `json:"block_devices,omitempty"`

	// CPUs and CPUInfo are the contents of "/proc/cpuinfo".
	CPUs    []map[string]any `json:"cpus,omitempty"`
	CPUInfo map[string]any   `json:"cpu_info,omitempty"`

	// MachineID is the contents of "/etc/machine-id".
	MachineID string `json:"machine_id,omitempty"`

	// Memory is the contents of "/proc/meminfo".
	Memory map[string]any `json:"memory,omitempty"`

	// OS is the contents of "/etc/os-release".
	OS map[string]string `json:"operating_system,omitempty"`
}

// Gather returns a [HostInfo] describing a host.
func Gather(ctx context.Context, invoker invoker.Invoker) (*HostInfo, error) {
	result := &HostInfo{}
	success := false

	gi := gatherInvoker{ctx, invoker}
	for _, op := range []gatherer{
		gatherDiskAttrs,
		gatherCPUInfo,
		gatherMachineID,
		gatherMemInfo,
		gatherOSRelease,
	} {
		if err := op(&gi, result); err == nil {
			success = true
		} else {
			logger.Ctx(ctx).Warn().
				Str("item", op.String()).
				AnErr("reason", err).
				Msg("Gather failed")
		}
	}

	if !success {
		return nil, errors.New("all gatherers failed")
	}

	return result, nil
}

// A gatherer is a function that populates part of a HostInfo.
type gatherer func(*gatherInvoker, *HostInfo) error

// String returns the name of a gatherer, for error messages etc.
func (op gatherer) String() string {
	name := runtime.FuncForPC(reflect.ValueOf(op).Pointer()).Name()
	dot := strings.LastIndex(name, ".")
	if dot > 0 {
		name = name[dot+1:]
	}
	name, _ = strings.CutPrefix(name, "gather")
	return name
}

// gatherInvoker modifies Invoker.
type gatherInvoker struct {
	context context.Context
	invoker invoker.Invoker
}

// Logger returns the Logger associated with the gatherInvoker's
// context, or an appropriate (non-nil) default if the invoker's
// context has no logger associated.
func (gi *gatherInvoker) Logger() *logger.Logger {
	return logger.Ctx(gi.context)
}

// Invoke wraps Invoker.Invoke.
func (gi *gatherInvoker) Invoke(name string, arg ...string) (string, error) {
	ctx := gi.context
	invoker := gi.invoker

	gi.Logger().Debug().
		Strs("command", append([]string{name}, arg...)).
		Msg("Invoking")

	out, err := invoker.Invoke(ctx, name, arg...)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// ReadFile works like [os.ReadFile].
func (gi *gatherInvoker) ReadFile(name string) (string, error) {
	return gi.Invoke("cat", name)
}
