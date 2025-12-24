package amigo

import (
	"context"
)

type RunnerUpOpts struct {
	Steps int
}

type RunnerUpOptsFunc func(*RunnerUpOpts)

func RunnerUpOptionSteps(steps int) RunnerUpOptsFunc {
	return func(opts *RunnerUpOpts) {
		opts.Steps = steps
	}
}

func DefaultRunnerUpOpts() RunnerUpOpts {
	return RunnerUpOpts{
		Steps: -1,
	}
}

func (r Runner) Up(ctx context.Context, migrations []Migration, opts ...RunnerUpOptsFunc) error {
	for result := range r.UpIterator(ctx, migrations, opts...) {
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}
