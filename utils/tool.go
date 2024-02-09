package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(strings.Trim(str, "\n"))))
}
func Base64Decode(str string) string {
	bstr, err := base64.StdEncoding.DecodeString(strings.TrimSpace(strings.Trim(str, "\n")))
	if err != nil {
		return str
	}
	return string(bstr)
}
func Hash(str string) string {
	b := sha256.Sum224([]byte(str))
	return hex.EncodeToString(b[:])
}

// 类似php的array_column($a, null, 'key')
func ListToMap(list interface{}, key string) map[string]interface{} {

	data := make([]interface{}, 0)
	if v := reflect.ValueOf(list); v.Kind() != reflect.Slice {
		data = append(data, list)
	} else {
		for i := range v.Len() {
			data = append(data, v.Index(i).Interface())
		}
	}

	res := make(map[string]interface{}, len(data))
	for _, value := range data {
		res[reflect.ValueOf(value).FieldByName(key).String()] = value
	}

	return res
}

func CreateFile(path string) (err error) {
	file, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		paths, _ := filepath.Split(path)

		if _, err = os.Stat(paths); err != nil {
			if err = os.MkdirAll(paths, os.ModePerm); err != nil {
				return
			}
		}

		fi, e := os.Create(path)
		if e != nil {
			return e
		}
		fi.Close()

		return
	}

	file.Close()

	return
}

func InSlice(slice *[]string, value string) int {
	for i, item := range *slice {
		if item == value {
			return i
		}
	}
	return -1
}
