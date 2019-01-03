/*
Package diff provides formatted diffs for objects.

	type Config struct {
		Debug   bool
		Path    string
		Timeout int
	}

	c1 := Config{}
	c2 := Config{Timeout: 30}

	changes, _ := diff.Objects(c1, c2)
	fmt.Println(changes[0]) // ".Timeout changed from 0 to 30"

	// Arbitrary formatting of output string.
	format := Format{
		Change: "{{.Before}} --> {{.After}} ({{.Name}})"
	}
	changes, _ = diff.ObjectsF(format, c1, c2)
	fmt.Println(changes[0]) // "0 --> 30 (.Timeout)"

*/
package diff

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

/*
Format contains strings that will be passed to the
standard library's text/template package. The only
available arguments are ".Name", ".Before" and ".After"
*/
type Format struct {
	Change string
	Add    string
	Delete string
}

const msgChange = "{{.Name}} changed from {{.Before}} to {{.After}}"
const msgAdd = "{{.Name}} added {{.After}}"
const msgDelete = "{{.Name}} deleted {{.Before}}"

func Objects(before, after interface{}) (changes []string, err error) {
	return objects(Format{
		Change: msgChange,
		Add:    msgAdd,
		Delete: msgDelete,
	}, before, after)
}

func ObjectsF(format Format, before, after interface{}) (changes []string, err error) {
	if format.Change == "" {
		format.Change = msgChange
	}
	if format.Add == "" {
		format.Add = msgAdd
	}
	if format.Delete == "" {
		format.Delete = msgDelete
	}
	return objects(format, before, after)
}

func objects(format Format, before, after interface{}) (changes []string, err error) {

	t1 := reflect.TypeOf(before)
	t2 := reflect.TypeOf(after)

	if err := isObj(t1, t2); err != nil {
		return nil, err
	}
	if err := sameKind(t1, t2); err != nil {
		return nil, err
	}
	if err := sameNamedType(t1, t2); err != nil {
		return nil, err
	}

	v1 := reflect.ValueOf(before)
	v2 := reflect.ValueOf(after)

	t, err := template.New("change").Parse(format.Change)
	if err != nil {
		return nil, err
	}
	t, err = t.New("add").Parse(format.Add)
	if err != nil {
		return nil, err
	}
	t, err = t.New("delete").Parse(format.Delete)
	if err != nil {
		return nil, err
	}

	d := differ{templates: t}
	err = d.diff(&v1, &v2)
	if err != nil {
		return nil, err
	}

	return d.changes, nil
}

type differ struct {
	changes   []string
	path      []string
	templates *template.Template
}

func (d *differ) popPath() {
	if len(d.path) == 0 {
		return
	}
	d.path = d.path[0 : len(d.path)-1]
}

func (d *differ) diff(v1, v2 *reflect.Value) error {

	var kind string
	if v1 == nil {
		kind = v2.Kind().String()
	} else {
		kind = v1.Kind().String()
	}

	var err error

	switch kind {
	case "struct":
		err = d.diffStruct(v1, v2)
	case "map":
		err = d.diffMap(v1, v2)
	case "array", "slice":
		err = d.diffSequence(v1, v2)
	default:
		err = d.diffAtom(v1, v2)
	}

	if err != nil {
		return err
	}

	return nil
}

func (d *differ) diffStruct(v1, v2 *reflect.Value) error {

	var fields int
	if v1 == nil {
		fields = v2.NumField()
	} else {
		fields = v1.NumField()
	}

	for i := 0; i < fields; i++ {

		var name string
		var field1 *reflect.Value
		var field2 *reflect.Value

		switch {
		case v1 == nil:
			name = reflect.TypeOf(v2.Interface()).Field(i).Name
			field1 = nil
			f2 := v2.Field(i)
			field2 = &f2
		case v2 == nil:
			name = reflect.TypeOf(v1.Interface()).Field(i).Name
			f1 := v1.Field(i)
			field1 = &f1
			field2 = nil
		default:
			name = reflect.TypeOf(v1.Interface()).Field(i).Name
			f1 := v1.Field(i)
			f2 := v2.Field(i)
			field1 = &f1
			field2 = &f2
		}

		d.path = append(d.path, "."+name)
		err := d.diff(field1, field2)
		if err != nil {
			return err
		}
		d.popPath()
	}

	return nil
}

