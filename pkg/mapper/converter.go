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

// sqlNullTimeToTime - converter for sql.NullTime -> time.Time
func sqlNullTimeToTime(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	nt := src.Interface().(sql.NullTime)
	if !nt.Valid {
		return reflect.Zero(destType), nil
	}
	return reflect.ValueOf(nt.Time).Convert(destType), nil
}

// sqlNullPtrTimeToTime - converter for pointer sql.NullTime -> time.Time
func sqlNullPtrTimeToTime(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	nt := src.Interface().(sql.NullTime)
	if !nt.Valid {
		return reflect.Zero(destType), nil
	}
	return reflect.ValueOf(nt.Time).Convert(destType), nil
}

// timeToSqlNullTime - converter for time.Time -> sql.NullTime
func timeToSqlNullTime(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	t := src.Interface().(time.Time)
	return reflect.ValueOf(sql.NullTime{Time: t, Valid: !t.IsZero()}), nil
}

// nullStringToString - converter for sql.NullString -> string
func nullStringToString(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	ns := src.Interface().(sql.NullString)
	if !ns.Valid {
		return reflect.Zero(destType), nil
	}
	return reflect.ValueOf(ns.String).Convert(destType), nil
}

// sqlNullTimeToString - converter for sql.NullTime -> string
func sqlNullTimeToString(f TimeFormatter) Converter {
	return func(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
		nt := src.Interface().(sql.NullTime)
		if !nt.Valid {
			return reflect.Zero(destType), nil
		}

		s := f(nt.Time)
		return reflect.ValueOf(s).Convert(destType), nil
	}
}
