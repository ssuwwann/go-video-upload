package exec

import (
	"context"
	"os/exec"
)

type Runner interface {
	Run(context context.Context, name string, args ...string) ([]byte, error)
	RunWithInput(context context.Context, input []byte, name string, args ...string) ([]byte, error)
}

type CommandRunner struct{}

func NewCommandRunner() *CommandRunner {
	return &CommandRunner{}
}

func (runner *CommandRunner) Run(context context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(context, name, args...)

	return cmd.CombinedOutput()
}

func (runner *CommandRunner) RunWithInput(context context.Context, input []byte, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(context, name, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close()
		stdin.Write(input)
	}()

	return cmd.CombinedOutput()
}
