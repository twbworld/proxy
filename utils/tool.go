package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"reflect"
	"strings"
)


func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Trim(strings.TrimSpace(str), "\n")))
}
func Hash(str string) string {
	b := sha256.Sum224([]byte(str))
	return hex.EncodeToString(b[:])
}


func ListToMap(list interface{}, key string) map[string]interface{} {

	v := reflect.ValueOf(list)
	res := make(map[string]interface{})
	data := make([]interface{}, 0)
	if v.Kind() != reflect.Slice {
		data = append(data, list)
	}else {
		for i := 0; i < v.Len(); i++ {
			data = append(data, v.Index(i).Interface())
		}
	}

	for _, value := range data{
		val := reflect.ValueOf(value)
		res[val.FieldByName(key).String()] = value
	}

	return res
}
