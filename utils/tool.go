package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/twbworld/proxy/global"
)

type timeNumber interface {
	~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

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

func TimeFormat[T timeNumber](t T) string {
	return time.Unix(int64(t), 0).In(global.Tz).Format(time.DateTime)
}

// 四舍五入保留小数位
func NumberFormat[T ~float32 | ~float64](f T, n ...uint) float64 {
	num := uint(2)
	if len(n) > 0 {
		num = n[0]
	}
	nu := math.Pow(10, float64(num))
	return math.Round(float64(f)*nu) / nu
}

// 文件是否存在
func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// 创建目录
func Mkdir(path string) error {
	// 从路径中取目录
	dir := filepath.Dir(path)
	// 获取信息, 即判断是否存在目录
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// 生成目录
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// 创建文件
// 可能存在跨越目录创建文件的风险
func CreateFile(path string) error {
	if FileExist(path) {
		return nil
	}

	if err := Mkdir(path); err != nil {
		return err
	}

	fi, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fi.Close()

	return nil
}

// func ListToMap(list any, key string) map[string]any {
// 	v := reflect.ValueOf(list)
// 	if v.Kind() != reflect.Slice {
// 		return nil
// 	}

// 	res := make(map[string]any, v.Len())
// 	for i := 0; i < v.Len(); i++ {
// 		item := v.Index(i).Interface()
// 		itemValue := reflect.ValueOf(item)
// 		keyValue := itemValue.FieldByName(key)
// 		if keyValue.IsValid() && keyValue.Kind() == reflect.String {
// 			res[keyValue.String()] = item
// 		}
// 	}

//		return res
//	}
//
// ListToMap 转换为map(使用泛型替代反射重构)
// 类似php的array_column($a, null, 'key')
func ListToMap[E any, K comparable](list []E, keyFunc func(E) K) map[K]E {
	res := make(map[K]E, len(list))
	for _, v := range list {
		res[keyFunc(v)] = v
	}
	return res
}

// 判断一个字符串是否包含多个子字符串中的任意一个
func ContainsAny(str string, substrs []string) bool {
	return slices.ContainsFunc(substrs, func(s string) bool {
		return strings.Contains(str, s)
	})
}

// 取两个切片的交集
func Union[T comparable](slice1, slice2 []T) []T {
	set1 := make(map[T]struct{}, len(slice1))
	for _, elem := range slice1 {
		set1[elem] = struct{}{}
	}

	var result []T
	for _, elem := range slice2 {
		if _, exists := set1[elem]; exists {
			result = append(result, elem)
		}
	}
	return slices.Clip(result)
}
