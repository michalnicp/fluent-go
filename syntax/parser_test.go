package syntax

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	paths, err := filepath.Glob("testdata/*.ftl")
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		name := filepath.Base(path[:len(path)-4]) // strip .ftl
		t.Run(name, func(t *testing.T) {
			expected, err := ioutil.ReadFile(path[:len(path)-4] + ".json")
			require.NoError(t, err)

			input, err := ioutil.ReadFile(path)
			require.NoError(t, err)

			resource, err := Parse(input) // ignore parse errors, should appear as junk
			t.Logf("parse:\n%+v", err)

			actual, err := marshal(resource)
			require.NoError(t, err)

			require.JSONEq(t, string(expected), string(actual))
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	comment := Comment{"Standalone Comment"}

	actual, err := json.Marshal(comment)
	require.NoError(t, err)

	expected := `{"type":"Comment","content":"Standalone Comment"}`

	require.JSONEq(t, string(actual), expected)
}
