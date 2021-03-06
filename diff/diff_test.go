package diff

import (
	"fmt"
	"testing"
)

type config struct {
	Debug   bool
	Version string
	Timeout int
}
type notConfig struct {
	Debug   bool
	Version string
	Timeout int
}
type nestedTest struct {
	Mapping map[string][]string
}
type mapTest struct {
	Mapping map[string]string
}

func TestObjects(t *testing.T) {

	cases := []struct {
		before  interface{}
		after   interface{}
		want    []string
		wantErr bool
	}{

		// Nil object.
		{
			[]string{"hi", "there"},
			nil,
			nil,
			true,
		},

		// General maps.
		{
			map[string]string{
				"yo": "hello",
			},
			map[string]string{
				"hi": "there",
			},
			[]string{
				`["yo"] deleted "hello"`,
				`["hi"] added "there"`,
			},
			false,
		},

		// Nested maps where one is nil.
		{
			nestedTest{},
			nestedTest{
				Mapping: map[string][]string{
					"yo": []string{"hi"},
				},
			},
			[]string{`.Mapping["yo"][0] added "hi"`},
			false,
		},

		// Nested arrays where one is nil.
		{
			nestedTest{
				Mapping: map[string][]string{
					"yo": []string{"hi", "there"},
				},
			},
			nestedTest{
				Mapping: map[string][]string{},
			},
			[]string{
				`.Mapping["yo"][0] deleted "hi"`,
				`.Mapping["yo"][1] deleted "there"`,
			},
			false,
		},

		// Non structs of same type.
		{
			[3]int{1, 2, 3},
			[3]int{1, 2, 3},
			nil,
			false,
		},

		// Non structs of same type with different contents.
		{
			[3]int{1, 2, 3},
			[3]int{1, 2},
			[]string{
				`[2] changed from 3 to 0`,
			},
			false,
		},

		// One non struct.
		{
			config{},
			[3]int{1, 2, 3},
			nil,
			true,
		},

		// Pointers to structs of same type.
		{
			&config{},
			&config{},
			nil,
			true,
		},

		// Structs of different types.
		{
			config{},
			notConfig{},
			nil,
			true,
		},

		// Empty structs of same type.
		{
			config{},
			config{},
			nil,
			false,
		},

		// Filled structs of same type.
		{
			config{true, "0.0.0", 30},
			config{true, "0.0.0", 30},
			nil,
			false,
		},

		// Filled structs of same type, different values.
		{
			config{true, "0.0.0", 30},
			config{false, "0.0.1", 15},
			[]string{
				`.Debug changed from true to false`,
				`.Version changed from "0.0.0" to "0.0.1"`,
				`.Timeout changed from 30 to 15`,
			},
			false,
		},
	}

	for i, c := range cases {

		errStr := "nil"
		if c.wantErr {
			errStr = "error"
		}

		got, err := Objects(c.before, c.after)
		if !equal(got, c.want) || err == nil && c.wantErr {
			fmt.Printf("Case #%d:\n", i+1)
			t.Errorf(
				"Objects(%v, %v)\n"+
					"    return %v, %v\n"+
					"    wanted %v, %v",
				c.before, c.after, got, err, c.want, errStr)
		}
	}
}

func TestObjectsF(t *testing.T) {

	cases := []struct {
		before  interface{}
		after   interface{}
		format  Format
		want    []string
		wantErr bool
	}{
		// Unavailable field in format string.
		{
			config{true, "0.0.0", 30},
			config{true, "0.0.1", 15},
			Format{
				Change: `{{.Name}}: {{.Apple}}`,
				Add:    `{{.Name}}: {{.Apple}}`,
				Delete: `{{.Name}}: {{.Apple}}`,
			},
			nil,
			true,
		},

		// Correctly formatted strings.
		{
			config{true, "0.0.0", 30},
			config{true, "0.0.1", 15},
			Format{
				Change: `{{.Name}}: {{.After}}`,
				Add:    `{{.Name}}: {{.Before}}{{.After}}`,
				Delete: `{{.Name}}: {{.Before}}`,
			},
			[]string{
				`.Version: "0.0.1"`,
				`.Timeout: 15`,
			},
			false,
		},
		{
			config{true, "0.0.0", 30},
			config{false, "0.0.1", 0},
			Format{
				Change: `{{.Before}} --> {{.After}}`,
				Add:    `{{.Before}} --> {{.After}}`,
				Delete: `{{.Before}} --> {{.After}}`,
			},
			[]string{
				`true --> false`,
				`"0.0.0" --> "0.0.1"`,
				`30 --> 0`,
			},
			false,
		},
	}

	for i, c := range cases {

		errStr := "nil"
		if c.wantErr {
			errStr = "error"
		}

		got, err := ObjectsF(c.format, c.before, c.after)
		if !equal(got, c.want) || err == nil && c.wantErr {
			fmt.Printf("Case #%d:\n", i+1)
			t.Errorf(
				"ObjectsF(%v, %v, %v)\n"+
					"    return %v, %v"+
					"    wanted %v, %v",
				c.format, c.before, c.after, got, err, c.want, errStr)
		}
	}
}

func equal(s1, s2 []string) bool {

	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}
