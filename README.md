# MageConfig

MageConfig is a Go library designed for flexible configuration management. It enables the loading of parameters from diverse sources including configuration files, environment variables, and command-line arguments into a designated configuration struct. This is achieved through the use of Go's struct tags, providing control over the origin and method of loading for each parameter. In addition, MageConfig offers the capability to set default values and manage required parameters effectively.

## Features

- **Loading Configuration**: `mageconfig` allows you to load configuration parameters from multiple sources into a configuration struct, including files, environment variables, and command-line arguments.
- **Default Values**: You can set default values for configuration fields using the `default` tag. If a value is not provided through other sources, the default value will be used.
- **Required Parameters**: Mark configuration fields as required using the `required` tag. If a required parameter is not set, an error will be returned.
- **Multiple Data Types**: `mageconfig` supports various data types for configuration fields, including `bool`, `int`, `[]int`, `uint`, `[]uint`, `float`, `[]float`, `string`, `[]string`, `time.Duration`, `time.Time`, and `map[string]bool|int|uint|float|string|time.Duration|time.Time`.
- **Usage Help**: `mageconfig` provides a built-in usage help functionality that can be triggered by passing the `-help` or `--help` command-line argument.
- Command line arguments that come after the target argument (with the specified prefix) can be removed using DropArgsAfterTarget function.
- The configuration loading process follows a specific priority order: configuration file values are overwritten by environment variable values, which in turn are overwritten by argument values. This means that if the same configuration parameter is specified in multiple places, the argument value will take precedence over the environment variable value, which will take precedence over the configuration file value.

## Limitations

- It's important to note that the configuration you provide to mageconfig should be a pointer to a struct or mageconfig will return an error.

- When running `mage -help`, it will display the help information for `mage` itself. To view the help for a specific target in `mage.go`, you can use the command `mage dummy -help` or invoke it after compiling it into a binary file. Also, it's worth noting that executing a command in the format `mage -v -debug --arg-name=value` will not be possible.

## Supported Tags

- `file`: Defines the name of the parameter in the configuration file.
- `env`: Defines the name of the environment variable.
- `arg`: Defines the name of the command-line argument.
- `default`: Defines the default value of the parameter.
- `required`: If set to "true", the parameter is required. If a required parameter is not set, the Load function will return an error.
- `desc`: The description of the parameter, used for the help print.

## Default Naming Convention

In mageconfig, you specify the names of configuration parameters using struct tags for each field in your configuration struct. For example, you can specify the name of a parameter in a configuration file with the `file` tag, an environment variable with the `env` tag, and a command-line argument with the `arg` tag.

However, if you don't specify a name using these tags, mageconfig will default to using the name of the struct field itself. The name will be converted to lower case and will be expected in this form in your configuration file, as an environment variable, and as a command-line argument.

## Configuration File Format

The configuration file should be a plain text file where each line defines a parameter. The parameter name and its value should be separated by a colon.

```txt
param1: value1
param2: value2
```

Lines that do not conform to this format will be ignored.

## Command Line Interface
You can run `mage` with various options and targets, followed by arguments for mageconfig:

```bash
mage [mage-options] [targets] [mageconfig-arguments]
```

Mageconfig arguments can be specified in two ways:

- With a single dash and the argument name (e.g., `-arg-name`)
- With two dashes and the argument name (e.g., `--arg-name`)

The value for an argument can be specified either by appending it directly with an equal sign, or by providing it after a space. For example:

- `-arg-name1="value1"`
- `--arg-name2 "value2"`

Boolean arguments can be specified in the same way, but they also have an additional feature: if you provide the argument name without a value, it will automatically be interpreted as `true`. For example:

- `--arg-bool1`
- `--arg-bool2=false`

In this case, `--arg-bool1` is equivalent to `--arg-bool1=true`.

## Installation

To use `mageconfig` in your Go project, you can install it using the `go get` command:

```
go get github.com/grffio/mageconfig
```

## How to Use
An example `mage.go` is located in the `examples` directory.

You can execute the example with the following command:

```bash
env DB_URL="postgresql://localhost:5432" mage -debug -v showconfig --api-key "abc123"
```

This command sets the `DB_URL` environment variable, runs the `mage` command with the `showconfig` target, includes mage's debug option, and specifies the `--api-key` argument.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter any issues, please open an issue on the GitHub repository. We'll do our best to address it promptly.