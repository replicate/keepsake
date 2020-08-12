package param

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTypesGood(t *testing.T) {
	for _, tt := range []struct {
		input    string
		expected filter
	}{
		{"foo=0.001", filter{"foo", OperatorEqual, Float(0.001)}},
		{"foo = bar", filter{"foo", OperatorEqual, String("bar")}},
		{"f = bar", filter{"f", OperatorEqual, String("bar")}},
		{" foo=bar ", filter{"foo", OperatorEqual, String("bar")}},
		{"foo = 1", filter{"foo", OperatorEqual, Int(1)}},
		{"foo = 1.5", filter{"foo", OperatorEqual, Float(1.5)}},
		{"foo = true", filter{"foo", OperatorEqual, Bool(true)}},
		{"foo = false", filter{"foo", OperatorEqual, Bool(false)}},
		{"foo = null", filter{"foo", OperatorEqual, None()}},
		{"foo = None", filter{"foo", OperatorEqual, None()}},
	} {
		actual, err := parse(tt.input)
		require.NoError(t, err)
		require.Equal(t, &tt.expected, actual)
	}
}

func TestParseOperatorsGood(t *testing.T) {
	for _, tt := range []struct {
		input    string
		expected filter
	}{
		{"foo = bar", filter{"foo", OperatorEqual, String("bar")}},
		{"foo != bar", filter{"foo", OperatorNotEqual, String("bar")}},
		{"foo < bar", filter{"foo", OperatorLessThan, String("bar")}},
		{"foo <= bar", filter{"foo", OperatorLessOrEqual, String("bar")}},
		{"foo > bar", filter{"foo", OperatorGreaterThan, String("bar")}},
		{"foo >= bar", filter{"foo", OperatorGreaterOrEqual, String("bar")}},
		{"foo foo >= bar", filter{"foo foo", OperatorGreaterOrEqual, String("bar")}},
		{"foo >= bar bar", filter{"foo", OperatorGreaterOrEqual, String("bar bar")}},
	} {
		actual, err := parse(tt.input)
		require.NoError(t, err)
		require.Equal(t, &tt.expected, actual)
	}
}

func TestParseBad(t *testing.T) {
	for _, input := range []string{
		"",
		" ",
		"foo",
		"foo =",
		"=",
		"= bar",
		"foo >> bar",
	} {
		_, err := parse(input)
		require.Error(t, err)
	}
}
