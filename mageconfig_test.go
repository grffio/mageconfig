package mageconfig

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	DField0 bool   `arg:"dfield0"`
	DField1 bool   `arg:"dfield1" depends:"Field1,DField0"`
	Field0  bool   `arg:"field0" default:"false"`
	Field1  string `arg:"field1" default:"default1"`
	Field2  int    `arg:"field2" default:"2"`
	Field3  string `env:"FIELD3" default:"envDefault"`
	Field4  string `file:"field4" default:"fileDefault"`
	Field5  string `file:"field5" arg:"field5" env:"FIELD5" default:"allDefault"`
	Field6  string `file:"field6" arg:"field6" env:"FIELD6" required:"true"`
}

func TestLoad(t *testing.T) {
	testCases := []struct {
		name       string
		file       string
		env        map[string]string
		args       []string
		wantConfig Config
		wantErr    error
	}{
		{
			name: "Load with default values",
			file: "",
			env:  map[string]string{},
			args: []string{"-dfield0=true", "-field6=required"},
			wantConfig: TestConfig{
				DField0: true,
				Field1:  "default1",
				Field2:  2,
				Field3:  "envDefault",
				Field4:  "fileDefault",
				Field5:  "allDefault",
				Field6:  "required",
			},
			wantErr: nil,
		},
		{
			name: "Load with args values",
			file: "",
			env:  map[string]string{},
			args: []string{"-dfield0=true", "-field1=arg1", "-field2", "3", "--field0", "-field6=required"},
			wantConfig: TestConfig{
				DField0: true,
				Field0:  true,
				Field1:  "arg1",
				Field2:  3,
				Field3:  "envDefault",
				Field4:  "fileDefault",
				Field5:  "allDefault",
				Field6:  "required",
			},
			wantErr: nil,
		},
		{
			name: "Load with env values",
			file: "",
			env: map[string]string{
				"FIELD3": "env3",
				"FIELD5": "env5",
				"FIELD6": "required",
			},
			args: []string{"-dfield0=true"},
			wantConfig: TestConfig{
				DField0: true,
				Field1:  "default1",
				Field2:  2,
				Field3:  "env3",
				Field4:  "fileDefault",
				Field5:  "env5",
				Field6:  "required",
			},
			wantErr: nil,
		},
		{
			name: "Load with file values",
			file: "testdata/config.file",
			env:  map[string]string{},
			args: []string{"-dfield0=true"},
			wantConfig: TestConfig{
				DField0: true,
				Field1:  "default1",
				Field2:  2,
				Field3:  "envDefault",
				Field4:  "file4",
				Field5:  "file5",
				Field6:  "file6",
			},
			wantErr: nil,
		},
		{
			name: "Load with all sources",
			file: "testdata/config.file",
			env: map[string]string{
				"FIELD3": "env3",
				"FIELD5": "env5",
			},
			args: []string{"-dfield0=true", "-field1", "arg1", "-field2=3", "--field6=required", "--field0"},
			wantConfig: TestConfig{
				DField0: true,
				Field0:  true,
				Field1:  "arg1",
				Field2:  3,
				Field3:  "env3",
				Field4:  "file4",
				Field5:  "env5",
				Field6:  "required",
			},
			wantErr: nil,
		},
		{
			name:       "Required field not set",
			file:       "",
			env:        map[string]string{},
			args:       []string{"-dfield0=true"},
			wantConfig: TestConfig{},
			wantErr:    fmt.Errorf("%s: %s", ErrRequiredNotSet.Error(), "Field6"),
		},
		{
			name:       "Depends field not set",
			file:       "",
			env:        map[string]string{},
			args:       []string{"-field6=required"},
			wantConfig: TestConfig{},
			wantErr:    fmt.Errorf("%s: %s", ErrDependsNotSet.Error(), "DField0"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare the environment.
			os.Args = append([]string{"cmd"}, tc.args...)
			for k, v := range tc.env {
				os.Setenv(k, v)
			}

			// Create the config and load it.
			cfg := TestConfig{}
			err := Load(&cfg, tc.file)
			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantConfig, cfg)
			}

			isLoaded = false
			// Not pointer configuration test.
			assert.Equal(t, Load(cfg, tc.file), errors.New("config must be a pointer to a struct"))

			// Cleanup the environment.
			for k := range tc.env {
				os.Unsetenv(k)
			}
		})
	}
}
