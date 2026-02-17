package mapper

import (
	"database/sql"
	"reflect"
	"time"
)

// Converter is a function that converts from src value to a dest reflect.Value.
// It should return the converted value (as reflect.Value) or an error.
type Converter func(src reflect.Value, destType reflect.Type) (reflect.Value, error)

type TimeFormatter func(time.Time) string

type TimeParser func(string) (time.Time, error)

// Registry of converters keyed by srcType.String()+"->"+destType.String()
var converters = map[string]Converter{}

// RegisterConverter adds a custom converter for a specific src->dest pair.
func RegisterConverter(srcType, destType reflect.Type, conv Converter) {
	key := convKey(srcType, destType)
	converters[key] = conv
}

func convKey(a, b reflect.Type) string {
	return a.PkgPath() + "." + a.Name() + "->" + b.PkgPath() + "." + b.Name()
}

// SqlNullTimeToTime - converter for sql.NullTime -> time.Time
func SqlNullTimeToTime(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	nt := src.Interface().(sql.NullTime)
	if !nt.Valid {
		return reflect.Zero(destType), nil
	}
	return reflect.ValueOf(nt.Time).Convert(destType), nil
}

// SqlNullPtrTimeToTime - converter for pointer sql.NullTime -> time.Time
func SqlNullPtrTimeToTime(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	nt := src.Interface().(sql.NullTime)
	if !nt.Valid {
		return reflect.Zero(destType), nil
	}
	return reflect.ValueOf(nt.Time).Convert(destType), nil
}

// TimeToSqlNullTime - converter for time.Time -> sql.NullTime
func TimeToSqlNullTime(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	t := src.Interface().(time.Time)
	return reflect.ValueOf(sql.NullTime{Time: t, Valid: !t.IsZero()}), nil
}

// NullStringToString - converter for sql.NullString -> string
func NullStringToString(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	ns := src.Interface().(sql.NullString)
	if !ns.Valid {
		return reflect.Zero(destType), nil
	}
	return reflect.ValueOf(ns.String).Convert(destType), nil
}

// SqlNullTimeToString - converter for sql.NullTime -> string
func SqlNullTimeToString(f TimeFormatter) Converter {
	return func(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
		nt := src.Interface().(sql.NullTime)
		if !nt.Valid {
			return reflect.Zero(destType), nil
		}

		s := f(nt.Time)
		return reflect.ValueOf(s).Convert(destType), nil
	}
}

// StringToSqlNullTime - converter for string -> sql.NullTime
func StringToSqlNullTime(f TimeParser) Converter {
	return func(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
		s := src.String()

		// empty string case, return invalid nullTime
		if s == "" {
			return reflect.ValueOf(sql.NullTime{Valid: false}), nil
		}

		// execute the time parser
		valid := true
		parsedTime, err := f(s)
		if err != nil {
			valid = false
		}

		nullTime := sql.NullTime{
			Time:  parsedTime,
			Valid: valid,
		}

		return reflect.ValueOf(nullTime), nil
	}
}
