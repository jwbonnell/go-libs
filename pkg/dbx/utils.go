package dbx

import (
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
)

func SliceToNamedArgs(s interface{}) (pgx.NamedArgs, error) {
	if namedArgs, ok := s.(pgx.NamedArgs); ok {
		return namedArgs, nil
	}

	val := reflect.ValueOf(s)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("input must be a slice of structs or a pointer to a slice")
	}

	namedArgs := make(pgx.NamedArgs)

	// Iterate through each struct in the slice
	for i := 0; i < val.Len(); i++ {
		elemVal := val.Index(i)

		// Dereference if element is a pointer
		if elemVal.Kind() == reflect.Ptr {
			elemVal = elemVal.Elem()
		}

		if elemVal.Kind() != reflect.Struct {
			return nil, fmt.Errorf("slice elements must be structs")
		}

		elemType := elemVal.Type()
		for j := 0; j < elemVal.NumField(); j++ {
			field := elemType.Field(j)
			fieldValue := elemVal.Field(j)

			// Get the tag name, defaulting to field name if no "db" tag
			tagName := field.Tag.Get("db")
			if tagName == "" {
				tagName = field.Name
			}

			// Append index to create unique parameter names: name1, name2, etc.
			paramName := fmt.Sprintf("%s%d", tagName, i+1)
			namedArgs[paramName] = fieldValue.Interface()
		}
	}

	return namedArgs, nil
}

func StructToNamedArgs(s interface{}) (pgx.NamedArgs, error) {
	if s, ok := s.(pgx.NamedArgs); ok {
		//Check if pgx.NamedArgs was already passed in and return if so.
		return s, nil
	}

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
