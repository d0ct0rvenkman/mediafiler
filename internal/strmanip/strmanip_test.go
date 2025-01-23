package strmanip

import (
	"testing"
)

func TestStrReplace(t *testing.T) {
	type args struct {
		str     string
		replace map[string]string
	}

	replacements := make(map[string]string)
	replacements["foo"] = "bar"
	replacements["barf"] = "spaz"
	replacements["canonical"] = "ubuntooo"

	tests := []struct {
		name string
		args args
		want string
	}{
		{"foo => bar", args{"foo", replacements}, "bar"},
		{"barf => spaz", args{"barf", replacements}, "spaz"},
		{"canonical => ubuntooo", args{"canonical", replacements}, "ubuntooo"},
		{"empty string", args{"", replacements}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrReplace(tt.args.str, tt.args.replace); got != tt.want {
				t.Errorf("StrReplace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStrtr(t *testing.T) {
	type args struct {
		str     string
		replace map[string]string
	}

	lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	/*
		// The following would be a nice test, but doesn't work because the ordering of the
		// replacements isn't guaranteed within the map. Making Strtr deterministic would be good.

		lorem_multi_replace_1 := "Lodom ipsum dolor sit amet, consectetur adipiscing elit, awk do eiusmod tempor incididunt ut dobodo et dolodo magna aliqua. Ut enim ad donim veniam, quis nostrud exercitadoon uldomco doboris nisi ut aliquip ex ea commodo consequat. Duis aute irudo dolor in dopdohenderit in voluptate velit esse cillum dolodo eu fugiat nuldo pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est doborum."

		replacements_1 := make(map[string]string)
		replacements_1["sed"] = "awk"
		replacements_1["do"] = "re"
		replacements_1["re"] = "mi"
		replacements_1["mi"] = "fa"
		replacements_1["fa"] = "so"
		replacements_1["so"] = "la"
		replacements_1["la"] = "ti"
		replacements_1["ti"] = "do"
	*/

	replacements_2 := make(map[string]string)
	replacements_2["Lorem"] = "Datam"
	replacements_2["adipiscing"] = "a word that surely exists"
	lorem_multi_replace_2 := "Datam ipsum dolor sit amet, consectetur a word that surely exists elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

	replacements_empty := make(map[string]string)

	tests := []struct {
		name string
		args args
		want string
	}{
		// {"test name", args{"", replacements}, ""},
		// {"lorem forward", args{lorem, replacements_1}, lorem_multi_replace_1},
		{"lorem multi-word 2", args{lorem, replacements_2}, lorem_multi_replace_2},
		{"lorem empty replacement", args{lorem, replacements_empty}, lorem},
		{"empty input valid replacement", args{"", replacements_2}, ""},
		{"empty input empty replacement", args{"", replacements_empty}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Strtr(tt.args.str, tt.args.replace); got != tt.want {
				t.Errorf("Strtr() = '%v',\n want '%v'", got, tt.want)
			}
		})
	}
}
