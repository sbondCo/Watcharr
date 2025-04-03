package main

import (
	"log/slog"
	"reflect"

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
