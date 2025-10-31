package mapper

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// MapSlice maps a slice/array of src elements (value or pointer) to a slice of dest elements of type T.
// Returns a new slice of type []T or an error.
func MapSlice[T any](src any) ([]T, error) {
	var out []T
	if src == nil {
		return out, nil
	}
	sv := reflect.ValueOf(src)
	kind := sv.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return nil, fmt.Errorf("MapSlice: src must be slice or array, got %s", kind)
	}
	ln := sv.Len()
	out = make([]T, 0, ln)
	for i := 0; i < ln; i++ {
		elem := sv.Index(i).Interface()
		// If element is nil pointer, append zero T
		if reflect.ValueOf(elem).Kind() == reflect.Pointer && reflect.ValueOf(elem).IsNil() {
			out = append(out, *new(T))
			continue
		}
		mapped, err := MapStruct[T](elem)
		if err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
		out = append(out, mapped)
	}
	return out, nil
}

// MapStruct maps from src (any struct) to a new instance of dest type T.
// It returns the mapped value (type T) or an error.
func MapStruct[T any](src any) (T, error) {
	var zero T
	if src == nil {
		return zero, nil
	}
	srcV := reflect.ValueOf(src)
	if srcV.Kind() == reflect.Pointer {
		if srcV.IsNil() {
			return zero, nil
		}
		srcV = srcV.Elem()
	}
	if srcV.Kind() != reflect.Struct {
		return zero, errors.New("MapStruct: src must be a struct or pointer to struct")
	}
	destT := reflect.TypeOf((*T)(nil)).Elem()
	if destT.Kind() == reflect.Pointer {
		return zero, errors.New("MapStruct: T must be a struct type, not pointer to struct")
	}
	outV := reflect.New(destT).Elem()
	if err := mapValue(srcV, outV); err != nil {
		return zero, err
	}
	return outV.Interface().(T), nil
}

// mapValue maps src reflect.Value (struct) into dest reflect.Value (struct).
func mapValue(src, dest reflect.Value) error {
	if src.Kind() != reflect.Struct || dest.Kind() != reflect.Struct {
		return fmt.Errorf("mapValue: both src and dest must be struct, got %s and %s", src.Kind(), dest.Kind())
	}

	// Build dest field map: case-insensitive field name -> field index
	destFieldMap := make(map[string]int)
	destType := dest.Type()
	for i := 0; i < dest.NumField(); i++ {
		f := destType.Field(i)
		if f.PkgPath != "" { // unexported
			continue
		}
		name := strings.ToLower(f.Name)
		destFieldMap[name] = i
	}

	srcType := src.Type()
	for i := 0; i < src.NumField(); i++ {
		sf := srcType.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}
		name := strings.ToLower(sf.Name)
		dfIndex, ok := destFieldMap[name]
		if !ok {
			continue
		}
		srcField := src.Field(i)
		destField := dest.Field(dfIndex)
		if !destField.CanSet() {
			continue
		}
		if err := setAssignable(srcField, destField); err != nil {
			return fmt.Errorf("field %s -> %s: %w", sf.Name, destType.Field(dfIndex).Name, err)
		}
	}
	return nil
}

// setAssignable tries to set dest from src with conversions, recursion for structs/slices.
func setAssignable(src, dest reflect.Value) error {
	// Handle pointers by dereferencing or creating pointer as needed.
	// Normalize src nil pointer to zero value.
	for src.Kind() == reflect.Pointer {
		if src.IsNil() {
			// zero value for dest: leave as zero (no set)
			return nil
		}
		src = src.Elem()
	}

	// If dest is pointer, create and set underlying value.
	if dest.Kind() == reflect.Pointer {
		// ensure underlying type
		under := dest.Type().Elem()
		newVal := reflect.New(under)
		if err := setAssignable(src, newVal.Elem()); err != nil {
			return err
		}
		dest.Set(newVal)
		return nil
	}

	// If types are exactly assignable
	if src.Type().AssignableTo(dest.Type()) {
		dest.Set(src)
		return nil
	}

	// If convertible by Go rules
	if src.Type().ConvertibleTo(dest.Type()) {
		dest.Set(src.Convert(dest.Type()))
		return nil
	}

	// Custom converter lookup
	key := convKey(src.Type(), dest.Type())
	if conv, ok := converters[key]; ok {
		convVal, err := conv(src, dest.Type())
		if err != nil {
			return err
		}
		if !convVal.Type().AssignableTo(dest.Type()) {
			// try convert
			if convVal.Type().ConvertibleTo(dest.Type()) {
				dest.Set(convVal.Convert(dest.Type()))
				return nil
			}
			return fmt.Errorf("converter returned incompatible type %s for dest %s", convVal.Type(), dest.Type())
		}
		dest.Set(convVal)
		return nil
	}

	// Handle structs by recursion
	if src.Kind() == reflect.Struct && dest.Kind() == reflect.Struct {
		return mapValue(src, dest)
	}

	// Handle slices/arrays
	if src.Kind() == reflect.Slice && (dest.Kind() == reflect.Slice || dest.Kind() == reflect.Array) {
		// create slice of dest element type
		destElemType := dest.Type().Elem()
		srcLen := src.Len()
		newSlice := reflect.MakeSlice(reflect.SliceOf(destElemType), srcLen, srcLen)
		for i := 0; i < srcLen; i++ {
			srcElem := src.Index(i)
			destElem := newSlice.Index(i)
			if err := setAssignable(srcElem, destElem); err != nil {
				return fmt.Errorf("slice index %d: %w", i, err)
			}
		}
		if dest.Kind() == reflect.Array {
			if dest.Len() < newSlice.Len() {
				return fmt.Errorf("destination array too small: %d < %d", dest.Len(), newSlice.Len())
			}
			reflect.Copy(dest, newSlice)
		} else {
			dest.Set(newSlice)
		}
		return nil
	}

	// Handle map -> map (simple key/value assignable)
	if src.Kind() == reflect.Map && dest.Kind() == reflect.Map {
		if dest.IsNil() {
			dest.Set(reflect.MakeMap(dest.Type()))
		}
		for _, k := range src.MapKeys() {
			sv := src.MapIndex(k)
			// try to create converted key/value
			dk := reflect.New(dest.Type().Key()).Elem()
			if err := setAssignable(k, dk); err != nil {
				return fmt.Errorf("map key: %w", err)
			}
			dv := reflect.New(dest.Type().Elem()).Elem()
			if err := setAssignable(sv, dv); err != nil {
				return fmt.Errorf("map value: %w", err)
			}
			dest.SetMapIndex(dk, dv)
		}
		return nil
	}

	return fmt.Errorf("cannot assign %s to %s", src.Type(), dest.Type())
}

// --- Register useful converters for common cases ---

func init() {
	// sql.NullTime -> time.Time
	RegisterConverter(reflect.TypeOf(sql.NullTime{}), reflect.TypeOf(time.Time{}), sqlNullTimeToTime)

	// *sql.NullTime -> time.Time
	RegisterConverter(reflect.TypeOf(&sql.NullTime{}).Elem(), reflect.TypeOf(time.Time{}), sqlNullPtrTimeToTime)

	// time.Time -> sql.NullTime
	RegisterConverter(reflect.TypeOf(time.Time{}), reflect.TypeOf(sql.NullTime{}), timeToSqlNullTime)

	// sql.NullString -> string
	RegisterConverter(reflect.TypeOf(sql.NullString{}), reflect.TypeOf(""), nullStringToString)
}
