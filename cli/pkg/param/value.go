package param

// TODO(andreas): rename to something more computer sciency -- this is effectively a union type

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

type Type string

const (
	TypeInt    Type = "int"
	TypeFloat  Type = "float"
	TypeString Type = "string"
	TypeBool   Type = "bool"
	TypeObject Type = "object"
	TypeNone   Type = "none"
)

var Types = []Type{TypeInt, TypeFloat, TypeString, TypeBool}

// TODO(bfirsh): could complexity be reduced here if it were implemented as interface{}?

type Value struct {
	intVal    *int
	floatVal  *float64
	stringVal *string
	boolVal   *bool
	// objectVal is anything that can't be unmarshalled into the above types, including lists
	objectVal interface{}
	isNone    bool
}

func (v Value) MarshalJSON() ([]byte, error) {
	switch {
	case v.boolVal != nil:
		return json.Marshal(v.boolVal)
	case v.intVal != nil:
		return json.Marshal(v.intVal)
	case v.floatVal != nil:
		return json.Marshal(v.floatVal)
	case v.stringVal != nil:
		return json.Marshal(v.stringVal)
	case v.objectVal != nil:
		return json.Marshal(v.objectVal)
	case v.isNone:
		return []byte("null"), nil
	}
	return nil, fmt.Errorf("No default parameter has been defined")
}

// Note that unlike everywhere else, this function uses a pointer for v. This is because this
// function needs to modify the struct in place.
func (v *Value) UnmarshalJSON(data []byte) error {
	// FIXME(bfirsh): this might be more robust if it unmarshalled to interface{}
	// then we used reflect? the error returned from json.Unmarshal might be other things
	if i := new(int); json.Unmarshal(data, i) == nil {
		v.intVal = i
		return nil
	}
	if f := new(float64); json.Unmarshal(data, f) == nil {
		v.floatVal = f
		return nil
	}
	if b := new(bool); json.Unmarshal(data, b) == nil {
		v.boolVal = b
		return nil
	}
	if string(data) == "\"null\"" || string(data) == "\"None\"" {
		v.isNone = true
		return nil
	}
	if s := new(string); json.Unmarshal(data, s) == nil {
		v.stringVal = s
		return nil
	}
	return json.Unmarshal(data, &v.objectVal)
}

// ParseFromString attempts to turn a string passed to a filter into a value.
// They aren't valid JSON because they're provided by a human (strings might not have quotes, etc).
func ParseFromString(s string) Value {
	data := []byte(s)
	v := Value{}
	if s == "null" || s == "None" {
		v.isNone = true
		return v
	}
	if i := new(int); json.Unmarshal(data, i) == nil {
		v.intVal = i
		return v
	}
	if f := new(float64); json.Unmarshal(data, f) == nil {
		v.floatVal = f
		return v
	}
	if strings.ToLower(s) == "false" {
		val := false
		v.boolVal = &val
		return v
	}
	if strings.ToLower(s) == "true" {
		val := true
		v.boolVal = &val
		return v
	}
	if json.Unmarshal(data, &v.objectVal) == nil {
		return v
	}
	v.stringVal = &s
	return v
}

func (v Value) String() string {
	if v.Type() == TypeString {
		return v.StringVal()
	}
	data, err := json.Marshal(v)
	if err != nil {
		panic("Failed to marshal value")
	}
	return string(data)
}

func Truncate(s string, maxLength int) string {
	if len(s) > maxLength && maxLength > 3 {
		return s[:maxLength-3] + "..."
	}
	return s
}

// ShortString returns a shorter version of the string, useful for displaying
// in the user interface when there isn't much space
//
// Small floats will be truncated to precision decimal points.
// Big floats will be truncated to maxLength
// Strings will be truncated to maxLength.
// Everything else is just default.
//
// TODO: some interesting stuff could be done with color here (e.g. "..." and "none" could be dimmed)
// so maybe this should be lifted out into a generic shortener in the console package.
func (v Value) ShortString(maxLength int, precision int) string {
	switch v.Type() {
	case TypeFloat:
		f := v.FloatVal()

		// For big numbers, don't truncate so eagerly
		if f > math.Pow10(precision) && f < math.Pow10(maxLength) && maxLength > precision {
			precision = maxLength
		}

		return strconv.FormatFloat(f, 'g', precision, 64)
	// Strings that need truncating
	case TypeString:
		return Truncate(v.StringVal(), maxLength)
	case TypeObject:
		return Truncate(v.String(), maxLength)
	}
	// Everything else doesn't get truncated (int, bool, none)
	return v.String()
}

func (v Value) Type() Type {
	switch {
	case v.boolVal != nil:
		return TypeBool
	case v.intVal != nil:
		return TypeInt
	case v.floatVal != nil:
		return TypeFloat
	case v.stringVal != nil:
		return TypeString
	case v.objectVal != nil:
		return TypeObject
	case v.isNone:
		return TypeNone
	}
	panic("Uninitialized param.Value")
}

func (v Value) IsNone() bool {
	return v.isNone
}

func (v Value) BoolVal() bool {
	if v.Type() != TypeBool {
		panic(fmt.Sprintf("Can't use %s as bool", v))
	}
	return *v.boolVal
}

func (v Value) IntVal() int {
	if v.Type() != TypeInt {
		panic(fmt.Sprintf("Can't use %s as int", v))
	}
	return *v.intVal
}

func (v Value) FloatVal() float64 {
	if v.Type() != TypeFloat {
		panic(fmt.Sprintf("Can't use %s as float", v))
	}
	return *v.floatVal
}

func (v Value) StringVal() string {
	if v.Type() != TypeString {
		panic(fmt.Sprintf("Can't use %s as string", v))
	}
	return *v.stringVal
}