func (d *differ) diffSequence(v1, v2 *reflect.Value) error {

	longest := v1.Len()
	if v2.Len() > longest {
		longest = v2.Len()
	}

	for i := 0; i < longest; i++ {

		var elem1 *reflect.Value
		var elem2 *reflect.Value

		switch {
		case i > v1.Len()-1:
			elem1 = nil
			e2 := v2.Index(i)
			elem2 = &e2
		case i > v2.Len()-1:
			e1 := v1.Index(i)
			elem1 = &e1
			elem2 = nil
		default:
			e1 := v1.Index(i)
			e2 := v2.Index(i)
			elem1 = &e1
			elem2 = &e2
		}

		d.path = append(d.path, fmt.Sprintf("[%d]", i))
		err := d.diff(elem1, elem2)
		if err != nil {
			return err
		}
		d.popPath()
	}

	return nil
}

func (d *differ) diffMap(v1, v2 *reflect.Value) error {

	m := alignMapKeys(v1, v2)

	for k, ok := range m {

		var elem1 *reflect.Value
		var elem2 *reflect.Value

		switch {
		case !ok.before:
			elem1 = nil
			e2 := v2.MapIndex(reflect.ValueOf(k))
			elem2 = &e2
		case !ok.after:
			e1 := v1.MapIndex(reflect.ValueOf(k))
			elem1 = &e1
			elem2 = nil
		default:
			e1 := v1.MapIndex(reflect.ValueOf(k))
			e2 := v2.MapIndex(reflect.ValueOf(k))
			elem1 = &e1
			elem2 = &e2
		}

		key := formatInterface(k)
		d.path = append(d.path, fmt.Sprintf("[%v]", key))
		err := d.diff(elem1, elem2)
		if err != nil {
			return err
		}
		d.popPath()
	}

	return nil
}

type val struct {
	before bool
	after  bool
}

func alignMapKeys(m1, m2 *reflect.Value) map[interface{}]val {

	k1 := m1.MapKeys()
	k2 := m2.MapKeys()

	m := make(map[interface{}]val, len(k1)*(len(k2)/2))

	for _, k := range k1 {
		m[k.Interface()] = val{before: true}
	}
	for _, k := range k2 {
		v := m[k.Interface()]
		v.after = true
		m[k.Interface()] = v
	}

	return m
}

func (d *differ) diffAtom(v1, v2 *reflect.Value) error {

	s := struct {
		Name   string
		Before interface{}
		After  interface{}
	}{
		Name: strings.Join(d.path, ""),
	}

	var tmplName string

	switch {
	case v1 == nil:
		tmplName = "add"
		s.After = formatInterface(v2.Interface())
		s.Before = ""
	case v2 == nil:
		tmplName = "delete"
		s.Before = formatInterface(v1.Interface())
		s.After = ""
	case v1.Interface() != v2.Interface():
		tmplName = "change"
		s.Before = formatInterface(v1.Interface())
		s.After = formatInterface(v2.Interface())
	default:
		return nil
	}

	return d.render(tmplName, s)
}

func (d *differ) render(tmplName string, data interface{}) error {
	var buf bytes.Buffer
	err := d.templates.Lookup(tmplName).Execute(&buf, data)
	if err != nil {
		return err
	}
	d.changes = append(d.changes, buf.String())
	return nil
}

func formatInterface(i interface{}) interface{} {
	if s, ok := i.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return i
}

func sameNamedType(t1, t2 reflect.Type) error {
	if t1.Name() != t2.Name() {
		return errors.New(fmt.Sprintf(
			`objects must be same type - "before" was %s, "after" was %s`,
			t1.Name(), t2.Name()))
	}
	return nil
}

func sameKind(t1, t2 reflect.Type) error {
	kind1 := t1.Kind().String()
	kind2 := t2.Kind().String()
	if kind1 != kind2 {
		return errors.New(fmt.Sprintf(
			`objects must be same kind - "before" was %s, "after" was %s`,
			kind1, kind2))
	}
	return nil
}

var objectKinds = []string{"struct", "array", "slice", "map"}

func isObj(t1, t2 reflect.Type) error {

	if kind := t1.Kind().String(); !in(objectKinds, kind) {
		return errors.New(fmt.Sprintf(
			`argument "before" was of kind %q, wanted kind %s`,
			kind, quotedList(objectKinds, "or")))
	}

	if kind := t2.Kind().String(); !in(objectKinds, kind) {
		return errors.New(fmt.Sprintf(
			`argument "after" was of kind %q, wanted kind %s`,
			kind, quotedList(objectKinds, "or")))
	}

	return nil
}

func in(ss []string, s string) bool {
	for _, item := range ss {
		if item == s {
			return true
		}
	}
	return false
}

func quotedList(ss []string, lastPrefix string) (list string) {

	for i, v := range ss {

		if i == len(ss)-1 {
			list += lastPrefix + " "
		}

		list += `"` + v + `"`

		if i != len(ss)-1 {
			list += ", "
		}
	}

	return list
}
