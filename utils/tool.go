package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"maps"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"
	"uuid"
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
	return HashBytes([]byte(str))
}

func HashBytes(data []byte) string {
	b := sha256.Sum224(data)
	return hex.EncodeToString(b[:])
}

func TimeFormat[T timeNumber](t T, loc *time.Location) string {
	return time.Unix(int64(t), 0).In(loc).Format(time.DateTime)
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

// 类似php的array_column($a, null, 'key')
func ListToMap(list any, key string) map[string]any {
	v := reflect.ValueOf(list)
	if v.Kind() != reflect.Slice {
		return nil
	}

	res := make(map[string]any, v.Len())
	for i := range v.Len() {
		item := v.Index(i).Interface()
		itemValue := reflect.ValueOf(item)
		keyValue := itemValue.FieldByName(key)
		if keyValue.IsValid() && keyValue.Kind() == reflect.String {
			if strVal, ok := reflect.TypeAssert[string](keyValue); ok {
				res[strVal] = item
			} else {
				res[keyValue.String()] = item
			}
		}
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
func Intersect[T string | Number](slice1, slice2 []T) []T {
	// 创建一个空的哈希集合用于存储第一个切片的元素
	set1 := make(map[T]struct{})
	for _, elem := range slice1 {
		set1[elem] = struct{}{}
	}

	// 创建一个空的哈希集合用于存储交集
	intersectionSet := make(map[T]struct{})
	for _, elem := range slice2 {
		if _, exists := set1[elem]; exists {
			intersectionSet[elem] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(intersectionSet))
}

// 生成文件路径和文件名
func ReadyFile(staticDir string, loc *time.Location, fileExt ...string) (string, string) {
	ext := ""
	if len(fileExt) > 0 {
		ext = fileExt[0]
	}

	return filepath.Join(staticDir, time.Now().In(loc).Format("2006/01/")) + "/", uuid.NewV7().String()[:10] + ext
}

// GetTTLWithJitter 为缓存TTL增加随机抖动，防止缓存雪崩
func GetTTLWithJitter(baseTTLInSeconds int64) time.Duration {
	if baseTTLInSeconds <= 0 {
		return 0
	}
	jitter := rand.Int64N(baseTTLInSeconds / 10)
	return time.Duration(baseTTLInSeconds+jitter) * time.Second
}

// ParseDateFromLogFileName 从日志文件名中解析日期
// 文件名格式如: gin.log.2025-10-28, run.log.2025-10-28
func ParseDateFromLogFileName(filename string, loc *time.Location) (time.Time, bool) {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return time.Time{}, false
	}

	// 日期部分应在最后
	dateStr := parts[len(parts)-1]
	// 使用 "2006-01-02" 格式解析
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
