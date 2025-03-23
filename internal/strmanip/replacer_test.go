package strmanip

import "testing"

/*
This test verfies the IsValid method passes/fails individual rules correctly
*/
func TestReplacerIsValid(t *testing.T) {

	tests := []struct {
		name string
		rule ReplacerRule
		want bool
	}{
		{"invalid bad type", ReplacerRule{Type: "footypebar", Find: "findme", ReplaceWith: "replacewithme"}, false},
		{"invalid bad type empty Find", ReplacerRule{Type: "footypebar", Find: "", ReplaceWith: "replacewithme"}, false},

		{"valid string replace", ReplacerRule{Type: "string", Find: "findme", ReplaceWith: "replacewithme"}, true},
		{"invalid string replace empty Find", ReplacerRule{Type: "string", Find: "", ReplaceWith: "replacewithme"}, false},

		{"valid regex replace", ReplacerRule{Type: "regex", Find: "^$", ReplaceWith: "replacewithme"}, true},
		{"invalid regex replace bad Find regex", ReplacerRule{Type: "regex", Find: "^((", ReplaceWith: "replacewithme"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, e := tt.rule.IsValid(); got != tt.want {
				t.Errorf("ReplacerRule.IsValid() = %v, want %v", got, tt.want)
				t.Logf("e: %v", e)
			}
		})
	}

}

/*
This test verfies that AddRule properly adds individual rules to a Replacer
*/
func TestReplacerAddRule(t *testing.T) {
	rule1 := ReplacerRule{Type: "string", Find: "findme", ReplaceWith: "replacewithme"}                 // good
	rule2 := ReplacerRule{Type: "string", Find: "", ReplaceWith: "replacewithme"}                       // bad
	rule3 := ReplacerRule{Type: "regex", Find: "^$", ReplaceWith: "replacewithme"}                      // good
	rule4 := ReplacerRule{Type: "regex", Find: "^((", ReplaceWith: "replacewithme"}                     // bad
	rule5 := ReplacerRule{Type: "string", Find: "findyou", ReplaceWith: "replacewithyou"}               // good
	rule6 := ReplacerRule{Type: "string", Find: "findagain", ReplaceWith: "replacewithagain"}           // good
	rule7 := ReplacerRule{Type: "string", Find: "findrepeatedly", ReplaceWith: "replacewithrepeatedly"} // good
	rule8 := ReplacerRule{Type: "strang", Find: "findexcessively", ReplaceWith: "replaceexcessively"}   // bad

	var r Replacer
	if len(r) != 0 {
		t.Errorf("ReplacerRule.AddRule(): new replacer isn't empty. huh?")
	}

	r.AddRule(rule1)
	if l := len(r); l != 1 {
		t.Errorf("ReplacerRule.AddRule(): rule count is %v, expected 1", l)
	}

	r.AddRule(rule2) // should fail to add
	if l := len(r); l != 1 {
		t.Errorf("ReplacerRule.AddRule(): rule count is %v, expected 1", l)
	}

	r.AddRule(rule3)
	if l := len(r); l != 2 {
		t.Errorf("ReplacerRule.AddRule(): rule count is %v, expected 2", l)
	}

	r.AddRule(rule4) // should fail to add
	if l := len(r); l != 2 {
		t.Errorf("ReplacerRule.AddRule(): rule count is %v, expected 2", l)
	}

	r.AddRule(rule5)
	if l := len(r); l != 3 {
		t.Errorf("ReplacerRule.AddRule(): rule count is %v, expected 3", l)
	}

	r.AddRule(rule6)
	if l := len(r); l != 4 {
		t.Errorf("ReplacerRule.AddRule(): rule count is %v, expected 4", l)
	}

	r.AddRule(rule7)
	if l := len(r); l != 5 {
		t.Errorf("ReplacerRule.AddRule(): rule count is %v, expected 5", l)
	}

	r.AddRule(rule8)
	if l := len(r); l != 5 {
		t.Errorf("ReplacerRule.AddRule(): rule count is %v, expected 5", l)
	}

	if r[1] != rule3 {
		t.Errorf("ReplacerRule.AddRule(): rule 1 is expected to match test rule 3, but doesn't")
	}

}

