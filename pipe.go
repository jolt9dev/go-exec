package exec

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"
)

type Pipeline struct {
	cmds []*Cmd
}

func (p *Pipeline) Pipe(subcommands ...*Cmd) *Pipeline {
	p.cmds = append(p.cmds, subcommands...)
	return p
}

func (p *Pipeline) PipeCommand(subcommands ...string) *Pipeline {
	set := make([]*Cmd, 0)
	for _, cmd := range subcommands {
		c := Command(cmd)
		set = append(set, c)
	}

	p.cmds = append(p.cmds, set...)

	return p
}

func (c *Cmd) Pipe(subcommands ...*Cmd) *Pipeline {
	set := make([]*Cmd, 0)
	set = append(set, c)
	set = append(set, subcommands...)
	p := &Pipeline{cmds: set}
	return p
}

func (c *Cmd) PipeCommand(subcommands ...string) *Pipeline {
	set := make([]*Cmd, 0)
	set = append(set, c)
	for _, cmd := range subcommands {
		c := Command(cmd)
		set = append(set, c)
	}

	p := &Pipeline{cmds: set}

	return p
}

func (p *Pipeline) Output() (*PsOutput, error) {
	var o PsOutput
	o.Stdout = make([]byte, 0)
	o.Stderr = make([]byte, 0)
	o.StartedAt = time.Now().UTC()

	lastIndex := len(p.cmds) - 1
	r, w := io.Pipe()

	errs := make([]error, 0)
	count := 0
	var outb, errb bytes.Buffer

	prev := p.cmds[0]
	for i, cmd := range p.cmds {
		count++
		if i == 0 {
			cmd.Stdout = w
			err := cmd.Start()
			if err != nil {
				errs = append(errs, err)
				break
			}
		} else if i == lastIndex {
			cmd.Stdin = r
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err := cmd.Start()
			if err != nil {
				errs = append(errs, err)
				break
			}

			err = prev.Wait()
			if err != nil {
				errs = append(errs, err)
			}

			err = w.Close()
			if err != nil {
				errs = append(errs, err)
			}

			err = cmd.Wait()
			o.EndedAt = time.Now().UTC()
			if err != nil {
				errs = append(errs, err)
			}

			o.FileName = cmd.Path
			o.Args = cmd.Args
			o.Code = cmd.Cmd.ProcessState.ExitCode()
			o.Stdout = outb.Bytes()
			o.Stderr = errb.Bytes()
		} else {
			r2, w2 := io.Pipe()
			cmd.Stdin = r
			cmd.Stdout = w2
			err := cmd.Start()
			if err != nil {
				errs = append(errs, err)
				break
			}

			err = prev.Wait()
			if err != nil {
				errs = append(errs, err)
			}

			err = w.Close()
			if err != nil {
				errs = append(errs, err)
			}

			w = w2
			prev = cmd
			if err != r.Close() {
				errs = append(errs, err)
			}

			r = r2
		}
	}

	if len(errs) > 0 {

		e := errors.Join(errs...)
		return &o, e
	}

	return &o, nil
}

func (p *Pipeline) Run() (*PsOutput, error) {
	var o PsOutput
	o.Stdout = make([]byte, 0)
	o.Stderr = make([]byte, 0)
	o.StartedAt = time.Now().UTC()

	lastIndex := len(p.cmds) - 1
	r, w := io.Pipe()

	errs := make([]error, len(p.cmds))
	count := 0

	prev := p.cmds[0]
	for i, cmd := range p.cmds {
		count++
		if i == 0 {
			cmd.Stdout = w
			err := cmd.Start()
			if err != nil {
				errs = append(errs, err)
				break
			}
		} else if i == lastIndex {
			cmd.Stdin = r
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Start()
			if err != nil {
				errs = append(errs, err)
				break
			}

			err = prev.Wait()
			if err != nil {
				errs = append(errs, err)
			}

			err = w.Close()
			if err != nil {
				errs = append(errs, err)
			}

			err = cmd.Wait()
			o.EndedAt = time.Now().UTC()
			if err != nil {
				errs = append(errs, err)
			}

			o.FileName = cmd.Path
			o.Args = cmd.Args
			o.Code = cmd.Cmd.ProcessState.ExitCode()
		} else {
			r2, w2 := io.Pipe()
			cmd.Stdin = r
			cmd.Stdout = w2
			err := cmd.Start()
			if err != nil {
				errs = append(errs, err)
				break
			}

			err = prev.Wait()
			if err != nil {
				errs = append(errs, err)
			}

			err = w.Close()
			if err != nil {
				errs = append(errs, err)
			}

			w = w2
			prev = cmd
			if err != r.Close() {
				errs = append(errs, err)
			}

			r = r2
		}
	}

	if len(errs) > 0 {
		e := errors.Join(errs...)
		return &o, e
	}

	return &o, nil
}
