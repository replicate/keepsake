package param

import (
	"strings"
)

type Sorter struct {
	Key        string
	Descending bool
}

func NewSorter(sortString string) *Sorter {
	key := sortString
	desc := false
	if strings.HasSuffix(sortString, "-desc") {
		key = strings.TrimSuffix(sortString, "-desc")
		desc = true
	} else if strings.HasSuffix(sortString, "-asc") {
		key = strings.TrimSuffix(sortString, "-asc")
	}
	return &Sorter{Key: key, Descending: desc}
}

func (s *Sorter) LessThan(x ValueGetter, y ValueGetter) bool {
	xVal := x.GetValue(s.Key)
	yVal := y.GetValue(s.Key)
	var isLess bool
	var err error
	isLess, err = xVal.LessThan(yVal)
	if err != nil {
		panic(err)
	}
	if s.Descending {
		return !isLess
	}
	return isLess
}
