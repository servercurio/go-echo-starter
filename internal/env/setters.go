package env

import (
	"os"
	"strconv"
	"time"
)

func SetStringValue(prefix, key string, value *string) bool {
	if v, ok := os.LookupEnv(AddPrefix(prefix, key)); ok {
		*value = v
		return true
	}
	return false
}

func SetBoolValue(prefix, key string, value *bool) bool {
	if v, ok := os.LookupEnv(AddPrefix(prefix, key)); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			*value = b
			return true
		}
	}
	return false
}

func SetIntValue(prefix, key string, value *int) bool {
	if v, ok := os.LookupEnv(AddPrefix(prefix, key)); ok {
		if i, err := strconv.Atoi(v); err == nil {
			*value = i
			return true
		}
	}
	return false
}

func SetUint16Value(prefix, key string, value *uint16) bool {
	if v, ok := os.LookupEnv(AddPrefix(prefix, key)); ok {
		if i, err := strconv.ParseUint(v, 10, 16); err == nil {
			*value = uint16(i)
			return true
		}
	}
	return false
}

func SetFloatValue(prefix, key string, value *float64) bool {
	if v, ok := os.LookupEnv(AddPrefix(prefix, key)); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			*value = f
			return true
		}
	}
	return false
}

func SetDurationValue(prefix, key string, value *time.Duration) bool {
	if v, ok := os.LookupEnv(AddPrefix(prefix, key)); ok {
		if d, err := time.ParseDuration(v); err == nil {
			*value = d
			return true
		}
	}
	return false
}
