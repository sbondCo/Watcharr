package main

import (
	"log/slog"
	"reflect"
	"strconv"

	"github.com/robfig/go-cache"
)

// Extension of the `.Get` method for `go-cache`.
// This method will simplify our usage so we don't need
// to type assert everywhere, this method will handle
// everything related to getting the value from cache.
// Returns `true` if `rv` was set to the cached value.
// Returns `false` if we couldn't get anything from cache.
func GetCache(c *cache.Cache, k string, rv any) bool {
	if val, found := c.Get(k); found {
		v := reflect.ValueOf(rv)
		if v.Type().Kind() == reflect.Ptr && v.Elem().CanSet() {
			v.Elem().Set(reflect.ValueOf(val))
			slog.Debug("cachefunc: Cache found.", "key", k)
			return true
		}
		slog.Error("cachefunc: Cache not set", "key", k)
		return false
	}
	slog.Debug("cachefunc: Cache not found", "key", k)
	return false
}

// Create a cache key for our in-mem cache.
//
// `name` should be the name of the function response we are caching.
//
// `...u` can be any amount of values that will make this key unique.
// Currently supports types:
//   - `string`
//   - `map[string]string`
//   - `int`
func CreateCacheKey(name string, u ...any) string {
	str := name
	appnd := func(s string) {
		str += "-" + s
	}
	for _, v := range u {
		switch vv := v.(type) {
		case string:
			appnd(vv)
		case map[string]string:
			for k, e := range vv {
				appnd(k + "_" + e)
			}
		case int:
			appnd(strconv.Itoa(vv))
		default:
			// This should never happen, but incase of unknown
			// value passed, hopefully this should make it easier
			// to catch in logs.
			str = str + "KEYTYPEERR"
		}
	}
	return str
}
