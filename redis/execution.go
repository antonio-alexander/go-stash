package redis

import errors "github.com/pkg/errors"

func parseKey(key interface{}) (string, error) {
	switch key := key.(type) {
	default:
		return "", errors.Errorf("unsupported key type: %T", key)
	case string:
		return key, nil
	}
}
