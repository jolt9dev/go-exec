package exec_test

import (
	"testing"

	exec "github.com/jolt9dev/go-exec"
	assert2 "github.com/stretchr/testify/assert"
)

func TestSplitArgs(t *testing.T) {
	assert := assert2.New(t)

	args := exec.SplitArgs("test")
	assert.Equal(1, len(args))
	assert.Equal("test", args[0])

	args = exec.SplitArgs("test test")
	assert.Equal(2, len(args))
	assert.Equal("test", args[0])
	assert.Equal("test", args[1])

	args = exec.SplitArgs("test \"test\"")
	assert.Equal(2, len(args))
	assert.Equal("test", args[0])
	assert.Equal("test", args[1])

	args = exec.SplitArgs("--test 'test'")
	assert.Equal(2, len(args))
	assert.Equal("--test", args[0])
	assert.Equal("test", args[1])

	// multiline
	// be forgiving with \n when its not quoted and not escaped
	args = exec.SplitArgs(`--test 'test'
--test2 'test2'`)
	assert.Equal(4, len(args))
	assert.Equal("--test", args[0])
	assert.Equal("test", args[1])
	assert.Equal("\n--test2", args[2])
	assert.Equal("test2", args[3])

	// multiline with with backslash
	args = exec.SplitArgs(`--test 'test' \
--test2 'test2'`)
	assert.Equal(4, len(args))
	assert.Equal("--test", args[0])
	assert.Equal("test", args[1])
	assert.Equal("--test2", args[2])
	assert.Equal("test2", args[3])

	// multiline with with backtick
	args = exec.SplitArgs("--test 'test' `\n--test2 'test2'")
	assert.Equal(4, len(args))
	assert.Equal("--test", args[0])
	assert.Equal("test", args[1])
	assert.Equal("--test2", args[2])
	assert.Equal("test2", args[3])

	// multiline with quotes, quotes should be inclusive of the quotes
	args = exec.SplitArgs(`--test 'test' "
--test2 'test2'"`)
	assert.Equal(3, len(args))
	assert.Equal("--test", args[0])
	assert.Equal("test", args[1])
	assert.Equal("\n--test2 'test2'", args[2])
}
