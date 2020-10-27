package param

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/replicate/replicate/go/pkg/testutil"
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
		{`"bar"`, false, &Value{stringVal: SP("bar")}},
		{`"bar, baz"`, false, &Value{stringVal: SP("bar, baz")}},
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

	// NOTE(bfirsh): It was quite hard to debug which tests in the above code were failing,
	// because the line numbers were meaningless, so I added these new ones as explicit,
	// more verbose tests

	var val Value

	require.NoError(t, json.Unmarshal([]byte(`{"foo": "bar"}`), &val))
	require.Equal(t, Value{objectVal: map[string]interface{}{"foo": "bar"}}, val)

	require.NoError(t, json.Unmarshal([]byte(`["bar", "baz"]`), &val))
	require.Equal(t, Value{objectVal: []interface{}{"bar", "baz"}}, val)

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

	require.Equal(t,
		shim([]byte(`{"foo":"bar"}`), nil),
		shim(json.Marshal(Object(map[string]interface{}{"foo": "bar"}))),
	)
}

func TestShortString(t *testing.T) {
	require.Equal(t, "hell", String("hell").ShortString(5, 5))
	require.Equal(t, "he...", String("helloo").ShortString(5, 5))

	require.Equal(t, `{"foo":"bar"}`, Object(map[string]interface{}{"foo": "bar"}).ShortString(15, 5))
	require.Equal(t, `{"bar":"baz"...`, Object(map[string]interface{}{"foo": "bar", "bar": "baz"}).ShortString(15, 5))

	// Ints don't get truncated
	require.Equal(t, "1234567890", Int(1234567890).ShortString(5, 5))

	// Floats
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

func TestEqual(t *testing.T) {
	foobar := Object(map[string]interface{}{"foo": "bar"})
	require.Equal(t, shim(true, nil), shim(foobar.Equal(Object(map[string]interface{}{"foo": "bar"}))))
	require.Equal(t, shim(false, nil), shim(foobar.Equal(Object(map[string]interface{}{"foo": "baz"}))))
	require.Equal(t, shim(false, nil), shim(foobar.Equal(Object([]interface{}{"foo", "baz"}))))
	require.Equal(t, shim(true, nil), shim(None().Equal(None())))
	require.Equal(t, shim(false, nil), shim(Int(1).Equal(None())))
	require.Equal(t, shim(false, nil), shim(None().Equal(Int(1))))
}

func TestGreaterThan(t *testing.T) {
	require.Equal(t, shim(true, nil), shim(Float(1.5).GreaterThan(Int(1))))
	require.Equal(t, shim(false, nil), shim(Int(1).GreaterThan(Float(1.5))))
	require.Equal(t, shim(false, nil), shim(Object(map[string]interface{}{"foo": "bar"}).GreaterThan(Object(map[string]interface{}{"foo": "bar"}))))
	require.Equal(t, shim(false, nil), shim(None().GreaterThan(None())))
	require.Equal(t, shim(false, nil), shim(None().GreaterThan(Int(1))))
	require.Equal(t, shim(false, nil), shim(Int(1).GreaterThan(None())))
}

func TestLessThan(t *testing.T) {
	require.Equal(t, shim(false, nil), shim(Float(1.5).LessThan(Int(1))))
	require.Equal(t, shim(true, nil), shim(Int(1).LessThan(Float(1.5))))
	require.Equal(t, shim(false, nil), shim(Object(map[string]interface{}{"foo": "bar"}).GreaterThan(Object(map[string]interface{}{"foo": "bar"}))))
	require.Equal(t, shim(false, nil), shim(None().LessThan(None())))
	require.Equal(t, shim(false, nil), shim(None().LessThan(Int(1))))
	require.Equal(t, shim(false, nil), shim(Int(1).LessThan(None())))
}

func TestType(t *testing.T) {
	require.Equal(t, TypeObject, Object(map[string]interface{}{"foo": "bar"}).Type())
}

func TestVal(t *testing.T) {
	require.Equal(t, map[string]interface{}{"foo": "bar"}, Object(map[string]interface{}{"foo": "bar"}).ObjectVal())
}

func TestPythonString(t *testing.T) {
	require.Equal(t, `{"foo":"bar"}`, Object(map[string]interface{}{"foo": "bar"}).PythonString())
}

func shim(v ...interface{}) []interface{} {
	return v
}
