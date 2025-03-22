package config

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/d0ct0rvenkman/mediafiler/internal/strmanip"
)

type cli_args []string

func Test_CommandLine_Basic(t *testing.T) {
	testNameSlug := "commandlinebasic-"
	// t.Run("",  func(t *testing.T) {})

	t.Run("empty command line", func(t *testing.T) {
		var args cli_args

		Initialize(args)

		if Config.GetBool("dry-run") != false {
			t.Error("dry-run is not false")
		}

		if Config.GetBool("debug") != false {
			t.Error("debug is not false")
		}

		if Config.GetString("config-file") != "" {
			t.Error("config-file is not empty")
		}

		if Config.GetBool("dump-example-config") != false {
			t.Error("dump-example-config is not false")
		}

		if Config.GetString("exiftool-binary") != "" {
			t.Error("exiftool-binary is not empty")
		}
	})

	boolTests := []struct {
		name  string
		args  cli_args
		key   string
		found bool
		want  bool
	}{
		{"dry-run true implicit", cli_args{"--dry-run"}, "dry-run", true, true},
		{"dry-run true explicit", cli_args{"--dry-run=true"}, "dry-run", true, true},
		{"dry-run false explicit", cli_args{"--dry-run=false"}, "dry-run", true, false},
		{"dry-run nonsense", cli_args{"--dry-run=nonsense"}, "dry-run", false, false},

		{"debug true implicit", cli_args{"--debug"}, "debug", true, true},
		{"debug true explicit", cli_args{"--debug=true"}, "debug", true, true},
		{"debug false explicit", cli_args{"--debug=false"}, "debug", true, false},
		{"debug nonsense", cli_args{"--debug=nonsense"}, "debug", false, false},

		{"dump-example-config true implicit", cli_args{"--dump-example-config"}, "dump-example-config", true, true},
		{"dump-example-config true explicit", cli_args{"--dump-example-config=true"}, "dump-example-config", true, true},
		{"dump-example-config false explicit", cli_args{"--dump-example-config=false"}, "dump-example-config", true, false},
		{"dump-example-config nonsense", cli_args{"--dump-example-config=nonsense"}, "dump-example-config", false, false},

		{"use-default-config true implicit", cli_args{"--use-default-config"}, "use-default-config", true, true},
		{"use-default-config true explicit", cli_args{"--use-default-config=true"}, "use-default-config", true, true},
		{"use-default-config false explicit", cli_args{"--use-default-config=false"}, "use-default-config", true, false},
		{"use-default-config nonsense", cli_args{"--use-default-config=nonsense"}, "use-default-config", false, false},
	}
	for _, v := range boolTests {
		t.Run(testNameSlug+"flag_"+v.name, func(t *testing.T) {
			Initialize(v.args)

			if found := Config.IsSet(v.key); found != v.found {
				t.Errorf(v.key+" found %v, expected %v", found, v.found)
			}

			if got := Config.GetBool(v.key); got != v.want {
				t.Errorf(v.key+" is %v, wanted %v", got, v.want)
			}
		})

	}

	stringTests := []struct {
		name  string
		args  cli_args
		key   string
		found bool
		want  string
	}{
		{"config-file incomplete", cli_args{"--config-file"}, "config-file", false, ""},
		{"config-file equals string", cli_args{"--config-file=nonsense"}, "config-file", true, "nonsense"},
		{"config-file space string", cli_args{"--config-file", "nonsense"}, "config-file", true, "nonsense"},

		{"exiftool-binary incomplete", cli_args{"--exiftool-binary"}, "exiftool-binary", false, ""},
		{"exiftool-binary equals string", cli_args{"--exiftool-binary=nonsense"}, "exiftool-binary", true, "nonsense"},
		{"exiftool-binary space string", cli_args{"--exiftool-binary", "nonsense"}, "exiftool-binary", true, "nonsense"},
	}
	for _, v := range stringTests {
		t.Run(testNameSlug+"flag_"+v.name, func(t *testing.T) {
			Initialize(v.args)

			if found := Config.IsSet(v.key); found != v.found {
				t.Errorf(v.key+" found %v, expected %v", found, v.found)
			}

			if got := Config.GetString(v.key); got != v.want {
				t.Errorf(v.key+" is %v, wanted %v", got, v.want)
			}
		})
	}

}

