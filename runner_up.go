package amigo

import (
	"context"
)

type runnerUpOpts struct {
	Steps int
}

type RunnerUpOptsFunc func(*runnerUpOpts)

func RunnerUpOptionSteps(steps int) RunnerUpOptsFunc {
	return func(opts *runnerUpOpts) {
		opts.Steps = steps
	}
}

func defaultRunnerUpOpts() runnerUpOpts {
	return runnerUpOpts{
		Steps: -1,
	}
}

func (r *Runner) Up(ctx context.Context, migrations []Migration, opts ...RunnerUpOptsFunc) error {
	for result := range r.UpIterator(ctx, migrations, opts...) {
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}
