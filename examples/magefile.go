//go:build mage

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/grffio/mageconfig"
)

// Config is a struct where each field represents a configuration parameter.
// Fields are tagged with special tags 'file', 'env', 'arg', 'required', 'default', and 'desc' to define where
// the configuration parameter can come from, whether it's required, its default value, and its description.
type Config struct {
	DatabaseURL string        `file:"dbURL" env:"DB_URL" arg:"db-url" required:"true" desc:"Database URL"`
	APIKey      string        `file:"apiKey" env:"API_KEY" arg:"api-key" required:"true" desc:"API Key"`
	MaxRetries  int           `file:"maxRetries" env:"MAX_RETRIES" arg:"max-retries" default:"3" desc:"Maximum number of retries"`
	Timeout     time.Duration `file:"timeout" env:"TIMEOUT" arg:"timeout" default:"5s" desc:"Timeout duration"`
}

// appConfig is a global instance of Config, to hold the loaded configuration parameters.
var appConfig = &Config{}

// init is automatically called when the package is initialized. It is responsible for loading
// the configuration parameters into appConfig using the mageconfig package. If the configuration
// fails to load, an error message is printed to stderr and the application exits with status code 1.
// After successful loading, it drops any command-line arguments after the Mage target to avoid
// misinterpretation as Mage targets or task arguments.
func init() {
	if err := mageconfig.Load(appConfig, "mage.config"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	mageconfig.DropArgsAfterTarget()
}

var Default = ShowConfig

// ShowConfig displays the loaded configuration parameters.
func ShowConfig() {
	fmt.Println("Database URL:", appConfig.DatabaseURL)
	fmt.Println("API Key:", appConfig.APIKey)
	fmt.Println("Max Retries:", appConfig.MaxRetries)
	fmt.Println("Timeout:", appConfig.Timeout)
}
