package mapper

import (
	"database/sql"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Converter is a function that converts from src value to a dest reflect.Value.
// It should return the converted value (as reflect.Value) or an error.
type Converter func(src reflect.Value, destType reflect.Type) (reflect.Value, error)

type TimeFormatter func(time.Time) string

type TimeParser func(string) (time.Time, error)

var (
	convertersMu sync.RWMutex
	converters   = map[string]Converter{}
)

// RegisterConverter adds a custom converter for a specific src->dest pair.
func RegisterConverter(srcType, destType reflect.Type, conv Converter) {
	key := convKey(srcType, destType)
	convertersMu.Lock()
	converters[key] = conv
	convertersMu.Unlock()
}

func convKey(a, b reflect.Type) string {
	return a.String() + "->" + b.String()
}

// SqlNullTimeToTime - converter for sql.NullTime -> time.Time
func SqlNullTimeToTime(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
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

// StringToNullString - converter for string -> sql.NullString
func StringToNullString(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	s := src.Interface().(string)
	if s == "" {
		return reflect.ValueOf(sql.NullString{Valid: false}), nil
	}
	return reflect.ValueOf(sql.NullString{String: s, Valid: true}), nil
}

// NullInt64ToInt64 - converter for sql.NullInt64 -> int64
func NullInt64ToInt64(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	n := src.Interface().(sql.NullInt64)
	if !n.Valid {
		return reflect.Zero(destType), nil
	}
	return reflect.ValueOf(n.Int64).Convert(destType), nil
}

// Int64ToNullInt64 - converter for int64 -> sql.NullInt64
func Int64ToNullInt64(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	v := src.Interface().(int64)
	return reflect.ValueOf(sql.NullInt64{Int64: v, Valid: v != 0}), nil
}

// NullInt32ToInt32 - converter for sql.NullInt32 -> int32
func NullInt32ToInt32(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	n := src.Interface().(sql.NullInt32)
	if !n.Valid {
		return reflect.Zero(destType), nil
	}
	return reflect.ValueOf(n.Int32).Convert(destType), nil
}

// Int32ToNullInt32 - converter for int32 -> sql.NullInt32
func Int32ToNullInt32(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	v := src.Interface().(int32)
	return reflect.ValueOf(sql.NullInt32{Int32: v, Valid: v != 0}), nil
}

// NullFloat64ToFloat64 - converter for sql.NullFloat64 -> float64
func NullFloat64ToFloat64(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	n := src.Interface().(sql.NullFloat64)
	if !n.Valid {
		return reflect.Zero(destType), nil
	}
	return reflect.ValueOf(n.Float64).Convert(destType), nil
}

// Float64ToNullFloat64 - converter for float64 -> sql.NullFloat64
func Float64ToNullFloat64(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	v := src.Interface().(float64)
	return reflect.ValueOf(sql.NullFloat64{Float64: v, Valid: v != 0}), nil
}

// NullBoolToBool - converter for sql.NullBool -> bool
func NullBoolToBool(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	n := src.Interface().(sql.NullBool)
	if !n.Valid {
		return reflect.Zero(destType), nil
	}
	return reflect.ValueOf(n.Bool).Convert(destType), nil
}

// BoolToNullBool - converter for bool -> sql.NullBool
func BoolToNullBool(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	v := src.Interface().(bool)
	return reflect.ValueOf(sql.NullBool{Bool: v, Valid: true}), nil
}

// StringToSqlNullTime - converter for string -> sql.NullTime
func StringToSqlNullTime(f TimeParser) Converter {
	return func(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
		s := src.String()
		if s == "" {
			return reflect.ValueOf(sql.NullTime{Valid: false}), nil
		}
		parsedTime, err := f(s)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(sql.NullTime{Time: parsedTime, Valid: true}), nil
	}
}

// UUIDToString - converter for uuid.UUID -> string
func UUIDToString(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	u := src.Interface().(uuid.UUID)
	return reflect.ValueOf(u.String()).Convert(destType), nil
}

// StringToUUID - converter for string -> uuid.UUID
func StringToUUID(src reflect.Value, destType reflect.Type) (reflect.Value, error) {
	u, err := uuid.Parse(src.String())
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(u), nil
}
