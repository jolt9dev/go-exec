package exec_test

import (
	"strings"
	"testing"

	exec "github.com/jolt9dev/go-exec"
	"github.com/stretchr/testify/assert"
)

func TestNewCommandOutput(t *testing.T) {
	echo, ok := exec.Which("echo")
	if !ok {
		t.Skip("echo not found")
	}

	o, err := exec.New(echo, "hello").Output()
	assert.NoError(t, err)
	assert.Equal(t, 0, o.Code)
	assert.Equal(t, "hello", strings.TrimSpace(o.Text()))
}

func TestCommandOutput(t *testing.T) {
	_, ok := exec.Which("echo")
	if !ok {
		t.Skip("echo not found")
	}

	cmd := "echo 'hello world'"

	o, err := exec.Command(cmd).Output()
	assert.NoError(t, err)
	assert.Equal(t, 0, o.Code)
	assert.Equal(t, "hello world", strings.TrimSpace(o.Text()))
}

func TestPipeCommand(t *testing.T) {
	_, hasGrep := exec.Which("grep")
	_, hasEcho := exec.Which("echo")

	if !hasEcho || !hasGrep {
		t.Skip("grep or echo not found")
	}

	o, err := exec.Command("echo 'Hello World'").PipeCommand("grep Hello").Output()
	assert.NoError(t, err)
	assert.Equal(t, 0, o.Code)
	assert.Equal(t, "Hello World", strings.TrimSpace(o.Text()))
}
