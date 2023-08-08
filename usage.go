package mageconfig

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// isHelpRequested checks if the help flag (-help or --help) was provided in the command-line arguments.
func isHelpRequested() bool {
	for _, arg := range os.Args {
		if arg == "-help" || arg == "--help" {
			return true
		}
	}

	return false
}

// printUsage prints the usage instructions for the application, including the available configurations,
// their types, default values, and whether they are required.
func printUsage(cfgType reflect.Type) {
	const helpMessage = "This application is configured via the config file," +
		" environment variables, or command-line arguments.\n" +
		"The following configurations can be used:\n" +
		"[CONFIG FILE KEY, ENVIRONMENT VARIABLE, CLI ARGUMENT]"

	fmt.Fprintln(flag.CommandLine.Output(), "Usage of", os.Args[0])
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), helpMessage)
	fmt.Fprintln(flag.CommandLine.Output())

	// Iterate over each field in the configuration type and print its details.
	for i := 0; i < cfgType.NumField(); i++ {
		field := cfgType.Field(i)

		// Retrieve the field details from its tags.
		argName := getTagOrDefault(field, tagArg)
		envName := field.Tag.Get(tagEnv)
		fileFieldName := field.Tag.Get(tagFile)
		defaultValue := field.Tag.Get(tagDefault)
		description := field.Tag.Get(tagDesc)
		required := field.Tag.Get(tagRequired)
		dependsStr := field.Tag.Get(tagDepends)

		// Define a placeholder for unused fields.
		const notUsedStr = "<NOTUSED>"
		if envName == "" {
			envName = notUsedStr
		}
		if fileFieldName == "" {
			fileFieldName = notUsedStr
		}

		// Determine the type of the field for the help message.
		typeStr := "String" // default type as string.
		switch field.Type.Kind() {
		case reflect.Bool:
			typeStr = "True or False"
		case reflect.Int, reflect.Int32, reflect.Int64:
			typeStr = "Integer"
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			typeStr = "Unsigned Integer"
		case reflect.Float32, reflect.Float64:
			typeStr = "Float"
		case reflect.Slice:
			typeStr = "List"
		case reflect.Map:
			typeStr = "Map"
		}

		fmt.Fprintf(flag.CommandLine.Output(), "%s, %s, --%s:\n", fileFieldName, envName, argName)
		fmt.Fprintf(flag.CommandLine.Output(), "    description: %s\n", description)
		fmt.Fprintf(flag.CommandLine.Output(), "    type:        %s\n", typeStr)
		if defaultValue != "" {
			fmt.Fprintf(flag.CommandLine.Output(), "    default:     %s\n", defaultValue)
		}
		if required == "true" {
			fmt.Fprintf(flag.CommandLine.Output(), "    required:    true\n")
		}
		if dependsStr != "" {
			fmt.Fprintf(flag.CommandLine.Output(), "     depends:     %s\n", strings.Split(dependsStr, ","))
		}
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
	}
}
