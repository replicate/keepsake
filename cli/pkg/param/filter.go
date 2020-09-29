package param

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/araddon/dateparse"
)

type ValueGetter interface {
	GetValue(name string) *Value
}

type Filters struct {
	filters []*filter
}

type filter struct {
	name     string
	operator Operator
	value    *Value
}

type Operator int

const (
	OperatorEqual Operator = iota
	OperatorNotEqual
	OperatorGreaterThan
	OperatorGreaterOrEqual
	OperatorLessThan
	OperatorLessOrEqual
)

var parseRegex = regexp.MustCompile("^([-a-zA-Z0-9_ ]*[-a-zA-Z0-9_]+) *([<>=!]+) *(.+)$")

func MakeFilters(strings []string) (*Filters, error) {
	filters := &Filters{}
	for _, s := range strings {
		if err := filters.appendParsed(s); err != nil {
			return nil, err
		}
	}
	return filters, nil
}

// SetExclusive sets a filter exclusively, deleting any previous
// filters with that name
func (fs *Filters) SetExclusive(name string, operator Operator, value *Value) {
	filters := []*filter{&filter{
		name:     name,
		operator: operator,
		value:    value,
	}}
	for _, f := range fs.filters {
		if f.name != name {
			filters = append(filters, f)
		}
	}
	fs.filters = filters
}

func (fs *Filters) appendParsed(s string) error {
	f, err := parse(s)
	if err != nil {
		return err
	}
	fs.filters = append(fs.filters, f)
	return nil
}

func (fs *Filters) Matches(obj ValueGetter) (bool, error) {
	for _, f := range fs.filters {
		match, err := f.matches(obj)
		if err != nil {
			return false, fmt.Errorf("Error applying filter to %s: %s", f.name, err)
		}
		if !match {
			return false, nil
		}
	}
	return true, nil
}

func (f *filter) matches(obj ValueGetter) (bool, error) {
	value := obj.GetValue(f.name)
	if value == nil {
		return f.value.IsNone() && f.operator == OperatorEqual, nil
	}
	if f.value.IsNone() {
		if f.operator == OperatorEqual {
			return value.IsNone(), nil
		}
		return !value.IsNone(), nil
	}

	switch f.operator {
	case OperatorEqual:
		return value.Equal(f.value)
	case OperatorNotEqual:
		return value.NotEqual(f.value)
	case OperatorLessThan:
		return value.LessThan(f.value)
	case OperatorLessOrEqual:
		return value.LessOrEqual(f.value)
	case OperatorGreaterThan:
		return value.GreaterThan(f.value)
	case OperatorGreaterOrEqual:
		return value.GreaterOrEqual(f.value)
	}
	panic("Unknown operator")
}

func parse(s string) (*filter, error) {
	parseErr := fmt.Errorf(`Failed to parse filter: "%s".

Filters must be in the format "<name> <operator> <value>",
where <operator> can be
  "=" (equal),
  "!=" (not equal),
  "<" (less than),
  "<=" (less than or equal),
  ">" (greater than), or
  ">=" (greater than or equal)`, s)

	s = strings.TrimSpace(s)
	matches := parseRegex.FindStringSubmatch(s)
	if len(matches) == 0 {
		return nil, parseErr
	}
	name := matches[1]
	operator := matches[2]
	value := matches[3]

	f := filter{name: name}
	switch operator {
	case "=":
		f.operator = OperatorEqual
	case "!=":
		f.operator = OperatorNotEqual
	case "<":
		f.operator = OperatorLessThan
	case "<=":
		f.operator = OperatorLessOrEqual
	case ">":
		f.operator = OperatorGreaterThan
	case ">=":
		f.operator = OperatorGreaterOrEqual
	default:
		return nil, parseErr
	}

	// TODO(andreas): this is a hack since we don't have a "time" parameter type. we should perhaps add that
	if f.name == "started" {
		t, err := dateparse.ParseLocal(value)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse created time: %s", err)
		}
		f.value = Float(float64(t.Unix()))
	} else {
		f.value = ParseFromString(value)
	}
	return &f, nil
}
