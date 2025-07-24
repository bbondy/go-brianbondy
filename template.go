package main

import (
	"reflect"
)

func avail(name string, data interface{}) bool {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	field := v.FieldByName(name)
	if !field.IsValid() {
		return false
	}

	// Check if the field is a string and not empty
	if field.Kind() == reflect.String {
		return field.String() != ""
	}
	// Return true if the field is not a string but is valid
	return true
}
