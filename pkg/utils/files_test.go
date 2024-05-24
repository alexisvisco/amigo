package utils

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestCreateOrOpenFile(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name    string
		path    string
		prepare func()
		assert  func(*os.File, error)
	}{
		{
			name: "Test with fresh new file and directory not exists",
			path: os.TempDir() + "/abc/test.txt",
			assert: func(file *os.File, err error) {
				require.NoError(t, err)

				// check that path dir exists
				_, err = os.Stat(os.TempDir() + "/abc")
				require.NoError(t, err)
			},
		},
		{
			name: "Test with a file that exists",
			path: os.TempDir() + "/abc/test.txt",
			prepare: func() {
				f, _ := os.Create(os.TempDir() + "/efg/test.txt")
				f.WriteString("bonjour")
			},
			assert: func(file *os.File, err error) {
				require.NoError(t, err)

				// check that path dir exists
				_, err = os.Stat(os.TempDir() + "/abc")
				require.NoError(t, err)

				// check that file is empty
				stat, _ := file.Stat()
				require.Equal(t, int64(0), stat.Size())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare()
			}
			file, err := CreateOrOpenFile(tt.path)
			tt.assert(file, err)
		})
	}
}
