//go:build mage

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/grffio/mageconfig"
)

// Config represents a set of configuration parameters.
// Each field in the struct corresponds to a configuration parameter and may have tags such as 'file', 'env', 'arg', 'required',
// 'default', 'desc', and 'depends'. These tags dictate the source of the parameter, its necessity, default value,
// description, and any dependencies it might have on other parameters.
type Config struct {
	DatabaseURL       string `file:"dbURL" env:"DB_URL" arg:"db-url" required:"true" desc:"Database URL"`
	BackupDatabaseURL string `file:"backupDBURL" env:"BACKUP_DB_URL" arg:"backup-db-url" desc:"Backup Database URL" depends:"DatabaseURL"`

	APIKey     string        `file:"apiKey" env:"API_KEY" arg:"api-key" required:"true" desc:"API Key"`
	MaxRetries int           `file:"maxRetries" env:"MAX_RETRIES" arg:"max-retries" default:"3" desc:"Maximum number of retries"`
	Timeout    time.Duration `file:"timeout" env:"TIMEOUT" arg:"timeout" default:"5s" desc:"Timeout duration"`
}

// appConfig is a global instance of Config, to hold the loaded configuration parameters.
var appConfig = &Config{}

// init is invoked automatically during package initialization. Its primary role is to load the
// configuration parameters into appConfig using the mageconfig package. If the configuration
// loading fails, the function outputs an error message to stderr and terminates the application with
// a status code of 1. Upon successful configuration loading, the function discards any command-line
// arguments following the Mage target to prevent them from being misinterpreted as additional Mage
// targets or task arguments.
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
	fmt.Println("Backup Database URL:", appConfig.BackupDatabaseURL)
	fmt.Println("API Key:", appConfig.APIKey)
	fmt.Println("Max Retries:", appConfig.MaxRetries)
	fmt.Println("Timeout:", appConfig.Timeout)
}
