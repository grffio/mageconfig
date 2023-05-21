package mageconfig

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetTagOrDefault(t *testing.T) {
	testCases := []struct {
		field reflect.StructField
		tag   string
		want  string
	}{
		{
			field: reflect.StructField{Name: "TestField", Tag: `arg:"testArg"`},
			tag:   "arg",
			want:  "testArg",
		},
		{
			field: reflect.StructField{Name: "TestField", Tag: `arg:""`},
			tag:   "arg",
			want:  "testfield",
		},
	}

	for _, testCase := range testCases {
		got := getTagOrDefault(testCase.field, testCase.tag)
		assert.Equal(t, testCase.want, got)
	}
}

func TestSetFields(t *testing.T) {
	type TestConfig struct {
		A string `file:"a"`
		B int    `file:"b"`
	}

	testCases := []struct {
		name     string
		cfg      Config
		setValue func(field reflect.StructField, value reflect.Value) error
		err      string
	}{
		{
			name: "valid configuration",
			cfg:  &TestConfig{},
			setValue: func(field reflect.StructField, value reflect.Value) error {
				switch field.Name {
				case "A":
					value.SetString("Test")
				case "B":
					value.SetInt(42)
				default:
					return fmt.Errorf("unexpected field: %s", field.Name)
				}
				return nil
			},
			err: "",
		},
		{
			name: "set value returns error",
			cfg:  &TestConfig{},
			setValue: func(field reflect.StructField, value reflect.Value) error {
				return fmt.Errorf("forced error")
			},
			err: "forced error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := setFields(tc.cfg, tc.setValue)
			if tc.err != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
				cfg := tc.cfg.(*TestConfig)
				assert.Equal(t, "Test", cfg.A)
				assert.Equal(t, 42, cfg.B)
			}
		})
	}
}

func TestSetFieldByKind(t *testing.T) {
	type TestConfig struct {
		StringSlice []string       `default:"one,two,three"`
		IntMap      map[string]int `default:"key1:1,key2:2,key3:3"`
		String      string         `default:"test"`
	}

	testCases := []struct {
		name   string
		field  reflect.StructField
		value  reflect.Value
		strVal string
		err    string
	}{
		{
			name:   "slice field",
			field:  reflect.TypeOf(TestConfig{}).Field(0),
			value:  reflect.New(reflect.TypeOf(TestConfig{}).Field(0).Type).Elem(),
			strVal: "one,two,three",
			err:    "",
		},
		{
			name:   "map field",
			field:  reflect.TypeOf(TestConfig{}).Field(1),
			value:  reflect.New(reflect.TypeOf(TestConfig{}).Field(1).Type).Elem(),
			strVal: "key1:1,key2:2,key3:3",
			err:    "",
		},
		{
			name:   "string field",
			field:  reflect.TypeOf(TestConfig{}).Field(2),
			value:  reflect.New(reflect.TypeOf(TestConfig{}).Field(2).Type).Elem(),
			strVal: "test",
			err:    "",
		},
		{
			name:   "invalid map field",
			field:  reflect.TypeOf(TestConfig{}).Field(1),
			value:  reflect.New(reflect.TypeOf(TestConfig{}).Field(1).Type).Elem(),
			strVal: "key1,key2:2,key3:3", // Missing value for key1.
			err:    "invalid map default value: key1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := setFieldByKind(tc.field, tc.value, tc.strVal)
			if tc.err != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
				// Verify the values are set correctly.
				switch v := tc.value.Interface().(type) {
				case []string:
					assert.Equal(t, strings.Split(tc.strVal, sliceSeparator), v)
				case map[string]int:
					expected := make(map[string]int)
					for _, pair := range strings.Split(tc.strVal, sliceSeparator) {
						kv := strings.Split(pair, kvSeparator)
						if len(kv) == 2 {
							i, _ := strconv.Atoi(kv[1])
							expected[kv[0]] = i
						}
					}
					assert.Equal(t, expected, v)
				case string:
					assert.Equal(t, tc.strVal, v)
				}
			}
		})
	}
}

func TestParseStringToType(t *testing.T) {
	testCases := []struct {
		name  string
		s     string
		t     reflect.Type
		value reflect.Value
		err   error
	}{
		{
			name:  "parse string to time.Duration",
			s:     "1h30m",
			t:     reflect.TypeOf(time.Duration(0)),
			value: reflect.ValueOf(90 * time.Minute),
			err:   nil,
		},
		{
			name:  "parse invalid string to time.Duration",
			s:     "invalid",
			t:     reflect.TypeOf(time.Duration(0)),
			value: reflect.Value{},
			err:   errors.New("time: invalid duration \"invalid\""),
		},
		{
			name:  "parse string to time.Time",
			s:     "2006-01-02T15:04:05Z",
			t:     reflect.TypeOf(time.Time{}),
			value: reflect.ValueOf(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)),
			err:   nil,
		},
		{
			name:  "parse invalid string to time.Time",
			s:     "invalid",
			t:     reflect.TypeOf(time.Time{}),
			value: reflect.Value{},
			err:   errors.New("parsing time \"invalid\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"invalid\" as \"2006\""),
		},
		{
			name:  "parse string to bool",
			s:     "true",
			t:     reflect.TypeOf(bool(false)),
			value: reflect.ValueOf(true),
			err:   nil,
		},
		{
			name:  "parse invalid string to bool",
			s:     "invalid",
			t:     reflect.TypeOf(bool(false)),
			value: reflect.Value{},
			err:   errors.New("strconv.ParseBool: parsing \"invalid\": invalid syntax"),
		},
		{
			name:  "parse string to int",
			s:     "42",
			t:     reflect.TypeOf(int(0)),
			value: reflect.ValueOf(int(42)),
			err:   nil,
		},
		{
			name:  "parse invalid string to int",
			s:     "invalid",
			t:     reflect.TypeOf(int(0)),
			value: reflect.Value{},
			err:   errors.New("strconv.ParseInt: parsing \"invalid\": invalid syntax"),
		},
		{
			name:  "parse string to uint",
			s:     "42",
			t:     reflect.TypeOf(uint(0)),
			value: reflect.ValueOf(uint(42)),
			err:   nil,
		},
		{
			name:  "parse invalid string to uint",
			s:     "invalid",
			t:     reflect.TypeOf(uint(0)),
			value: reflect.Value{},
			err:   errors.New("strconv.ParseUint: parsing \"invalid\": invalid syntax"),
		},
		{
			name:  "parse string to float",
			s:     "42.42",
			t:     reflect.TypeOf(float64(0)),
			value: reflect.ValueOf(float64(42.42)),
			err:   nil,
		},
		{
			name:  "parse invalid string to float",
			s:     "invalid",
			t:     reflect.TypeOf(float64(0)),
			value: reflect.Value{},
			err:   errors.New("strconv.ParseFloat: parsing \"invalid\": invalid syntax"),
		},
		{
			name:  "parse string to string",
			s:     "test",
			t:     reflect.TypeOf(string("")),
			value: reflect.ValueOf("test"),
			err:   nil,
		},
		{
			name:  "parse string to unsupported type",
			s:     "test",
			t:     reflect.TypeOf([]byte{}),
			value: reflect.Value{},
			err:   errors.New("unsupported type"),
		},
	}

	assert := assert.New(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := parseStringToType(tc.s, tc.t)
			if tc.err != nil {
				assert.Error(err)
				assert.Equal(tc.err.Error(), err.Error())
			} else {
				assert.NoError(err)
				assert.Equal(tc.value.Interface(), value.Interface())
			}
		})
	}
}
