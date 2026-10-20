package config

import (
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// ListenAddress normalizes PORT-style values for net.Listen (":3000").
func ListenAddress(hostOrPort string) string {
	s := strings.TrimSpace(hostOrPort)
	if s == "" {
		return ""
	}
	if strings.Contains(s, ":") {
		return s
	}
	return ":" + s
}

func applyExplicitConfig(v *viper.Viper, dstVal, defaultsVal reflect.Value) {
	if !dstVal.IsValid() || !defaultsVal.IsValid() {
		return
	}
	walkConfigFields(v, dstVal, defaultsVal)
}

func walkConfigFields(v *viper.Viper, dst, defaults reflect.Value) {
	dst = reflect.Indirect(dst)
	defaults = reflect.Indirect(defaults)
	if dst.Kind() != reflect.Struct || defaults.Kind() != reflect.Struct {
		return
	}

	t := dst.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		tag := field.Tag.Get("mapstructure")
		if strings.Contains(tag, "squash") {
			walkConfigFields(v, dst.Field(i), defaults.Field(i))
			continue
		}

		key := mapstructureKey(field)
		if key == "" {
			continue
		}

		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		dstField := dst.Field(i)
		defField := defaults.Field(i)

		if field.Type.Kind() == reflect.String && strings.TrimSpace(dstField.String()) != "" {
			continue
		}

		if !wasExplicitlyProvided(v, key, envKey) {
			slog.Warn("config: env unset, using default", "key", envKey, "default", defField.Interface())
			dstField.Set(defField)
			continue
		}

		if isEmptyValue(v, key, envKey, field.Type) {
			slog.Warn("config: env empty, using default", "key", envKey, "default", defField.Interface())
			dstField.Set(defField)
		}
	}
}

func wasExplicitlyProvided(v *viper.Viper, key, envKey string) bool {
	if _, ok := os.LookupEnv(envKey); ok {
		return true
	}
	return v.InConfig(key)
}

func isEmptyValue(v *viper.Viper, key, envKey string, typ reflect.Type) bool {
	if typ.Kind() != reflect.String {
		return false
	}
	if s := strings.TrimSpace(v.GetString(key)); s != "" {
		return false
	}
	if s := strings.TrimSpace(v.GetString(envKey)); s != "" {
		return false
	}
	return strings.TrimSpace(os.Getenv(envKey)) == ""
}
