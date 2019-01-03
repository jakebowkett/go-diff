package diff

import (
	"errors"
	"fmt"
	"reflect"
	"text/template"
)

func Objects(before, after interface{}) (changes []string, err error) {

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

	d := differ{}
	d.diff("", &v1, &v2)

	return d.changes, nil
}

type differ struct {
	changes  []string
	template *template.Template
}

func (d *differ) diff(path string, v1, v2 *reflect.Value) {

	var kind string
	if v1 == nil {
		kind = v2.Kind().String()
	} else {
		kind = v1.Kind().String()
	}

	switch kind {
	case "struct":
		d.diffStruct(path, v1, v2)
	case "map":
		d.diffMap(path, v1, v2)
	case "array", "slice":
		d.diffSequence(path, v1, v2)
	default:
		d.diffAtom(path, v1, v2)
	}
}

func (d *differ) diffStruct(path string, v1, v2 *reflect.Value) {

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

		p := path + "." + name

		d.diff(p, field1, field2)
	}
}

func (d *differ) diffSequence(path string, v1, v2 *reflect.Value) {

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

		d.diff(fmt.Sprintf("%s[%d]", path, i), elem1, elem2)
	}
}

func (d *differ) diffMap(path string, v1, v2 *reflect.Value) {

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
		d.diff(fmt.Sprintf("%s[%v]", path, key), elem1, elem2)
	}
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

func (d *differ) diffAtom(path string, v1, v2 *reflect.Value) {

	if v1 == nil {
		d.changes = append(
			d.changes,
			fmt.Sprintf(
				"%s added %v",
				path, formatInterface(v2.Interface())),
		)
		return
	}

	if v2 == nil {
		d.changes = append(
			d.changes,
			fmt.Sprintf(
				"%s deleted %v",
				path, formatInterface(v1.Interface())),
		)
		return
	}

	if !reflect.DeepEqual(v1.Interface(), v2.Interface()) {
		d.changes = append(
			d.changes,
			fmt.Sprintf(
				"%s changed from %v to %v",
				path,
				formatInterface(v1.Interface()),
				formatInterface(v2.Interface())),
		)
		return
	}
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

var objects = []string{"struct", "array", "slice", "map"}

func isObj(t1, t2 reflect.Type) error {

	if kind := t1.Kind().String(); !in(objects, kind) {
		return errors.New(fmt.Sprintf(
			`argument "before" was of kind %q, wanted kind %s`,
			kind, quotedList(objects, "or")))
	}

	if kind := t2.Kind().String(); !in(objects, kind) {
		return errors.New(fmt.Sprintf(
			`argument "after" was of kind %q, wanted kind %s`,
			kind, quotedList(objects, "or")))
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
