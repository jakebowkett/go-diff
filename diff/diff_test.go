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

func TestObjects(t *testing.T) {

	cases := []struct {
		before  interface{}
		after   interface{}
		want    []string
		wantErr bool
	}{

		// Non structs of same type.
		{
			[3]int{1, 2, 3},
			[3]int{1, 2, 3},
			nil,
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
					"    return %v, %v"+
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
