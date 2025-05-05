package hostinfo

import "os/exec"

// An Invoker invokes commands and returns their output.
type Invoker interface {
	// Invoke runs the specified command and returns its combined
	// standard output and standard error.
	Invoke(name string, arg ...string) ([]byte, error)
}

// An ExecInvoker invokes commands using [exec.Command].
type ExecInvoker struct{}

// Invoke implements the Invoker interface.
func (e *ExecInvoker) Invoke(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).CombinedOutput()
}
