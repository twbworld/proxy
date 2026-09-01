package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestBase64(t *testing.T) {
	t.Parallel()
	raw := "hello proxy"

	encoded := Base64Encode(raw + "\n")
	if encoded == "" {
		t.Fatal("Base64编码失败")
	}

	if got := Base64Decode(encoded + "\n"); got != raw {
		t.Errorf("解码预期: %s, 实际: %s", raw, got)
	}

	// 异常解码测试兜底：按原业务逻辑，解码失败原样返回
	invalid := "invalid_base64!@#"
	if got := Base64Decode(invalid); got != invalid {
		t.Errorf("异常字符串解码应当原样返回, 实际: %s", got)
	}
}

func TestHash(t *testing.T) {
	t.Parallel()
	str := "test-hash"
	expectedBytes := sha256.Sum224([]byte(str))
	expected := hex.EncodeToString(expectedBytes[:])

	if got := Hash(str); got != expected {
		t.Errorf("Hash期望: %s, 实际: %s", expected, got)
	}
	if got := HashBytes([]byte(str)); got != expected {
		t.Errorf("HashBytes期望: %s, 实际: %s", expected, got)
	}
}

func TestTimeFormat(t *testing.T) {
	t.Parallel()
	expected := "2026-08-30 12:00:00"
	ts, _ := time.ParseInLocation(time.DateTime, expected, time.UTC)

	if got := TimeFormat(ts.Unix(), time.UTC); got != expected {
		t.Errorf("格式化时间错误, 预期: %s, 实际: %s", expected, got)
	}
}

func TestNumberFormat(t *testing.T) {
	t.Parallel()
	if got := NumberFormat(3.14159); got != 3.14 {
		t.Errorf("默认保留2位小数异常, 实际: %v", got)
	}
	if got := NumberFormat(3.145, 2); got != 3.15 {
		t.Errorf("四舍五入异常, 实际: %v", got)
	}
	if got := NumberFormat(3.14159, 4); got != 3.1416 {
		t.Errorf("指定位小数异常, 实际: %v", got)
	}
}

func TestFileOps(t *testing.T) {
	t.Parallel()
	// 依托新标准库 t.TempDir 管理临时文件，运行完毕自动回收物理内存空间
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "subDir", "test.txt")

	if FileExist(targetFile) {
		t.Errorf("初始状态文件不应存在")
	}

	// 核心检测Mkdir和CreateFile
	if err := CreateFile(targetFile); err != nil {
		t.Fatalf("创建文件异常: %v", err)
	}

	if !FileExist(targetFile) {
		t.Errorf("创建后文件理应存在")
	}
}

func TestListToMap(t *testing.T) {
	t.Parallel()
	type item struct {
		Key   string
		Value int
	}
	list := []item{{"A", 1}, {"B", 2}}

	res := ListToMap(list, "Key")
	if len(res) != 2 {
		t.Fatalf("预期长度2, 实际: %d", len(res))
	}

	if val, ok := res["A"].(item); !ok || val.Value != 1 {
		t.Errorf("映射A取值失败: %v", res["A"])
	}

	// 边界测试：非切片传参
	if edgeRes := ListToMap("not a slice", "Key"); edgeRes != nil {
		t.Errorf("非切片输入应安全返回nil, 实际: %v", edgeRes)
	}
}

func TestContainsAny(t *testing.T) {
	t.Parallel()
	if !ContainsAny("user-agent: v2ray-core", []string{"clash", "v2ray"}) {
		t.Error("子串匹配异常, 理应包含 v2ray")
	}
	if ContainsAny("normal text", []string{"clash", "v2ray"}) {
		t.Error("不应匹配任何给定子串")
	}
}

func TestIntersect(t *testing.T) {
	t.Parallel()
	s1 := []int{1, 2, 3, 4, 5}
	s2 := []int{4, 5, 6, 7}

	res := Intersect(s1, s2)
	// 由于 map 提取 keys 无序，必须排序后再做深度验证
	slices.Sort(res)
	if !slices.Equal(res, []int{4, 5}) {
		t.Errorf("交集断言失败, 实际: %v", res)
	}
}

func TestReadyFile(t *testing.T) {
	t.Parallel()
	dir, file := ReadyFile("static", time.UTC, ".txt")

	if dir == "" || filepath.Ext(file) != ".txt" {
		t.Errorf("ReadyFile 生成异常: dir=%s, file=%s", dir, file)
	}
}

func TestGetTTLWithJitter(t *testing.T) {
	t.Parallel()
	ttl := GetTTLWithJitter(100)

	// 随机抖动范围为 0 ~ base/10 毫秒(向上取整数)
	if ttl < 100*time.Second || ttl >= 110*time.Second {
		t.Errorf("TTL抖动范围异常, 实际: %v", ttl)
	}
	if GetTTLWithJitter(0) != 0 {
		t.Error("值为0时应当安全无抖动")
	}
}

func TestParseDateFromLogFileName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Name     string
		Filename string
		WantDate string
		WantOK   bool
	}{
		{"正常日志后缀", "gin.log.2026-08-30", "2026-08-30", true},
		{"无日期后缀截断", "gin.log", "", false},
		{"非法日期格式", "run.log.2026-13-45", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			date, ok := ParseDateFromLogFileName(tt.Filename, time.UTC)
			if ok != tt.WantOK {
				t.Errorf("状态预期: %v, 实际: %v", tt.WantOK, ok)
			}
			if tt.WantOK && date.Format(time.DateOnly) != tt.WantDate {
				t.Errorf("解析日期预期: %s, 实际: %s", tt.WantDate, date.Format(time.DateOnly))
			}
		})
	}
}
