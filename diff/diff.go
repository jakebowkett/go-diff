/*
Package diff provides a simple way to diff structs. It
is intended to assist with tasks such as logging changes
to config structs in a running program.
*/
package diff

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"text/template"
)

/*
Structs takes two structs of the same type and finds
the differences between them. Each string in fields
corresponds to a field whose value differs between
the structs. Fields with identical values are omitted.
The format of each string looks like this:

	"{{.Field}} changed from {{.Before}} to {{.After}}"

If the field's value is a string it will be quoted in
the output. Custom formats can be achived with StructsF.

Structs only iterates over the first level of a struct's
fields. It is not intended for structs whose fields are
data structures themselves. Pointers to structs must be
dereferenced when passing them to Structs. Differences
between unexported fields will not be detected.

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
	return structs("{{.Field}} changed from {{.Before}} to {{.After}}", before, after)
}

/*
StructsF works the same as Structs but takes a format string.
Formatting is styled after the standard library's text/template
package. The format string is rendered as a template which has
the following pipelines available to it:

	.Field
	.Before
	.After

These respectively refer to the changed field's name, its previous
value, and its new value. Fields may be omitted if desired.

Returns an error under the same conditions as Structs or if the
supplied format string refers to an unavailable field.

	type Config struct {
		Debug   bool
		Version string
		Timeout int
	}

	c1 := Config{}
	c2 := Config{"0.0.1", true, 30}

	fields, _ := diff.StructsF("{{.Name}}: {{.After}}", c1, c2)
	fmt.Println(fields[0]) // `Debug: true`
	fmt.Println(fields[1]) // `Version: "0.0.1"`
	fmt.Println(fields[2]) // `Timeout: 30`

*/
func StructsF(format string, before, after interface{}) (fields []string, err error) {
	return structs(format, before, after)
}

func structs(format string, before, after interface{}) ([]string, error) {

	if err := valid(before, after); err != nil {
		return nil, err
	}

	t, err := template.New("field").Parse(format)
	if err != nil {
		return nil, err
	}

	s1 := reflect.ValueOf(before)
	s2 := reflect.ValueOf(after)

	var fields []string
	var buf bytes.Buffer

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

		err := t.Execute(&buf, struct {
			Field  string
			Before interface{}
			After  interface{}
		}{
			Field:  reflect.TypeOf(before).Field(i).Name,
			Before: v1,
			After:  v2,
		})
		if err != nil {
			return nil, err
		}

		fields = append(fields, buf.String())
		buf.Reset()
	}

	return fields, nil
}

func valid(before, after interface{}) error {

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
