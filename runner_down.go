package amigo

import (
	"context"
)

type RunnerDownOpts struct {
	Steps int
}

type RunnerDownOptsFunc func(*RunnerDownOpts)

func RunnerDownOptionSteps(steps int) RunnerDownOptsFunc {
	return func(opts *RunnerDownOpts) {
		opts.Steps = steps
	}
}

func DefaultRunnerDownOpts() RunnerDownOpts {
	return RunnerDownOpts{
		Steps: -1,
	}
}

func (r Runner) Down(ctx context.Context, migrations []Migration, opts ...RunnerDownOptsFunc) error {
	for result := range r.DownIterator(ctx, migrations, opts...) {
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}
