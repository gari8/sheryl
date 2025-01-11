package log

import (
	"github.com/gari8/sheryl/pkg/utils"
	"log/slog"
	"reflect"
)

type Attr[T any] struct {
	Val reflect.Value
	Typ reflect.Type
}

func NewAttr[T any](typ T) *Attr[T] {
	return &Attr[T]{
		Val: reflect.ValueOf(typ),
		Typ: reflect.TypeOf(typ),
	}
}

func (a *Attr[T]) Add(attrs ...any) []any {
	var val reflect.Value
	var typ reflect.Type
	if a.Val.Kind() == reflect.Ptr {
		val = a.Val.Elem()
		typ = a.Typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName, omitempty := utils.GetJsonTag(field)
		fieldValue := val.Field(i)
		if fieldValue.IsZero() && omitempty {
			continue
		}
		attrs = append(attrs, slog.Any(fieldName, fieldValue.Interface()))
	}
	return attrs
}