func (v Value) ObjectVal() interface{} {
	if v.Type() != TypeObject {
		panic(fmt.Sprintf("Can't use %s as object", v))
	}
	return v.objectVal
}

func (v Value) PythonString() string {
	switch v.Type() {
	case TypeBool:
		if *v.boolVal {
			return "True"
		}
		return "False"
	case TypeInt:
		return fmt.Sprintf("%d", *v.intVal)
	case TypeFloat:
		return fmt.Sprintf("%f", *v.floatVal)
	case TypeString:
		return fmt.Sprintf("\"%s\"", *v.stringVal)
	case TypeObject:
		// FIXME(bfirsh): JSON isn't quite valid Python, but this function isn't used for anything
		// so I'm not going to fix this properly now
		data, err := json.Marshal(v.objectVal)
		if err != nil {
			// FIXME(bfirsh): this function should really return an error, but in
			// theory we are always unmarshalling this from JSON
			panic("Error marshaling object: " + err.Error())
		}
		return string(data)
	case TypeNone:
		return "None"
	}
	panic("Uninitialized param.Value")
}

func (v Value) Equal(other Value) (bool, error) {
	if !v.IsNone() && other.IsNone() || v.IsNone() && !other.IsNone() {
		return false, nil
	}
	if v.Type() != other.Type() {
		return false, fmt.Errorf("Comparing values of different types: %s and %s", v.Type(), other.Type())
	}
	switch v.Type() {
	case TypeBool:
		return v.BoolVal() == other.BoolVal(), nil
	case TypeInt:
		return v.IntVal() == other.IntVal(), nil
	case TypeFloat:
		return v.FloatVal() == other.FloatVal(), nil
	case TypeString:
		return v.StringVal() == other.StringVal(), nil
	case TypeObject:
		return reflect.DeepEqual(v.ObjectVal(), other.ObjectVal()), nil
	case TypeNone:
		return other.IsNone(), nil
	}
	return false, fmt.Errorf("Unknown value type: %s", v.Type())
}

func (v Value) NotEqual(other Value) (bool, error) {
	eq, err := v.Equal(other)
	if err != nil {
		return false, err
	}
	return !eq, nil
}

func (v Value) GreaterThan(other Value) (bool, error) {
	// Special cases
	if v.Type() == TypeFloat && other.Type() == TypeInt {
		return v.FloatVal() > float64(other.IntVal()), nil
	}
	if v.Type() == TypeInt && other.Type() == TypeFloat {
		return float64(v.IntVal()) > other.FloatVal(), nil
	}
	if v.IsNone() || other.IsNone() {
		return false, nil
	}

	if v.Type() != other.Type() {
		return false, fmt.Errorf("Comparing values of different types: %s and %s", v.Type(), other.Type())
	}
	switch v.Type() {
	case TypeBool:
		return v.BoolVal() && !other.BoolVal(), nil
	case TypeInt:
		return v.IntVal() > other.IntVal(), nil
	case TypeFloat:
		return v.FloatVal() > other.FloatVal(), nil
	case TypeString:
		return v.StringVal() > other.StringVal(), nil
	case TypeObject:
		return false, nil
	case TypeNone:
		return false, nil
	}
	return false, fmt.Errorf("Unknown value type: %s", v.Type())
}

func (v Value) GreaterOrEqual(other Value) (bool, error) {
	gt, err := v.GreaterThan(other)
	if err != nil {
		return false, err
	}
	eq, err := v.Equal(other)
	if err != nil {
		return false, err
	}
	return gt || eq, nil
}

func (v Value) LessThan(other Value) (bool, error) {
	// Special cases
	if v.Type() == TypeFloat && other.Type() == TypeInt {
		return v.FloatVal() < float64(other.IntVal()), nil
	}
	if v.Type() == TypeInt && other.Type() == TypeFloat {
		return float64(v.IntVal()) < other.FloatVal(), nil
	}
	if v.IsNone() || other.IsNone() {
		return false, nil
	}

	if v.Type() != other.Type() {
		return false, fmt.Errorf("Comparing values of different types: %s and %s", v.Type(), other.Type())
	}
	switch v.Type() {
	case TypeBool:
		return !v.BoolVal() && other.BoolVal(), nil
	case TypeInt:
		return v.IntVal() < other.IntVal(), nil
	case TypeFloat:
		return v.FloatVal() < other.FloatVal(), nil
	case TypeString:
		return v.StringVal() < other.StringVal(), nil
	case TypeNone:
		return false, nil
	}
	return false, fmt.Errorf("Unknown value type: %s", v.Type())
}

func (v Value) LessOrEqual(other Value) (bool, error) {
	lt, err := v.LessThan(other)
	if err != nil {
		return false, err
	}
	eq, err := v.Equal(other)
	if err != nil {
		return false, err
	}
	return lt || eq, nil
}

func Bool(v bool) Value {
	return Value{boolVal: &v}
}

func Int(v int) Value {
	return Value{intVal: &v}
}

func Float(v float64) Value {
	return Value{floatVal: &v}
}

func String(v string) Value {
	return Value{stringVal: &v}
}

func Object(v interface{}) Value {
	return Value{objectVal: v}
}

func None() Value {
	return Value{isNone: true}
}

func ToJSON(params ValueMap) (string, error) {
	j, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("Failed to convert params to JSON: %s", err)
	}
	return string(j), nil
}

func FromJSON(j string) (ValueMap, error) {
	params := ValueMap{}
	err := json.Unmarshal([]byte(j), &params)
	if err != nil {
		return nil, fmt.Errorf("Failed to load params from JSON: %s", err)
	}
	return params, nil
}
