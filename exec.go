package exec

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	STDIO_INHERIT = 0
	STDIO_PIPED   = 1
	STDIO_NULL    = 2
)

var (
	logger func(cmd *Cmd)
)

type Cmd struct {
	*exec.Cmd
	logger        func(cmd *Cmd)
	disableLogger bool
}

func New(name string, args ...string) *Cmd {
	cmd := exec.Command(name, args...)
	return &Cmd{Cmd: cmd}
}

func SetLigger(f func(cmd *Cmd)) {
	logger = f
}

func (c *Cmd) SetLogger(f func(cmd *Cmd)) {
	c.logger = f
}

func (c *Cmd) DisableLogger() {
	c.disableLogger = true
}

// Command parses the command and arguments and returns a new Cmd
// with the parsed command and arguments
// Example:
//
//	Command("echo hello world")
//	Command("echo 'hello world'")
func Command(command string) *Cmd {
	args := SplitArgs(command)
	command = args[0]
	args = args[1:]
	return New(command, args...)
}

func Run(command string) (*PsOutput, error) {
	return Command(command).Run()
}

func Output(command string) (*PsOutput, error) {
	return Command(command).Output()
}

func (c *Cmd) AppendArgs(args ...string) *Cmd {
	c.Cmd.Args = append(c.Cmd.Args, args...)
	return c
}

func (c *Cmd) PrependArgs(args ...string) *Cmd {
	c.Cmd.Args = append([]string{args[0]}, c.Cmd.Args...)
	return c
}

func (c *Cmd) WithArgs(args ...string) *Cmd {
	c.Cmd.Args = args
	return c
}

func (c *Cmd) AppendEnv(env ...string) *Cmd {
	c.Cmd.Env = append(c.Cmd.Env, env...)
	return c
}

func (c *Cmd) PrependEnv(env ...string) *Cmd {
	c.Cmd.Env = append([]string{env[0]}, c.Cmd.Env...)
	return c
}

func (c *Cmd) WithEnvMap(env map[string]string) *Cmd {
	data := make([]string, 0)
	for k, v := range env {
		data = append(data, k+"="+v)
	}
	return c.WithEnv(data...)
}

func (c *Cmd) WithEnv(env ...string) *Cmd {
	c.Cmd.Env = env
	return c
}

func (c *Cmd) WithTimeout(timeout time.Duration) *Cmd {
	c.Cmd.WaitDelay = timeout
	return c
}

func (c *Cmd) WithCwd(dir string) *Cmd {
	c.Cmd.Dir = dir
	return c
}

func (c *Cmd) WithStdin(stdin io.Reader) *Cmd {
	c.Cmd.Stdin = stdin
	return c
}

func (c *Cmd) WithStdout(stdout io.Writer) *Cmd {
	c.Cmd.Stdout = stdout
	return c
}

func (c *Cmd) WithStderr(stderr io.Writer) *Cmd {
	c.Cmd.Stderr = stderr
	return c
}

func (c *Cmd) WithStdio(stdin, stdout, stderr int) *Cmd {
	switch stdin {
	case STDIO_INHERIT:
		c.Cmd.Stdin = os.Stdin
	case STDIO_PIPED:
		c.Cmd.Stdin = bytes.NewBuffer(nil)
	case STDIO_NULL:
		c.Cmd.Stdin = nil
	}

	switch stdout {
	case STDIO_INHERIT:
		c.Cmd.Stdout = os.Stdout
	case STDIO_PIPED:
		c.Cmd.Stdout = bytes.NewBuffer(nil)
	case STDIO_NULL:
		c.Cmd.Stdout = nil
	}

	switch stderr {
	case STDIO_INHERIT:
		c.Cmd.Stderr = os.Stderr
	case STDIO_PIPED:
		c.Cmd.Stderr = bytes.NewBuffer(nil)
	case STDIO_NULL:
		c.Cmd.Stderr = nil
	}

	return c
}

// Runs the command quietly, without any PsOutput
func (c *Cmd) Quiet() (*PsOutput, error) {
	c.Cmd.Stdout = nil
	c.Cmd.Stderr = nil
	var out PsOutput
	out.FileName = c.Cmd.Path
	out.Args = c.Cmd.Args
	out.Stdout = make([]byte, 0)
	out.Stderr = make([]byte, 0)
	// use utc time
	out.StartedAt = time.Now().UTC()

	err := c.Start()
	if err != nil {
		return nil, err
	}

	err = c.Wait()
	if err != nil {
		return nil, err
	}
	out.EndedAt = time.Now().UTC()
	out.Code = c.Cmd.ProcessState.ExitCode()

	return &out, nil
}

// Runs the command and waits for it to finish
// PsOutputs are inherited from the current process and
// are not captured
func (c *Cmd) Run() (*PsOutput, error) {
	c.Cmd.Stdout = os.Stdout
	c.Cmd.Stderr = os.Stderr
	c.Cmd.Stdin = os.Stdin
	var out PsOutput
	out.FileName = c.Cmd.Path
	out.Args = c.Cmd.Args
	// use utc time
	out.StartedAt = time.Now().UTC()
	out.Stdout = make([]byte, 0)
	out.Stderr = make([]byte, 0)

	err := c.Start()
	if err != nil {
		out.EndedAt = time.Now().UTC()
		out.Code = 1
		return &out, err
	}

	err = c.Wait()
	if err != nil {
		out.EndedAt = time.Now().UTC()
		out.Code = 1
		return &out, err
	}

	out.EndedAt = time.Now().UTC()
	out.Code = c.Cmd.ProcessState.ExitCode()
	return &out, nil
}

// Runs the command and captures the PsOutput
// PsOutputs are captured from the current process and
// are not inherited
func (c *Cmd) Output() (*PsOutput, error) {

	var out PsOutput
	out.Stdout = make([]byte, 0)
	out.Stderr = make([]byte, 0)
	out.StartedAt = time.Now().UTC()
	out.FileName = c.Cmd.Path
	out.Args = c.Cmd.Args

	var outb, errb bytes.Buffer
	c.Stdout = &outb
	c.Stderr = &errb

	err := c.Start()
	if err != nil {
		out.EndedAt = time.Now().UTC()
		out.Code = 1
		return &out, err
	}

	err = c.Wait()
	if err != nil {
		out.EndedAt = time.Now().UTC()
		out.Code = 1
		return &out, err
	}

	out.EndedAt = time.Now().UTC()
	out.Code = c.Cmd.ProcessState.ExitCode()
	out.Stdout = outb.Bytes()
	out.Stderr = errb.Bytes()

	return &out, nil
}

func (c *Cmd) Start() error {
	if c.disableLogger {
		return c.Cmd.Start()
	}

	if c.logger != nil {
		c.logger(c)
	}

	if logger != nil {
		logger(c)
	}

	p := c.Cmd.Path
	if p != "" && !filepath.IsAbs(p) {
		p2, err := Find(p, nil)
		if err == nil {
			c.Cmd.Path = p2
		}
	}

	return c.Cmd.Start()
}

func (c *Cmd) Wait() error {
	return c.Cmd.Wait()
}
