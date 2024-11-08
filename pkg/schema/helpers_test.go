package schema

import (
	"errors"
	"testing"
)

func Test_isTableDoesNotExists(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "bun driver",
			err:  errors.New("panic: error while fetching applied versions: ERROR: relation \"gwt.mig_schema_versions\" does not exist (SQLSTATE=42P01)\n"),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTableDoesNotExists(tt.err); got != tt.want {
				t.Errorf("isTableDoesNotExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
