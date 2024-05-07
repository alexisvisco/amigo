package schema

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTableName_Name(t *testing.T) {
	tts := []struct {
		name           string
		expectedSchema string
		expectedTable  string
	}{
		{
			name:           "public.table",
			expectedSchema: "public",
			expectedTable:  "table",
		},
		{
			name:           "table",
			expectedSchema: "public",
			expectedTable:  "table",
		},
		{
			name:           "schema.table",
			expectedSchema: "schema",
			expectedTable:  "table",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			tn := TableName(tt.name)
			require.Equal(t, tt.expectedSchema, tn.Schema())
			require.Equal(t, tt.expectedTable, tn.Name())
		})
	}
}
