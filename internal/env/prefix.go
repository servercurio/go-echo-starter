package env

import "strings"

const separator = "_"

func AddPrefix(prefix, key string) string {
	prefix = strings.ToUpper(strings.TrimSpace(prefix))
	key = strings.ToUpper(strings.TrimSpace(key))

	if prefix == "" {
		return key
	}

	if !strings.HasSuffix(prefix, separator) {
		prefix += separator
	}

	if strings.HasPrefix(key, separator) {
		key = key[len(separator):]
	}

	if strings.HasPrefix(key, prefix) {
		return key
	}

	return prefix + key
}

func RemovePrefix(prefix, key string) string {
	prefix = strings.ToUpper(strings.TrimSpace(prefix))
	key = strings.ToUpper(strings.TrimSpace(key))

	if prefix == "" {
		return key
	}

	if strings.HasPrefix(key, prefix) {
		return key[len(prefix):]
	}

	return key
}
