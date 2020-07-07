package param

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	. "replicate.ai/cli/pkg/testutil"
)

func TestUnmarshalValue(t *testing.T) {
	for _, tt := range []struct {
		input    string
		hasError bool
		expected *Value
	}{
		{"-1.0", false, &Value{floatVal: FP(-1.0)}},
		{"123.456", false, &Value{floatVal: FP(123.456)}},
		{"-1", false, &Value{intVal: IP(-1)}},
		{"0", false, &Value{intVal: IP(0)}},
		{"123456", false, &Value{intVal: IP(123456)}},
		{"true", false, &Value{boolVal: BP(true)}},
		{"false", false, &Value{boolVal: BP(false)}},
		{"\"bar\"", false, &Value{stringVal: SP("bar")}},
		{"\"bar, baz\"", false, &Value{stringVal: SP("bar, baz")}},
		{"[bar, baz]", true, nil},
	} {
		actual := Value{}
		err := json.Unmarshal([]byte(tt.input), &actual)
		if tt.hasError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, *tt.expected, actual)
		}
	}
}

func TestMarshalValue(t *testing.T) {
	for _, tt := range []struct {
		input    Value
		hasError bool
		expected string
	}{
		{Value{floatVal: FP(0.1)}, false, "0.1"},
		{Value{intVal: IP(10)}, false, "10"},
		{Value{boolVal: BP(false)}, false, "false"},
		{Value{stringVal: SP("bar")}, false, "\"bar\""},
		{Value{}, true, ""},
	} {
		actual, err := json.Marshal(&tt.input)
		if tt.hasError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(actual))
		}
	}
}
