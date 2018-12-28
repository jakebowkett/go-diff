package diff

import "testing"

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

func TestStructs(t *testing.T) {

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
			true,
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
				`Debug changed from true to false`,
				`Version changed from "0.0.0" to "0.0.1"`,
				`Timeout changed from 30 to 15`,
			},
			false,
		},
	}

	for _, c := range cases {

		errStr := "nil"
		if c.wantErr {
			errStr = "error"
		}

		got, err := Structs(c.before, c.after)
		if !equal(got, c.want) || err == nil && c.wantErr {
			t.Errorf(
				"Structs(%v, %v)\n"+
					"    return %v, %v"+
					"    wanted %v, %v",
				c.before, c.after, got, err, c.want, errStr)
		}
	}
}

func TestStructsF(t *testing.T) {

	cases := []struct {
		before  interface{}
		after   interface{}
		format  string
		want    []string
		wantErr bool
	}{
		// Unavailable field in format string.
		{
			config{true, "0.0.0", 30},
			config{true, "0.0.1", 15},
			`{{.Field}}: {{.Apple}}`,
			nil,
			true,
		},

		// Correctly formatted strings.
		{
			config{true, "0.0.0", 30},
			config{true, "0.0.1", 15},
			`{{.Field}}: {{.After}}`,
			[]string{
				`Version: "0.0.1"`,
				`Timeout: 15`,
			},
			false,
		},
		{
			config{true, "0.0.0", 30},
			config{false, "0.0.1", 0},
			`{{.Before}} --> {{.After}}`,
			[]string{
				`true --> false`,
				`"0.0.0" --> "0.0.1"`,
				`30 --> 0`,
			},
			false,
		},
	}

	for _, c := range cases {

		errStr := "nil"
		if c.wantErr {
			errStr = "error"
		}

		got, err := StructsF(c.format, c.before, c.after)
		if !equal(got, c.want) || err == nil && c.wantErr {
			t.Errorf(
				"Structs(%v, %v)\n"+
					"    return %v, %v"+
					"    wanted %v, %v",
				c.before, c.after, got, err, c.want, errStr)
		}
	}
}

func equal(s1, s2 []string) bool {

	if len(s1) != len(s2) {
		return false
	}

	for i, _ := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}
