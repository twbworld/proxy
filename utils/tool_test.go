package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/twbworld/proxy/global"
)

func TestBase64EncodeDecode(t *testing.T) {
	original := "hello world"
	encoded := Base64Encode(original)
	decoded := Base64Decode(encoded)
	assert.Equal(t, original, decoded)
}

func TestHash(t *testing.T) {
	original := "hello world"
	hashed := Hash(original)
	expected := "2f05477fc24bb4faefd86517156dafdecec45b8ad3cf2522a563582b"
	assert.Equal(t, expected, hashed)
}

func TestTimeFormat(t *testing.T) {
	global.Tz, _ = time.LoadLocation("Asia/Shanghai")
	timestamp := int64(1700000000)
	formatted := TimeFormat(timestamp)
	expected := time.Unix(timestamp, 0).In(global.Tz).Format(time.DateTime)
	assert.Equal(t, expected, formatted)
}

func TestNumberFormat(t *testing.T) {
	number := 123.456789
	formatted := NumberFormat(number, 2)
	expected := 123.46
	assert.Equal(t, expected, formatted)
}

func TestFileExist(t *testing.T) {
	path := "testfile.txt"
	file, err := os.Create(path)
	assert.NoError(t, err)
	file.Close()
	defer os.Remove(path)
	assert.True(t, FileExist(path))
}

func TestMkdirAndCreateFile(t *testing.T) {
	dir := "testdir"
	filePath := filepath.Join(dir, "testfile.txt")
	defer os.RemoveAll(dir)

	err := Mkdir(filePath)
	assert.NoError(t, err)
	assert.True(t, FileExist(dir))

	err = CreateFile(filePath)
	assert.NoError(t, err)
	assert.True(t, FileExist(filePath))
}

func TestListToMap(t *testing.T) {
	type Item struct {
		Key   string
		Value string
	}
	list := []Item{
		{Key: "a", Value: "1"},
		{Key: "b", Value: "2"},
	}
	result := ListToMap(list, "Key")
	expected := map[string]interface{}{
		"a": Item{Key: "a", Value: "1"},
		"b": Item{Key: "b", Value: "2"},
	}
	assert.Equal(t, expected, result)
}

func TestInSlice(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.Equal(t, 1, InSlice(slice, "b"))
	assert.Equal(t, -1, InSlice(slice, "d"))
}

func TestUnion(t *testing.T) {
	// 测试整数切片
	intSlice1 := []int{1, 2, 3, 4}
	intSlice2 := []int{3, 4, 5, 6}
	expectedIntResult := []int{3, 4}
	assert.ElementsMatch(t, expectedIntResult, Union(intSlice1, intSlice2))

	// 测试字符串切片
	strSlice1 := []string{"a", "b", "c"}
	strSlice2 := []string{"b", "c", "d"}
	expectedStrResult := []string{"b", "c"}
	assert.ElementsMatch(t, expectedStrResult, Union(strSlice1, strSlice2))

	// 测试无交集的情况
	noIntersectionSlice1 := []int{1, 2}
	noIntersectionSlice2 := []int{3, 4}
	expectedNoIntersectionResult := []int{}
	assert.ElementsMatch(t, expectedNoIntersectionResult, Union(noIntersectionSlice1, noIntersectionSlice2))

	// 测试空切片的情况
	emptySlice := []int{}
	expectedEmptyResult := []int{}
	assert.ElementsMatch(t, expectedEmptyResult, Union(emptySlice, noIntersectionSlice2))
	assert.ElementsMatch(t, expectedEmptyResult, Union(noIntersectionSlice1, emptySlice))
	assert.ElementsMatch(t, expectedEmptyResult, Union(emptySlice, emptySlice))
}
