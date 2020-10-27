package param

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalJSON(t *testing.T) {
	valmap := make(ValueMap)

	require.NoError(t, json.Unmarshal([]byte(`{"foo": 1}`), &valmap))
	require.Equal(t, ValueMap(ValueMap{"foo": Int(1)}), valmap)

	require.NoError(t, json.Unmarshal([]byte(`{"foo": {"baz": "bop"}}`), &valmap))
	require.Equal(t, ValueMap(ValueMap{"foo": Object(map[string]interface{}{"baz": "bop"})}), valmap)

	require.NoError(t, json.Unmarshal([]byte(`{"foo": null}`), &valmap))
	require.Equal(t, ValueMap(ValueMap{"foo": None()}), valmap)

}