func Test_DefaultConfigYAML(t *testing.T) {
	var args cli_args

	testNameSlug := "defaultconfig-"

	Initialize(args)
	if err := ApplyDefaultConfiguration(); err != nil {
		t.Fatalf(testNameSlug+"ApplyDefaultConfiguration() failed: reason: %s", err)
	}

	if err := ProcessConfiguration(); err != nil {
		t.Fatalf(testNameSlug+"ProcessConfiguration() failed: reason: %s", err)
	}

	var exp_model_replacer strmanip.Replacer
	var exp_path_filter PathIgnoreFilter

	boolTests := []struct {
		name  string
		key   string
		found bool
		want  bool
	}{
		{"dry-run present+false", "dry-run", true, false},
		{"debug present+false", "debug", true, false},
		{"dump-example-config absent+false", "dump-example-config", false, false},
	}
	for _, v := range boolTests {
		t.Run(testNameSlug+"flag_"+v.name, func(t *testing.T) {

			if found := Config.IsSet(v.key); found != v.found {
				t.Errorf(v.key+" found %v, expected %v", found, v.found)
			}

			if got := Config.GetBool(v.key); got != v.want {
				t.Errorf(v.key+" is %v, wanted %v", got, v.want)
			}
		})

	}

	stringTests := []struct {
		name  string
		key   string
		found bool
		want  string
	}{
		{"config-file absent+empty", "config-file", false, ""},
		{"exiftool-binary present+specified", "exiftool-binary", true, "/usr/bin/exiftool"},
	}
	for _, v := range stringTests {
		t.Run(testNameSlug+"flag_"+v.name, func(t *testing.T) {

			if found := Config.IsSet(v.key); found != v.found {
				t.Errorf(v.key+" found %v, expected %v", found, v.found)
			}

			if got := Config.GetString(v.key); got != v.want {
				t.Errorf(v.key+" is %v, wanted %v", got, v.want)
			}
		})
	}

	exp_model_replacer.AddRule(strmanip.ReplacerRule{Type: "string", Find: "FooBarMatic", ReplaceWith: "FBM"})
	exp_model_replacer.AddRule(strmanip.ReplacerRule{Type: "regex", Find: `\s+`, ReplaceWith: ""})

	exp_path_filter.AddPattern(PathIgnorePattern{Type: "string", Pattern: `.git/`})
	exp_path_filter.AddPattern(PathIgnorePattern{Type: "string", Pattern: `.git\`})
	exp_path_filter.AddPattern(PathIgnorePattern{Type: "regex", Pattern: `^.*[Ii][Cc][Oo]$`})

	t.Run(testNameSlug+"model-replacer-length", func(t *testing.T) {
		// we can't compare these directly, so we'll compare their lengths and then their individual rules

		if len(exp_model_replacer) != len(ModelReplacer) {
			t.Error("model replacer rules loaded from config are different from expected rules (rule count)")
		}

	})
	t.Run(testNameSlug+"model-replacer-contents", func(t *testing.T) {
		for compare_idx := range exp_model_replacer {
			if exp_model_replacer[compare_idx] != ModelReplacer[compare_idx] {
				t.Error("model replacer rules loaded from config are different from expected rules (specifc rule)")

			}
		}
	})

	// we can't compare these directly, so we'll compare their lengths and then their individual patterns
	if len(exp_path_filter) != len(PathIgnorer) {
		t.Error("path filter patterns loaded from config are different from expected patterns (rule count)")
	}

	for compare_idx := range exp_path_filter {
		if exp_path_filter[compare_idx] != PathIgnorer[compare_idx] {
			t.Error("path filter patterns loaded from config are different from expected patterns (specifc rule)")

		}
	}

}

func Test_PathSearch_1(t *testing.T) {
	var args cli_args

	testNameSlug := "pathsearch1-"

	Initialize(args)
	useSpecificConfigPath("../../test/config/path1")
	useSpecificConfigPath("../../test/config/path2")
	useSpecificConfigPath("../../test/config/path3")

	t.Run(testNameSlug+"readconfiguration", func(t *testing.T) {
		cfgRead, err := ReadConfiguration()
		if !cfgRead || (err != nil) {
			t.Fatalf(testNameSlug+"ReadConfiguration() failed: reason: %s", err)
		}
	})

	t.Run(testNameSlug+"processconfiguration", func(t *testing.T) {
		if err := ProcessConfiguration(); err != nil {
			t.Fatalf(testNameSlug+"ProcessConfiguration() failed: reason: %s", err)
		}
	})

	t.Run(testNameSlug+"configfileused", func(t *testing.T) {
		pathparts := strings.Split(Config.ConfigFileUsed(), string(os.PathSeparator))

		// only look at the last parts of the file path that exist within this source repo
		subpath := pathparts[len(pathparts)-4:]
		expected := []string{"test", "config", "path1", "mediafiler.yaml"}

		if !slices.Equal(expected, subpath) {
			t.Errorf("config file used did not match the expected path. got '%v', wanted '%v'", subpath, expected)
		}

	})

	t.Run(testNameSlug+"dry-run", func(t *testing.T) {
		if !(Config.IsSet("dry-run") && Config.GetBool("dry-run")) {
			t.Errorf("dry-run setting is not as expected (explicit true)")
		}
	})

	t.Run(testNameSlug+"debug", func(t *testing.T) {
		if Config.IsSet("debug") {
			t.Errorf("debug setting is not as expected (unset)")
		}
	})

	t.Run(testNameSlug+"exiftool-binary", func(t *testing.T) {
		want := "/usr/bin/path1/exiftool"
		if !(Config.IsSet("exiftool-binary") && (Config.GetString("exiftool-binary") == want)) {
			t.Errorf("exiftool-binary setting is not as expected (explicit %s)", want)
		}
	})

}

func Test_PathSearch_2(t *testing.T) {
	var args cli_args

	testNameSlug := "pathsearch2-"

	Initialize(args)
	useSpecificConfigPath("../../test/config/path3")
	useSpecificConfigPath("../../test/config/path2")
	useSpecificConfigPath("../../test/config/path1")

	t.Run(testNameSlug+"readconfiguration", func(t *testing.T) {
		cfgRead, err := ReadConfiguration()
		if !cfgRead || (err != nil) {
			t.Fatalf(testNameSlug+"ReadConfiguration() failed: reason: %s", err)
		}
	})

	t.Run(testNameSlug+"processconfiguration", func(t *testing.T) {
		if err := ProcessConfiguration(); err != nil {
			t.Fatalf(testNameSlug+"ProcessConfiguration() failed: reason: %s", err)
		}
	})

	t.Run(testNameSlug+"configfileused", func(t *testing.T) {
		pathparts := strings.Split(Config.ConfigFileUsed(), string(os.PathSeparator))

		// only look at the last parts of the file path that exist within this source repo
		subpath := pathparts[len(pathparts)-4:]
		expected := []string{"test", "config", "path2", "mediafiler.yaml"}

		if !slices.Equal(expected, subpath) {
			t.Errorf("config file used did not match the expected path. got '%v', wanted '%v'", subpath, expected)
		}

	})

	t.Run(testNameSlug+"debug", func(t *testing.T) {
		if !(Config.IsSet("debug") && Config.GetBool("debug")) {
			t.Errorf("debug setting is not as expected (explicit true)")
		}
	})

	t.Run(testNameSlug+"dry-run", func(t *testing.T) {
		if Config.IsSet("dry-run") {
			t.Errorf("dry-run setting is not as expected (unset)")
		}
	})

	t.Run(testNameSlug+"exiftool-binary", func(t *testing.T) {
		want := "/usr/bin/path2/exiftool"
		if !(Config.IsSet("exiftool-binary") && (Config.GetString("exiftool-binary") == want)) {
			t.Errorf("exiftool-binary setting is not as expected (explicit %s)", want)
		}
	})

}

func Test_SpecificCfgFile_Good(t *testing.T) {
	var args cli_args
	testNameSlug := "specificcfg-good-"

	args = append(args, "--config-file=../../test/config/path1/mediafiler.yaml")

	Initialize(args)

	t.Run(testNameSlug+"readconfiguration", func(t *testing.T) {
		cfgRead, err := ReadConfiguration()
		if !cfgRead || (err != nil) {
			t.Fatalf(testNameSlug+"ReadConfiguration() failed: reason: %s", err)
		}
	})

	t.Run(testNameSlug+"processconfiguration", func(t *testing.T) {
		if err := ProcessConfiguration(); err != nil {
			t.Fatalf(testNameSlug+"ProcessConfiguration() failed: reason: %s", err)
		}
	})

	t.Run(testNameSlug+"configfileused", func(t *testing.T) {
		pathparts := strings.Split(Config.ConfigFileUsed(), string(os.PathSeparator))

		// only look at the last parts of the file path that exist within this source repo
		subpath := pathparts[len(pathparts)-4:]
		expected := []string{"test", "config", "path1", "mediafiler.yaml"}

		if !slices.Equal(expected, subpath) {
			t.Errorf("config file used did not match the expected path. got '%v', wanted '%v'", subpath, expected)
		}

	})

	t.Run(testNameSlug+"dry-run", func(t *testing.T) {
		if !(Config.IsSet("dry-run") && Config.GetBool("dry-run")) {
			t.Errorf("dry-run setting is not as expected (explicit true)")
		}
	})

	t.Run(testNameSlug+"debug", func(t *testing.T) {
		if Config.IsSet("debug") {
			t.Errorf("debug setting is not as expected (unset)")
		}
	})

	t.Run(testNameSlug+"exiftool-binary", func(t *testing.T) {
		want := "/usr/bin/path1/exiftool"
		if !(Config.IsSet("exiftool-binary") && (Config.GetString("exiftool-binary") == want)) {
			t.Errorf("exiftool-binary setting is not as expected (explicit %s)", want)
		}
	})

}

/*
// This test won't work as conceived because viper is too lenient on what it considers a usable config file.
// See: https://github.com/spf13/viper/issues/1843

func Test_SpecificCfgFile_Bad(t *testing.T) {
	var args cli_args
	testNameSlug := "specificcfg-bad-"

	args = append(args, "--config-file=../../test/config/path3/mediafooler.jsooon")

	Initialize(args)

	t.Run(testNameSlug+"readconfiguration", func(t *testing.T) {
		cfgRead, err := ReadConfiguration()
		if cfgRead || (err == nil) {
			t.Fatal(testNameSlug + "ReadConfiguration() succeeded when it should have failed")
		}
	})

}
*/

func Test_SpecificCfgFile_Missing(t *testing.T) {
	var args cli_args
	testNameSlug := "specificcfg-missing-"

	args = append(args, "--config-file=../../test/config/path_that_does_not_exist/mediafiler.yaml")

	Initialize(args)

	t.Run(testNameSlug+"readconfiguration", func(t *testing.T) {
		expected := "no such file or directory"
		cfgRead, err := ReadConfiguration()

		msg := err.Error()
		offset := len(msg) - len(expected)
		t.Logf("err: %s", err)

		if cfgRead || (msg[offset:] != expected) {
			t.Fatal(testNameSlug + "ReadConfiguration() succeeded when it should have failed")
		}
	})

}
