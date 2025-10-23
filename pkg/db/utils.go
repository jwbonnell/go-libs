package db

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"reflect"
)

func StructToNamedArgs(s interface{}) (pgx.NamedArgs, error) {
	namedArgs := make(pgx.NamedArgs)
	val := reflect.ValueOf(s)

	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Dereference if a pointer
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or a pointer to a struct")
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Get the tag name, defaulting to field name if no "db" tag
		tagName := field.Tag.Get("db")
		if tagName == "" {
			tagName = field.Name
		}

		namedArgs[tagName] = fieldValue.Interface()
	}
	return namedArgs, nil
}
