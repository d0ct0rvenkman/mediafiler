package config

import (
	"fmt"
	"testing"
)

func TestPathIgnorePattern_IsValid(t *testing.T) {
	type fields struct {
		Type    string
		Pattern string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"invalid bad type", fields{Type: "footypebar", Pattern: "findme"}, false},
		{"invalid bad type empty Find", fields{Type: "footypebar", Pattern: ""}, false},

		{"valid string replace", fields{Type: "string", Pattern: "findme"}, true},
		{"invalid string replace empty Find", fields{Type: "string", Pattern: ""}, false},

		{"valid regex replace", fields{Type: "regex", Pattern: "^$"}, true},
		{"invalid regex replace bad Find regex", fields{Type: "regex", Pattern: "^(("}, false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PathIgnorePattern{
				Type:    tt.fields.Type,
				Pattern: tt.fields.Pattern,
			}
			got, _ := p.IsValid()
			if got != tt.want {
				t.Errorf("PathIgnorePattern.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIgnoreFilterAddRule(t *testing.T) {
	rule1 := PathIgnorePattern{Type: "string", Pattern: "findme"}         // good
	rule2 := PathIgnorePattern{Type: "string", Pattern: ""}               // bad
	rule3 := PathIgnorePattern{Type: "regex", Pattern: "^$"}              // good
	rule4 := PathIgnorePattern{Type: "regex", Pattern: "^(("}             // bad
	rule5 := PathIgnorePattern{Type: "string", Pattern: "findyou"}        // good
	rule6 := PathIgnorePattern{Type: "string", Pattern: "findagain"}      // good
	rule7 := PathIgnorePattern{Type: "string", Pattern: "findrepeatedly"} // good

	var p PathIgnoreFilter
	if len(p) != 0 {
		t.Errorf("PathIgnorePattern.AddRule(): new replacer isn't empty. huh?")
	}

	p.AddPattern(rule1)
	if l := len(p); l != 1 {
		t.Errorf("PathIgnorePattern.AddRule(): rule count is %v, expected 1", l)
	}

	p.AddPattern(rule2) // should fail to add
	if l := len(p); l != 1 {
		t.Errorf("PathIgnorePattern.AddRule(): rule count is %v, expected 1", l)
	}

	p.AddPattern(rule3)
	if l := len(p); l != 2 {
		t.Errorf("PathIgnorePattern.AddRule(): rule count is %v, expected 2", l)
	}

	p.AddPattern(rule4) // should fail to add
	if l := len(p); l != 2 {
		t.Errorf("PathIgnorePattern.AddRule(): rule count is %v, expected 2", l)
	}

	p.AddPattern(rule5)
	if l := len(p); l != 3 {
		t.Errorf("PathIgnorePattern.AddRule(): rule count is %v, expected 3", l)
	}

	p.AddPattern(rule6)
	if l := len(p); l != 4 {
		t.Errorf("PathIgnorePattern.AddRule(): rule count is %v, expected 4", l)
	}

	p.AddPattern(rule7)
	if l := len(p); l != 5 {
		t.Errorf("PathIgnorePattern.AddRule(): rule count is %v, expected 5", l)
	}

	if p[1] != rule3 {
		t.Errorf("PathIgnorePattern.AddRule(): rule 1 is expected to match test rule 3, but doesn't")
	}

}

func TestPathIgnoreFilter_IsPathFiltered(t *testing.T) {

	var p PathIgnoreFilter
	p.AddPattern(PathIgnorePattern{Type: "string", Pattern: `foo`})
	p.AddPattern(PathIgnorePattern{Type: "string", Pattern: `/dir/`})
	p.AddPattern(PathIgnorePattern{Type: "regex", Pattern: `\s\s`})
	p.AddPattern(PathIgnorePattern{Type: "string", Pattern: `file`})
	p.AddPattern(PathIgnorePattern{Type: "regex", Pattern: `b[a-z]r`})
	p.AddPattern(PathIgnorePattern{Type: "string", Pattern: `ll`})
	p.AddPattern(PathIgnorePattern{Type: "regex", Pattern: `^/file$`})

	tests := []struct {
		testPath string
		want     bool
	}{
		{testPath: "/path/to/foo", want: true},
		{testPath: "/path/to/bar", want: true},
		{testPath: "/path/to/bor", want: true},
		{testPath: "/path/to/b0r", want: false},
		{testPath: "/definitely/not/a/dir", want: false},
		{testPath: "/definitely/a/dir/", want: true},
		{testPath: "/very/much/a/file", want: true},
		{testPath: "/paths/with/spaces  are evil", want: true},
		{testPath: "/llamas/whip/behind", want: true},
		{testPath: "/usr/bin/top", want: false},
		{testPath: "mediafuuler/test/path1/mediafuuler.yaml", want: false},
		{testPath: "mediafiler/test/path1/mediafiler.yaml", want: true},
	}
	for tv, tt := range tests {
		testname := fmt.Sprintf("string-%d", tv)
		t.Run(testname, func(t *testing.T) {
			got, _ := p.IsPathFiltered(tt.testPath)
			if got != tt.want {
				t.Errorf("PathIgnoreFilter.IsPathFiltered() = %v, want %v", got, tt.want)
			}
		})
	}
}
