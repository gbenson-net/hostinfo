package hostinfo

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"gbenson.net/go/invoker"
	"gbenson.net/go/logger"
	"gotest.tools/v3/assert"
)

// testctx returns a [context.Context] suitable for tests.
func testctx(t *testing.T) context.Context {
	return logger.TestContext(t)
}

// exec runs the specified gatherer with an [invoker.ExecInvoker].
func exec(t *testing.T, g gatherer) (HostInfo, error) {
	t.Helper()
	return invoke(t, &invoker.ExecInvoker{}, g)
}

// invoke runs the specified gatherer with the specified [invoker.Invoker].
func invoke(t *testing.T, i invoker.Invoker, g gatherer) (result HostInfo, err error) {
	t.Helper()
	err = g(&gatherInvoker{testctx(t), i}, &result)
	return
}

// assertExec runs the specified gatherer with an [ExecInvoker],
// marking the test as having failed and stopping its execution
// if the gatherer returns an error.
func assertExec(t *testing.T, g gatherer) HostInfo {
	t.Helper()
	result, err := exec(t, g)
	assert.NilError(t, err)
	return result
}

// assertMock runs the specified gatherer with a [MockInvoker],
// marking the test as having failed and stopping its execution
// if the gatherer returns an error.
func assertMock(t *testing.T, g gatherer, mi *invoker.MockInvoker) HostInfo {
	t.Helper()
	result, err := invoke(t, mi, g)
	assert.NilError(t, err)
	assert.NilError(t, mi.ExpectationsWereMet())
	return result
}

// TestGather flexes the expected use case.
func TestGather(t *testing.T) {
	ctx := testctx(t)
	hostinfo, err := Gather(ctx, &invoker.ExecInvoker{})
	assert.NilError(t, err)

	data, err := json.Marshal(hostinfo)
	assert.NilError(t, err)

	const filename = "localhost.json"
	assert.NilError(t, os.WriteFile(filename, data, 0666))
	logger.Ctx(ctx).Debug().Str("filename", filename).Msg("Wrote")
}
