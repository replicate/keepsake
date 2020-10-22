package param

import (
	"encoding/json"
)

type ValueMap map[string]*Value

func (m *ValueMap) UnmarshalJSON(data []byte) error {
	// https://stackoverflow.com/questions/43176625/call-json-unmarshal-inside-unmarshaljson-function-without-causing-stack-overflow
	type valuemap2 ValueMap
	if err := json.Unmarshal(data, (*valuemap2)(m)); err != nil {
		return err
	}
	mval := *m
	for k := range mval {
		if mval[k] == nil {
			mval[k] = None()
		}
	}
	return nil
}
