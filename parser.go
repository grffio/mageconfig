package mageconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// getTagOrDefault retrieves the value of the tag for a given struct field.
// If the tag is absent, it returns the field name in lower case.
func getTagOrDefault(field reflect.StructField, tag string) string {
	value := field.Tag.Get(tag)
	if value == "" {
		return strings.ToLower(field.Name)
	}

	return value
}

// setFields iterates over each field in the given configuration and applies the setValue function to it.
// The setValue function is responsible for assigning a value to the field.
// This function is used to abstract the common pattern of iterating over struct fields.
func setFields(cfg Config, setValue func(field reflect.StructField, value reflect.Value) error) error {
	cfgValue := reflect.ValueOf(cfg)
	cfgType := reflect.TypeOf(cfg)

	// Dereference the pointer to get the actual struct value and type.
	cfgValue = cfgValue.Elem()
	cfgType = cfgType.Elem()
	// Iterate over each field in the struct and apply the setValue function to the current field.
	for i := 0; i < cfgType.NumField(); i++ {
		field := cfgType.Field(i)
		value := cfgValue.Field(i)
		if err := setValue(field, value); err != nil {
			return err
		}
	}

	return nil
}

// setFieldByKind assigns a value to a struct field based on its kind (type).
// It supports slice, map, and basic types.
func setFieldByKind(field reflect.StructField, value reflect.Value, strVal string) error {
	switch field.Type.Kind() {
	case reflect.Slice:
		// Handle slice types: split the string value into elements, create a new slice with the appropriate type and size
		// and iterate over each element in the string value.
		elems := strings.Split(strVal, sliceSeparator)
		slice := reflect.MakeSlice(field.Type, len(elems), len(elems))
		for i, e := range elems {
			// Convert the string element to the appropriate type and assign it to the slice.
			v, err := parseStringToType(strings.TrimSpace(e), field.Type.Elem())
			if err != nil {
				return err
			}
			slice.Index(i).Set(v)
		}
		value.Set(slice)

	case reflect.Map:
		// Handle map types: split the string value into key-value pairs, create a new map with the appropriate type
		// and iterate over each key-value pair in the string value.
		elems := strings.Split(strVal, sliceSeparator)
		mapType := reflect.MapOf(reflect.TypeOf(""), field.Type.Elem())
		mapValue := reflect.MakeMap(mapType)
		for _, pair := range elems {
			kv := strings.SplitN(pair, kvSeparator, 2)
			if len(kv) != 2 {
				return fmt.Errorf("invalid map default value: %s", pair)
			}
			// Convert the string value to the appropriate type and assign it to the map.
			v, err := parseStringToType(strings.TrimSpace(kv[1]), field.Type.Elem())
			if err != nil {
				return err
			}
			mapValue.SetMapIndex(reflect.ValueOf(strings.TrimSpace(kv[0])), v)
		}
		value.Set(mapValue)

	default:
		// For basic types, convert the string value to the appropriate type and assign it to the field.
		v, err := parseStringToType(strVal, field.Type)
		if err != nil {
			return fmt.Errorf("parse field: %w", err)
		}
		value.Set(v)
	}

	return nil
}

// parseStringToType is a helper function that parses a string into a specified type represented by reflect.Type.
// It supports bool, int, uint, float, string, time.Duration, and time.Time.
// This function is used to abstract the common pattern of parsing a string to different kinds of types.
func parseStringToType(s string, t reflect.Type) (reflect.Value, error) {
	switch t {
	case reflect.TypeOf(time.Duration(0)):
		v, err := time.ParseDuration(s)
		return reflect.ValueOf(v), err
	case reflect.TypeOf(time.Time{}):
		v, err := time.Parse(time.RFC3339, s)
		return reflect.ValueOf(v), err
	}

	switch t.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(s)
		return reflect.ValueOf(v), err
	case reflect.Int, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(s, 10, 64)
		return reflect.ValueOf(int(v)), err
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		return reflect.ValueOf(uint(v)), err
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(s, 64)
		return reflect.ValueOf(v), err
	case reflect.String:
		return reflect.ValueOf(s), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported type")
}
