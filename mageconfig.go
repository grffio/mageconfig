package mageconfig

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
)

// Tag constants used for struct field tags.
const (
	tagArg         = "arg"      // Defines the name of the command-line argument.
	tagEnv         = "env"      // Defines the name of the environment variable.
	tagFile        = "file"     // Defines the name of the parameter in the configuration file.
	tagDefault     = "default"  // Defines the default value of the parameter.
	tagDesc        = "desc"     // Provides a description for the parameter.
	tagDepends     = "depends"  // Specifies other parameters that this parameter depends on.
	tagRequired    = "required" // Specifies whether the parameter is required.
	argPrefix      = "-"        // The prefix used for command-line arguments.
	sliceSeparator = ","        // The separator used for slice elements.
	kvSeparator    = ":"        // The separator used for key-value pairs in the configuration file.
)

// List of default Mage commads and options.
var (
	defaultMageCommands = []string{"-l", "-h"}
	defaultMageOptions  = []string{"-h", "-t", "-v"}
)

// Global variables to manage loading state and ensure thread safety during configuration load.
var (
	once     sync.Once
	isLoaded bool
)

var (
	// ErrRequiredNotSet is the error returned when a required configuration value is not set.
	ErrRequiredNotSet = errors.New("required parameter not set")
	// ErrDependsNotSet is the error returned when a dependent field configuration value is not set.
	ErrDependsNotSet = errors.New("dependent parameter not set")
)

// Config is an interface that all configuration structs should implement.
// Supported types are: bool, int, []int, uint, []uint, float, []float, string, []string,
// time.Duration, and time.Time, map[string]bool|int|uint|float|string|time.Duration|time.Time.
// Slice elements are separated by comma.
type Config interface{}

// Load reads configuration parameters from a file, environment variables, and command-line arguments
// into a configuration struct. It also checks if any required parameters are not set and returns an
// error if any are missing.
func Load(cfg Config, file string) error {
	if isHelpRequested() {
		printUsage(reflect.TypeOf(cfg).Elem())
		os.Exit(0)
	}

	// If the configuration is loaded, there's nothing to do.
	if isLoaded {
		return nil
	}

	// Iterate over the command-line arguments and if the argument matches one of the default Mage commands,
	// then it means that this Mage command is being passed as an argument to the Mage itself,
	// not as an option to the Mage target.
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, argPrefix) {
			// In this case, we stop the execution of the Load function early to process the Mage command
			// as a regular command-line argument, and to avoid potential errors or conflicts.
			if contains(defaultMageCommands, arg) {
				return nil
			}
		}
	}

	// Check if the passed configuration is a pointer to a struct.
	cfgType := reflect.TypeOf(cfg)
	if cfgType.Kind() != reflect.Pointer || cfgType.Elem().Kind() != reflect.Struct {
		return errors.New("config must be a pointer to a struct")
	}

	// Map to keep track of which configuration parameters have been set.
	isSet := make(map[string]*bool)
	initializeIsSet(cfg, isSet)

	// Set the default values for configuration parameters.
	if err := setDefault(cfg, isSet); err != nil {
		return err
	}

	// Load the configuration from a file.
	if err := loadFromFile(cfg, file, isSet); err != nil {
		return err
	}

	// Load the configuration from environment variables.
	if err := loadFromEnv(cfg, isSet); err != nil {
		return err
	}

	// Load the configuration from command-line arguments.
	if err := loadFromArgs(cfg, isSet); err != nil {
		return err
	}

	// Ensure that the configuration is loaded only once.
	once.Do(func() {
		isLoaded = true
	})

	// Check that all required and dependent fields in the configuration have been set.
	return checkRequiredAndDepends(cfg, isSet)
}

// DropArgsAfterTarget removes command-line arguments that come after the target argument (with the specified prefix).
func DropArgsAfterTarget() {
	// If the configuration is not loaded, there's nothing to do.
	if !isLoaded {
		return
	}

	// Find the index of the first argument with the specified prefix (after target argument).
	for i, arg := range os.Args {
		if strings.HasPrefix(arg, argPrefix) {
			// If the argument is a default mage option, skip to the next iteration.
			if contains(defaultMageOptions, arg) {
				continue
			}
			os.Args = os.Args[:i] // Keep the target name and remove all arguments after it.
			return
		}
	}
}

// contains check if a string slice contains a specific string.
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// initializeIsSet initializes the isSet map to track which configuration parameters have been set.
func initializeIsSet(cfg Config, isSet map[string]*bool) {
	cfgType := reflect.TypeOf(cfg).Elem()
	for i := 0; i < cfgType.NumField(); i++ {
		field := cfgType.Field(i)
		fieldName := field.Name
		b := false
		isSet[fieldName] = &b
	}
}

