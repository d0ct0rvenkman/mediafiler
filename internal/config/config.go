package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/d0ct0rvenkman/mediafiler/internal/strmanip"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/pflag"

	"github.com/spf13/viper"
)

var Config viper.Viper
var FS *pflag.FlagSet

var ModelReplacer strmanip.Replacer
var PathIgnorer PathIgnoreFilter

var DEFAULT_CONFIG_USED string

/*
Initialize() sets up the config reader and the associate flag sets
for reading in command line arguments.

args []string -  typically will be os.Args[1:], but is also
useful for passing in test data for unit tests.
*/
func Initialize(args []string) {
	Config = *viper.New()

	DEFAULT_CONFIG_USED = "the default configuration was used"

	Config.SetConfigName("mediafiler")
	Config.SetConfigType("yaml")

	FS = pflag.NewFlagSet("mediafiler", pflag.ContinueOnError)
	FS.Bool("dry-run", false, "run in dry-run mode where actions are displayed but not executed")
	FS.Bool("debug", false, "increase logging verbosity to debug level")
	FS.Bool("dump-example-config", false, "dump example configuration file to standard output")
	FS.Bool("use-default-config", false, "use the default/example configuration if a config file cannot"+
		" be found via search paths. if a config file is specified via the 'config-file' argument but"+
		" not found, this flag will have no effect.")

	FS.String("config-file", "", "path to mediafiler configuration file. ")
	FS.String("exiftool-binary", "", "path to exiftool binary")

	err := FS.Parse(args)
	Config.BindPFlags(FS)

	// exit if -h or --help is found
	if err == pflag.ErrHelp {
		os.Exit(0)
	}

}

/*
ProcessFatalFlags() handles any flags that would result in the process exiting intentionally
before normal output or function begins
*/
func ProcessFatalFlags() {
	if ok, _ := FS.GetBool("dump-example-config"); ok {
		fmt.Print(string(defaultConfigYAML))
		os.Exit(0)
	}
}

/*
UseDefaultConfigPaths() configures the config reader to look in the
default paths for the application.
*/
func UseDefaultConfigPaths() {
	Config.AddConfigPath("$HOME/.config/mediafiler")
	Config.AddConfigPath("/etc/mediafiler/")
}

/*
useSpecificConfigPath() configures the config reader to look in specific
paths. Intended for use in test cases only.
*/
func useSpecificConfigPath(path string) {
	Config.AddConfigPath(path)
}

/*
useSpecificConfigFile() tells the configuration reader to use a specific
config file path instead of searching.
*/
func useSpecificConfigFile(path string) {
	Config.SetConfigFile(path)
}

/*
ReadConfiguration() attempts to load in the configuration using previously
configured path/file settings.

Returns:
0: bool - true if a configuration was loaded, false otherwise
1: error - returns information on failure cases. returns contents DEFAULT_CONFIG_USED if default values were loaded if "use-default-config" is specified by the user
*/
func ReadConfiguration() (bool, error) {
	config_file_specified := false

	if Config.IsSet("config-file") {
		file := Config.GetString("config-file")
		if file != "" {
			config_file_specified = true
			useSpecificConfigFile(file)
		} else {
			return false, errors.New("config file path specified via command line arguments is somehow empty")
		}
	}

	if err := Config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// if the user specifies a specific config file that doesn't exist, don't
			// apply default config even if they ask for it
			if Config.GetBool("use-default-config") && !config_file_specified {

				err = ApplyDefaultConfiguration()
				if err == nil {
					return true, errors.New(DEFAULT_CONFIG_USED)
				} else {
					return false, fmt.Errorf("attempt to load defaults resulted in error: %s", err)
				}
			} else {
				return false, errors.New("config file could not be found using configured paths/files")
			}

		} else {
			return false, fmt.Errorf("an error occurred while loading configuration: %s", err)
		}
	}
	return true, nil
}

/*
ApplyDefaultConfiguration() loads in the default/example configuration from defaultConfigYAML.

Returns the error value from the ReadConfig operation.
*/
func ApplyDefaultConfiguration() error {
	return Config.ReadConfig(bytes.NewBuffer(defaultConfigYAML))
}

/*
ProcessConfiguration() reads the structured data from the configuration file
and populates the Model Replacer and Path Ignore rules

returns an error object to indicate success or describe failure
*/

func ProcessConfiguration() error {

	// set these to new empty objects so it's safe to use repeatedly in tests
	ModelReplacer = strmanip.Replacer{}
	PathIgnorer = PathIgnoreFilter{}
	var err error
	var merr error

	if Config.IsSet("model-replace-rules") {
		mrr := Config.Get("model-replace-rules").([]interface{})
		for _, mrrv := range mrr {
			vv := mrrv.(map[string]interface{})
			vvv := strmanip.ReplacerRule{Type: vv["replace_type"].(string), Find: vv["find_pattern"].(string), ReplaceWith: vv["replace_with"].(string)}

			err = ModelReplacer.AddRule(vvv)
			if err != nil {
				merr = multierror.Append(fmt.Errorf("error adding Model Replacer rule: %s", err))
			}
		}
	}

	if Config.IsSet("path-ignore-patterns") {
		pip := Config.Get("path-ignore-patterns").([]interface{})
		for _, pipv := range pip {
			vv := pipv.(map[string]interface{})
			vvv := PathIgnorePattern{Type: vv["type"].(string), Pattern: vv["pattern"].(string)}

			err = PathIgnorer.AddPattern(vvv)

			if err != nil {
				merr = multierror.Append(fmt.Errorf("error adding path ignore rule: %s", err))
			}
		}
	}

	return merr
}
