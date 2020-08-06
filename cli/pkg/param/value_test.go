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

func TestShortString(t *testing.T) {
	require.Equal(t, "hello", String("hello").ShortString(5, 5))
	require.Equal(t, "he...", String("helloo").ShortString(5, 5))
	require.Equal(t, "0.12", Float(0.12).ShortString(10, 5))
	require.Equal(t, "12.34", Float(12.34).ShortString(10, 5))
	require.Equal(t, "0.12346", Float(0.1234567).ShortString(10, 5))
	// TODO: always keep a decimal point if it's float
	// require.Equal(t, "12345.6", Float(12345.6789).ShortString(10, 5))
	require.Equal(t, "0.0012346", Float(0.001234567).ShortString(10, 5))
	require.Equal(t, "1.2346e-06", Float(0.000001234567).ShortString(10, 5))
	require.Equal(t, "1.2346e-11", Float(0.00000000001234567).ShortString(10, 5))

	// Big numbers
	require.Equal(t, "123456.789", Float(123456.789).ShortString(10, 5))
	require.Equal(t, "123456789.1", Float(123456789.123456789).ShortString(10, 5))
	require.Equal(t, "1.2346e+13", Float(12345678912345.6789).ShortString(10, 5))

}
