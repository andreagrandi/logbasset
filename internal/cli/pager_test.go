package cli

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStartPager_SkipsWhenNotTTY verifies the pager is a no-op when stdout is
// not a terminal. Tests run with stdout redirected, so this is the common case.
func TestStartPager_SkipsWhenNotTTY(t *testing.T) {
	previous := IsTTY
	IsTTY = func() bool { return false }
	defer func() { IsTTY = previous }()

	assert.Nil(t, startPager())
}

// TestStartPager_PipesStdoutThroughPager exercises the full pipeline by
// pointing PAGER at `cat`, which behaves as a transparent pager: anything
// written to os.Stdout while the pager is active reappears on the original
// stdout once stop() finishes.
func TestStartPager_PipesStdoutThroughPager(t *testing.T) {
	previousTTY := IsTTY
	IsTTY = func() bool { return true }
	defer func() { IsTTY = previousTTY }()

	t.Setenv("PAGER", "cat")

	originalStdout := os.Stdout
	captureR, captureW, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = captureW

	defer func() {
		os.Stdout = originalStdout
	}()

	pager := startPager()
	require.NotNil(t, pager)
	require.NotEqual(t, captureW, os.Stdout, "startPager should swap os.Stdout for the pager pipe")

	fmt.Fprint(os.Stdout, "hello via pager\n")

	pager.stop()
	require.NoError(t, captureW.Close())

	got, err := io.ReadAll(captureR)
	require.NoError(t, err)
	assert.Equal(t, "hello via pager\n", string(got))
	assert.Equal(t, captureW, os.Stdout, "stop() should restore the original stdout")
}

// TestPagerStop_Idempotent ensures stop() can be called safely after the pager
// has already been torn down (e.g. once via errors.BeforeExit and once via the
// deferred cleanup in Execute).
func TestPagerStop_Idempotent(t *testing.T) {
	previousTTY := IsTTY
	IsTTY = func() bool { return true }
	defer func() { IsTTY = previousTTY }()

	t.Setenv("PAGER", "cat")

	originalStdout := os.Stdout
	_, captureW, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = captureW
	defer func() { os.Stdout = originalStdout }()

	pager := startPager()
	require.NotNil(t, pager)

	pager.stop()
	assert.NotPanics(t, func() { pager.stop() })
}

// TestPagerStop_NilReceiver guards against panicking when no pager was started.
func TestPagerStop_NilReceiver(t *testing.T) {
	var p *pagerProcess
	assert.NotPanics(t, func() { p.stop() })
}