/*
This test verifies basic Replacer functionality using string types.
*/
func TestReplacerReplace_StringBasic(t *testing.T) {
	type args struct {
		str     string
		replace Replacer ``
	}

	var r1 Replacer
	var r2 Replacer
	var r3 Replacer
	var r4 Replacer

	// r.AddRule(ReplacerRule{Type: "string", Find: "" , ReplaceWith: ""})

	r1.AddRule(ReplacerRule{Type: "string", Find: " ", ReplaceWith: "SPACE"})
	r1.AddRule(ReplacerRule{Type: "string", Find: "\t", ReplaceWith: "TAB"})
	r1.AddRule(ReplacerRule{Type: "string", Find: "%", ReplaceWith: "PERCENT"})

	r2.AddRule(ReplacerRule{Type: "string", Find: "good", ReplaceWith: "bad"})
	r2.AddRule(ReplacerRule{Type: "string", Find: "line", ReplaceWith: "circle"})
	r2.AddRule(ReplacerRule{Type: "string", Find: "is", ReplaceWith: "is not"})

	r3.AddRule(ReplacerRule{Type: "string", Find: "a", ReplaceWith: "A"})
	r3.AddRule(ReplacerRule{Type: "string", Find: "e", ReplaceWith: "E"})
	r3.AddRule(ReplacerRule{Type: "string", Find: "i", ReplaceWith: "I"})
	r3.AddRule(ReplacerRule{Type: "string", Find: "o", ReplaceWith: "O"})
	r3.AddRule(ReplacerRule{Type: "string", Find: "u", ReplaceWith: "U"})
	r3.AddRule(ReplacerRule{Type: "string", Find: "y", ReplaceWith: "Y"})

	r4.AddRule(ReplacerRule{Type: "string", Find: "superlongstringthatwillnotmatch", ReplaceWith: "whee"})

	s1 := "\t\ta line of text"
	s2 := "some%words@with)characters^for,good=measure"
	s3 := "testing is loads of fun"

	tests := []struct {
		name string
		args args
		want string
	}{
		// {"", args{s, r}, ""},
		{"stringbasic string1 rule1", args{s1, r1}, "TABTABaSPACElineSPACEofSPACEtext"},
		{"stringbasic string1 rule2", args{s1, r2}, "\t\ta circle of text"},
		{"stringbasic string1 rule3", args{s1, r3}, "\t\tA lInE Of tExt"},
		{"stringbasic string1 rule4", args{s1, r4}, s1},

		{"stringbasic string2 rule1", args{s2, r1}, "somePERCENTwords@with)characters^for,good=measure"},
		{"stringbasic string2 rule2", args{s2, r2}, "some%words@with)characters^for,bad=measure"},
		{"stringbasic string2 rule3", args{s2, r3}, "sOmE%wOrds@wIth)chArActErs^fOr,gOOd=mEAsUrE"},
		{"stringbasic string2 rule4", args{s2, r4}, s2},

		{"stringbasic string3 rule1", args{s3, r1}, "testingSPACEisSPACEloadsSPACEofSPACEfun"},
		{"stringbasic string3 rule2", args{s3, r2}, "testing is not loads of fun"},
		{"stringbasic string3 rule3", args{s3, r3}, "tEstIng Is lOAds Of fUn"},
		{"stringbasic string3 rule4", args{s3, r4}, s3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.args.replace.Replace(tt.args.str); got != tt.want {
				t.Errorf("Replacer.Replace() = '%v', want '%v'", got, tt.want)
			}
		})
	}

}

/*
This test verifies basic Replacer functionality using regex types.
*/
func TestReplacerReplace_RegexBasic(t *testing.T) {
	type args struct {
		str     string
		replace Replacer ``
	}

	var r1 Replacer
	var r2 Replacer
	var r3 Replacer
	var r4 Replacer

	// r.AddRule(ReplacerRule{Type: "regex", Find: "" , ReplaceWith: ""})
	r1.AddRule(ReplacerRule{Type: "regex", Find: `^\s`, ReplaceWith: "LEADINGWS"})
	r1.AddRule(ReplacerRule{Type: "regex", Find: `\S\s`, ReplaceWith: "NOTWS-WS"})
	r1.AddRule(ReplacerRule{Type: "regex", Find: `[a-z]$`, ReplaceWith: "ENDLETTER"})

	r2.AddRule(ReplacerRule{Type: "regex", Find: `([aeiou])([aeiou])`, ReplaceWith: "$2$1"})
	r2.AddRule(ReplacerRule{Type: "regex", Find: `[^A-Za-z]`, ReplaceWith: "_"})

	r3.AddRule(ReplacerRule{Type: "regex", Find: `[A-Za-z]`, ReplaceWith: "_"})
	r3.AddRule(ReplacerRule{Type: "regex", Find: `[,^=]`, ReplaceWith: "_"})
	r3.AddRule(ReplacerRule{Type: "regex", Find: `_`, ReplaceWith: ""})

	r4.AddRule(ReplacerRule{Type: "regex", Find: `(.*)$`, ReplaceWith: "fullreplace: $1"})

	s1 := "\t\ta line of text"
	s2 := "some%words@with)characters^for,good=measure"
	s3 := "testing is loads of fun!"

	tests := []struct {
		name string
		args args
		want string
	}{
		// {"", args{s, r}, ""},
		{"regexbasic string1 rule1", args{s1, r1}, "LEADINGWNOTWS-WSNOTWS-WSlinNOTWS-WSoNOTWS-WStexENDLETTER"},
		{"regexbasic string1 rule2", args{s1, r2}, "__a_line_of_text"},
		{"regexbasic string1 rule3", args{s1, r3}, "\t\t   "},
		{"regexbasic string1 rule4", args{s1, r4}, "fullreplace: " + s1},

		{"regexbasic string2 rule1", args{s2, r1}, "some%words@with)characters^for,good=measurENDLETTER"},
		{"regexbasic string2 rule2", args{s2, r2}, "some_words_with_characters_for_good_maesure"},
		{"regexbasic string2 rule3", args{s2, r3}, "%@)"},
		{"regexbasic string2 rule4", args{s2, r4}, "fullreplace: " + s2},

		{"regexbasic string3 rule1", args{s3, r1}, "testinNOTWS-WSiNOTWS-WSloadNOTWS-WSoNOTWS-WSfun!"},
		{"regexbasic string3 rule2", args{s3, r2}, "testing_is_laods_of_fun_"},
		{"regexbasic string3 rule3", args{s3, r3}, "    !"},
		{"regexbasic string3 rule4", args{s3, r4}, "fullreplace: " + s3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.args.replace.Replace(tt.args.str); got != tt.want {
				t.Errorf("Replacer.Replace() = '%v', want '%v'", got, tt.want)
			}
		})
	}

}

