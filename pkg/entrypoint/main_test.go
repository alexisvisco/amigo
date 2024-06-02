package entrypoint

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_sliceArrayOrDefault(t *testing.T) {
	type args struct {
		iarray []int
		x      int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "should return the last 2 migrations",
			args: args{
				iarray: []int{1, 2, 3},
				x:      2,
			},

			want: []int{2, 3},
		},

		{
			name: "should return the last 1 migration",
			args: args{
				iarray: []int{1, 2, 3},
				x:      1,
			},
			want: []int{3},
		},

		{
			name: "should return the last 3 migration if less than x",
			args: args{
				iarray: []int{1, 2, 3},
				x:      5,
			},
			want: []int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sliceArrayOrDefault(tt.args.iarray, tt.args.x)
			require.Equal(t, tt.want, got)
		})
	}
}
