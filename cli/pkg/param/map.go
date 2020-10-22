package param

import (
	"encoding/json"
)

type ValueMap map[string]Value

type pointerValueMap map[string]*Value

func (m *ValueMap) UnmarshalJSON(data []byte) error {
	var pointerMap pointerValueMap
	if err := json.Unmarshal(data, &pointerMap); err != nil {
		return err
	}
	*m = make(ValueMap)
	for k, v := range pointerMap {
		if v == nil {
			(*m)[k] = None()
		} else {
			(*m)[k] = *v
		}
	}
	return nil
}
