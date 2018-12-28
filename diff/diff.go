/*
Package diff provides a simple way to diff structs. It
is intended to assist with tasks such as logging changes
to config structs in a running program.
*/
package diff

import (
	"errors"
	"fmt"
	"reflect"
)

/*
Structs takes two structs of any type and finds the
differences between them. The fields return value is
a slice where each string corresponds to a field that
has changed. Unchanged fields are omitted. Its format
looks like this and string values will be quoted.

	"[field name] changed from [previous value] to [new value]"

Structs only iterates over the first level of a struct's
fields. It is not intended for structs whose fields are
data structures themselves. Pointers to structs must be
dereferenced when passing them to Structs. Changes to
unexported fields will not be detected.

If before or after are not structs or they are structs
of different types an error will be returned.

	type Config struct {
		Debug   bool
		Version string
		Timeout int
	}

	c1 := Config{}
	c2 := Config{"0.0.1", true, 30}

	fields, _ := diff.Structs(c1, c2)
	fmt.Println(fields[0]) // `Debug changed from false to true`
	fmt.Println(fields[1]) // `Version changed from "" to "0.0.1"`
	fmt.Println(fields[2]) // `Timeout changed from 0 to 30`
*/
func Structs(before, after interface{}) (fields []string, err error) {

	if err := sameType(before, after); err != nil {
		return nil, err
	}

	s1 := reflect.ValueOf(before)
	s2 := reflect.ValueOf(after)

	for i := 0; i < s1.NumField(); i++ {

		v1 := s1.Field(i).Interface()
		v2 := s2.Field(i).Interface()

		if reflect.DeepEqual(v1, v2) {
			continue
		}

		if _, ok := v1.(string); ok {
			v1 = fmt.Sprintf("%q", v1)
			v2 = fmt.Sprintf("%q", v2)
		}

		fields = append(fields, fmt.Sprintf(
			"%s changed from %v to %v",
			reflect.TypeOf(before).Field(i).Name, v1, v2))
	}

	return fields, nil
}

func sameType(before, after interface{}) error {

	s1 := reflect.TypeOf(before)
	s2 := reflect.TypeOf(after)

	if s1.Kind().String() != "struct" {
		return errors.New(`argument "before" must be struct`)
	}
	if s2.Kind().String() != "struct" {
		return errors.New(`argument "after" must be struct`)
	}

	if s1.Name() != s2.Name() {
		return errors.New("structs are of different types")
	}

	return nil
}
