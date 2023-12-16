package amaro

import (
	"reflect"
	"testing"
)

var tests = []struct {
	input string
	err   string
	want  map[string]arg
}{
	{
		input: "--name=\"John Doe\"",
		want: map[string]arg{
			"name": {
				name:  "name",
				value: "John Doe",
			},
		},
	},
	{
		input: "--name=\"John Doe\" --age 30",
		want: map[string]arg{
			"name": {
				name:  "name",
				value: "John Doe",
			},
			"age": {
				name:  "age",
				value: "30",
			},
		},
	},
	{
		input: "--important",
		want: map[string]arg{
			"important": {
				name:  "important",
				value: "",
			},
		},
	},
	{
		input: "--important --name=\"John Doe",
		want: map[string]arg{
			"important": {
				name:  "important",
				value: "",
			},
			"name": {
				name:  "name",
				value: "John Doe",
			},
		},
	},
	{
		input: " 0",
		err:   "expected an argument, got \"0\"",
	},
	{
		input: "000",
		err:   "expected an argument, got \"000\"",
	},
	{
		input: "-- ",
		want: map[string]arg{
			"*": {
				name:  "*",
				value: " ",
			},
		},
	},
	{
		input: "--confirm ",
		want: map[string]arg{
			"confirm": {
				name:  "confirm",
				value: "",
			},
		},
	},
	{
		input: "--confirm -",
		want: map[string]arg{
			"confirm": {
				name:  "confirm",
				value: "-",
			},
		},
	},
}

func TestParse(t *testing.T) {
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got, err := parse(test.input)

			if test.err != "" {
				if err == nil {
					t.Errorf("parse(%q) = %q, want %q", test.input, got, test.want)
					return
				}
				if err.Error() != test.err {
					t.Errorf("parse(%q) = %q, want %q", test.input, err, test.err)
				}

				return
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("parse(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func FuzzParse(f *testing.F) {
	for _, tc := range tests {
		f.Add(tc.input) // Use f.Add to provide a seed corpus
	}
	f.Fuzz(func(t *testing.T, orig string) {
		_, _ = parse(orig)
	})
}