// setDefault sets default values for each field in a struct based on the 'tagDefault' tag.
func setDefault(cfg Config, isSet map[string]*bool) error {
	return setFields(cfg, func(field reflect.StructField, value reflect.Value) error {
		defaultValue := field.Tag.Get(tagDefault)
		if defaultValue == "" {
			return nil
		}

		if err := setFieldByKind(field, value, defaultValue); err != nil {
			return err
		}
		*isSet[field.Name] = true

		return nil
	})
}

// loadFromFile loads configuration parameters from a file into a configuration struct.
func loadFromFile(cfg Config, file string, isSet map[string]*bool) error {
	if file == "" {
		return nil
	}

	f, err := os.Open(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()

	// Read the file into a map.
	fileContent := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, kvSeparator, 2)
		if len(parts) != 2 {
			continue // Skip lines with invalid format.
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Strip quotes from value if present.
		if len(value) > 0 &&
			(value[0] == '"' && value[len(value)-1] == '"' ||
				value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}

		fileContent[key] = value
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	// Load fields from the map.
	return setFields(cfg, func(field reflect.StructField, value reflect.Value) error {
		fileName := field.Tag.Get(tagFile)
		if fileName == "" {
			return nil
		}

		fileValue, ok := fileContent[fileName]
		if !ok {
			return nil
		}

		if err := setFieldByKind(field, value, fileValue); err != nil {
			return err
		}
		*isSet[field.Name] = true

		return nil
	})
}

// loadFromEnv loads configuration parameters from environment variables into a configuration struct.
func loadFromEnv(cfg Config, isSet map[string]*bool) error {
	return setFields(cfg, func(field reflect.StructField, value reflect.Value) error {
		envName := field.Tag.Get(tagEnv)
		if envName == "" {
			return nil
		}

		envValue, ok := os.LookupEnv(envName)
		if !ok {
			return nil
		}

		if err := setFieldByKind(field, value, envValue); err != nil {
			return err
		}
		*isSet[field.Name] = true

		return nil
	})
}

// loadFromArgs loads configuration parameters from command-line arguments into a configuration struct.
func loadFromArgs(cfg Config, isSet map[string]*bool) error {
	return setFields(cfg, func(field reflect.StructField, value reflect.Value) error {
		argName := getTagOrDefault(field, tagArg)

		argValue := getArgValue(argName, field.Type.Kind() == reflect.Bool)
		if argValue == "" { // No value found for this argument.
			return nil
		}

		if err := setFieldByKind(field, value, argValue); err != nil {
			return err
		}
		*isSet[field.Name] = true

		return nil
	})
}

// getArgValue scans the command-line arguments for the specified argument. For non-boolean arguments,
// it looks for a value specified with "=" or a space. For boolean arguments, it also accepts the lack
// of an explicitly specified value as "true".
func getArgValue(argName string, isBool bool) string {
	for i := 1; i < len(os.Args); i++ {
		arg := strings.TrimLeft(os.Args[i], argPrefix)
		equalIndex := strings.Index(arg, "=")

		if equalIndex > 0 { // Value is specified with "=".
			key := arg[:equalIndex]
			value := arg[equalIndex+1:]
			if key == argName {
				return value
			}
		} else if arg == argName { // Value is specified with a space or is missing.
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], argPrefix) {
				return os.Args[i+1] // Value is specified with a space.
			} else if isBool { // Value is missing, but it's a boolean argument.
				return "true" //nolint:goconst
			}
		}
	}

	return ""
}

// checkRequiredAndDepends verifies if all required and dependent configuration parameters have been set.
// If a parameter marked 'required' is not set, or
// if a parameter with a 'depends' tag doesn't have its dependencies met,
// it returns an error indicating which parameter is missing.
func checkRequiredAndDepends(cfg Config, isSet map[string]*bool) error {
	cfgType := reflect.TypeOf(cfg).Elem()

	for i := 0; i < cfgType.NumField(); i++ {
		field := cfgType.Field(i)

		required := field.Tag.Get(tagRequired)
		// If the field is marked as 'required' and not set in the 'isSet' map, return an error.
		if required == "true" && (isSet[field.Name] == nil || !*isSet[field.Name]) {
			return fmt.Errorf("%w: %s", ErrRequiredNotSet, field.Name)
		}

		dependsStr := field.Tag.Get(tagDepends)
		if dependsStr != "" {
			depends := strings.Split(dependsStr, ",")
			for _, depend := range depends {
				// If the dependent field is not set in the 'isSet' map, return an error.
				if isSet[depend] == nil || !*isSet[depend] {
					return fmt.Errorf("%w: %s", ErrDependsNotSet, depend)
				}
			}
		}
	}

	return nil
}