/*
This test verifies a more complicated Replacer rule set using both string and regex types
and ordered replacements.
*/
func TestReplacerReplace_Complex(t *testing.T) {
	type args struct {
		str     string
		replace Replacer ``
	}

	var r1 Replacer
	var r2 Replacer

	r1.AddRule(ReplacerRule{Type: "string", Find: "sed", ReplaceWith: "awk"})
	r1.AddRule(ReplacerRule{Type: "string", Find: "do", ReplaceWith: "re"})
	r1.AddRule(ReplacerRule{Type: "string", Find: "re", ReplaceWith: "mi"})
	r1.AddRule(ReplacerRule{Type: "string", Find: "mi", ReplaceWith: "fa"})
	r1.AddRule(ReplacerRule{Type: "string", Find: "fa", ReplaceWith: "so"})
	r1.AddRule(ReplacerRule{Type: "string", Find: "so", ReplaceWith: "la"})
	r1.AddRule(ReplacerRule{Type: "string", Find: "la", ReplaceWith: "ti"})
	r1.AddRule(ReplacerRule{Type: "string", Find: "ti", ReplaceWith: "do"})
	r1.AddRule(ReplacerRule{Type: "regex", Find: " ([a-z])[a-z]{3}([ ,.])", ReplaceWith: " $1***$2"})

	r2.AddRule(ReplacerRule{Type: "string", Find: "ti", ReplaceWith: "do"})
	r2.AddRule(ReplacerRule{Type: "string", Find: "la", ReplaceWith: "ti"})
	r2.AddRule(ReplacerRule{Type: "string", Find: "so", ReplaceWith: "la"})
	r2.AddRule(ReplacerRule{Type: "string", Find: "fa", ReplaceWith: "so"})
	r2.AddRule(ReplacerRule{Type: "string", Find: "mi", ReplaceWith: "fa"})
	r2.AddRule(ReplacerRule{Type: "string", Find: "re", ReplaceWith: "mi"})
	r2.AddRule(ReplacerRule{Type: "string", Find: "do", ReplaceWith: "re"})
	r2.AddRule(ReplacerRule{Type: "string", Find: "sed", ReplaceWith: "awk"})
	r2.AddRule(ReplacerRule{Type: "regex", Find: " ([a-z])[a-z]{3}([ ,.])", ReplaceWith: " $1***$2"})

	lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

	lorem_after_r1 := "Lodom ipsum dolor sit a***, consectetur adipiscing e***, awk do eiusmod tempor incididunt ut dobodo et dolodo magna aliqua. Ut e*** ad donim veniam, q*** nostrud exercitadoon uldomco doboris n*** ut aliquip ex ea commodo consequat. Duis a*** irudo dolor in dopdohenderit in voluptate velit e*** cillum dolodo eu fugiat nuldo pariatur. Excepteur s*** occaecat cupidatat non proident, s*** in culpa qui officia deserunt mollit a*** id est doborum."
	lorem_after_r2 := "Lomim ipsum relor sit a***, consectetur adipiscing e***, awk re eiusmod tempor incididunt ut tibomi et relomi magna aliqua. Ut e*** ad fanim veniam, q*** nostrud exercitareon ultimco tiboris n*** ut aliquip ex ea commore consequat. Duis a*** irumi relor in mipmihenderit in voluptate velit e*** cillum relomi eu fugiat nulti pariatur. Excepteur s*** occaecat cupidatat non proident, s*** in culpa qui officia deserunt mollit a*** id est tiborum."

	tests := []struct {
		name string
		args args
		want string
	}{
		{"complex lorem rule1", args{lorem, r1}, lorem_after_r1},
		{"complex lorem rule2", args{lorem, r2}, lorem_after_r2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.args.replace.Replace(tt.args.str); got != tt.want {
				t.Errorf("Replacer.Replace() = '%v', want '%v'", got, tt.want)
			}
		})
	}

}
