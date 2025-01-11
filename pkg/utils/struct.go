package utils

import (
	"cmp"
	"reflect"
	"strings"
)

func GetJsonTag(field reflect.StructField) (tag string, omitempty bool) {
	// json tag を取得
	var jsonTag string
	jsonTags := strings.Split(field.Tag.Get("json"), ",")
	if len(jsonTags) > 0 {
		jsonTag = jsonTags[0]
	}
	// omitempty があるか確認
	if len(jsonTags) > 1 {
		omitempty = jsonTags[1] == "omitempty"
	}
	return cmp.Or(jsonTag, strings.ToLower(field.Name)), omitempty
}
