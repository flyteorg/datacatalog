// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package config

import (
	"encoding/json"
	"reflect"

	"fmt"

	"github.com/spf13/pflag"
)

// If v is a pointer, it will get its element value or the zero value of the element type.
// If v is not a pointer, it will return it as is.
func (DbConfigSection) elemValueOrNil(v interface{}) interface{} {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		if reflect.ValueOf(v).IsNil() {
			return reflect.Zero(t.Elem()).Interface()
		} else {
			return reflect.ValueOf(v).Interface()
		}
	} else if v == nil {
		return reflect.Zero(t).Interface()
	}

	return v
}

func (DbConfigSection) mustMarshalJSON(v json.Marshaler) string {
	raw, err := v.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(raw)
}

// GetPFlagSet will return strongly types pflags for all fields in DbConfigSection and its nested types. The format of the
// flags is json-name.json-sub-name... etc.
func (cfg DbConfigSection) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("DbConfigSection", pflag.ExitOnError)
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "host"), *new(string), "")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "port"), *new(int), "")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "dbname"), *new(string), "")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "username"), *new(string), "")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "password"), *new(string), "")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "passwordPath"), *new(string), "")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "options"), *new(string), "")
	return cmdFlags
}