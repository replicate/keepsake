package slices

import (
	"reflect"
)

// ContainsString checks if a []string slice contains a query string
func ContainsString(strings []string, query string) bool {
	for _, s := range strings {
		if s == query {
			return true
		}
	}
	return false
}

// ContainsAnyString checks if an []interface{} slice contains a query string
func ContainsAnyString(strings interface{}, query interface{}) bool {
	return ContainsString(StringSlice(strings), query.(string))
}

// StringSlice converts an []interface{} slice to a []string slice
func StringSlice(strings interface{}) []string {
	if reflect.TypeOf(strings).Kind() != reflect.Slice {
		panic("strings is not a slice")
	}
	ret := []string{}
	vals := reflect.ValueOf(strings)
	for i := 0; i < vals.Len(); i++ {
		ret = append(ret, vals.Index(i).String())
	}
	return ret
}
