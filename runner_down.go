package amigo

import (
	"context"
)

type runnerDownOpts struct {
	Steps int
}

type RunnerDownOptsFunc func(*runnerDownOpts)

func RunnerDownOptionSteps(steps int) RunnerDownOptsFunc {
	return func(opts *runnerDownOpts) {
		opts.Steps = steps
	}
}

func defaultRunnerDownOpts() runnerDownOpts {
	return runnerDownOpts{
		Steps: -1,
	}
}

func (r *Runner) Down(ctx context.Context, migrations []Migration, opts ...RunnerDownOptsFunc) error {
	for result := range r.DownIterator(ctx, migrations, opts...) {
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}
