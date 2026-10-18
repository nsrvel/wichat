// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// Load merges defaults, optional envFile, and OS environment into dst.
// Precedence (low to high): defaults struct, env file, environment variables.
// envFile empty skips file loading; a missing file is not an error.
// defaults must be the same struct type as dst (value or pointer); env keys come from mapstructure tags.
func Load(envFile string, dst, defaults any) error {
	// Validation
	if dst == nil {
		return errors.New("config: dst is nil")
	}

	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Pointer || dstVal.IsNil() {
		return errors.New("config: dst must be a non-nil pointer")
	}

	defaultsVal, err := structValue(defaults)
	if err != nil {
		return err
	}

	// Defaults
	if err := applyDefaults(dstVal, defaultsVal); err != nil {
		return err
	}

	// Viper
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if defaultsVal.IsValid() {
		setViperDefaults(v, defaultsVal)
	}

	// Env file
	if err := readEnvFile(v, envFile); err != nil {
		return err
	}

	// Unmarshal
	if err := v.Unmarshal(dst); err != nil {
		return fmt.Errorf("config: unmarshal: %w", err)
	}

	applyExplicitConfig(v, dstVal.Elem(), defaultsVal)

	return nil
}

func applyDefaults(dstVal, defaultsVal reflect.Value) error {
	if !defaultsVal.IsValid() {
		return nil
	}

	dstElem := dstVal.Elem()
	if dstElem.Type() != defaultsVal.Type() {
		return fmt.Errorf("config: defaults type %s does not match dst type %s", defaultsVal.Type(), dstElem.Type())
	}

	dstElem.Set(defaultsVal)
	return nil
}

func readEnvFile(v *viper.Viper, envFile string) error {
	if envFile == "" {
		return nil
	}

	_, err := os.Stat(envFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("config: stat %q: %w", envFile, err)
	}

	v.SetConfigFile(envFile)
	v.SetConfigType("env")
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("config: read %q: %w", envFile, err)
	}

	return nil
}

func structValue(v any) (reflect.Value, error) {
	if v == nil {
		return reflect.Value{}, nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Pointer {
		if val.Kind() != reflect.Struct {
			return reflect.Value{}, errors.New("config: defaults must be a struct")
		}
		return val, nil
	}

	if val.IsNil() {
		return reflect.Value{}, errors.New("config: defaults pointer is nil")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return reflect.Value{}, errors.New("config: defaults must be a struct")
	}

	return val, nil
}

func setViperDefaults(v *viper.Viper, structVal reflect.Value) {
	t := structVal.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		key := mapstructureKey(field)
		if key == "" {
			continue
		}

		v.SetDefault(key, structVal.Field(i).Interface())
	}
}

func mapstructureKey(field reflect.StructField) string {
	tag := field.Tag.Get("mapstructure")
	if tag == "" || tag == "-" {
		return ""
	}

	name := strings.Split(tag, ",")[0]
	if name == "" || name == "squash" {
		return ""
	}

	return name
}
