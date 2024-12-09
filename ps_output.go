package exec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jolt9dev/go-platform"
)

type PsOutput struct {
	Stdout    []byte
	Stderr    []byte
	Code      int
	FileName  string
	Args      []string
	StartedAt time.Time
	EndedAt   time.Time
}

func (o *PsOutput) Text() string {
	return string(o.Stdout)
}

func (o *PsOutput) Lines() []string {
	r := bytes.Split(o.Stdout, []byte(platform.EOL))
	lines := []string{}
	for _, line := range r {
		lines = append(lines, string(line))
	}
	return lines
}

func (o *PsOutput) ErrorText() string {
	return string(o.Stderr)
}

func (o *PsOutput) ErrorLines() []string {
	r := bytes.Split(o.Stdout, []byte(platform.EOL))
	lines := []string{}
	for _, line := range r {
		lines = append(lines, string(line))
	}
	return lines
}

func (o *PsOutput) Json() (interface{}, error) {
	var out interface{}
	err := json.Unmarshal([]byte(o.Stdout), &out)
	return out, err
}

func (o *PsOutput) ErrorJson() (interface{}, error) {
	var out interface{}
	err := json.Unmarshal([]byte(o.Stderr), &out)
	return out, err
}

func (o *PsOutput) Validate() (bool, error) {
	return o.ValidateWith(nil)
}

func (o *PsOutput) ValidateWith(cb func(o *PsOutput) (bool, error)) (bool, error) {
	if cb == nil {
		cb = func(o *PsOutput) (bool, error) {
			if o.Code != 0 {
				return false, fmt.Errorf("command %s failed with code %d", o.FileName, o.Code)
			}

			return true, nil
		}
	}

	return cb(o)
}
