package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/twbworld/proxy/global"
	initGlobal "github.com/twbworld/proxy/initialize/global"
	"github.com/twbworld/proxy/initialize/system"
	"github.com/twbworld/proxy/model/config"
	"github.com/twbworld/proxy/router"
	"github.com/twbworld/proxy/utils"
)

func TestMain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. 初始化配置
	initGlobal.New("config.example.yaml").Start()

	// [关键修改]：重定向 ClashPath 到临时测试文件，防止覆盖生产环境的 clash.yaml
	global.Config.ClashPath = "clash.test.yaml"

	// 测试结束时清理临时文件
	defer func() {
		_ = os.Remove(global.Config.ClashPath)
	}()

	// 2. 注入测试用的代理配置 (覆盖文件中的配置，确保测试确定性)
	injectTestProxies()

	// 3. 启动数据库 (假设环境已配置好或使用 SQLite)
	if err := system.DbStart(); err != nil {
		t.Log("警告: 数据库连接失败, 部分依赖数据库的测试可能无法通过:", err)
	} else {
		defer system.DbClose()
	}

	ginServer := gin.Default()
	router.Start(ginServer)

	// 构造 XHTTP extra 参数的预期编码字符串 (融合 downloadSettings)
	extraData := map[string]any{
		"a": "b",
		"downloadSettings": map[string]any{
			"address":  "4.4.4.4",
			"port":     443,
			"network":  "xhttp",
			"security": "tls",
		},
		"mode": "auto",
	}
	extraBytes, _ := json.Marshal(extraData)
	encodedExtra := url.QueryEscape(string(extraBytes))

	// 定义测试用例
	testCases := []struct {
		name        string
		method      string
		url         string
		host        string
		userAgent   string
		status      int
		shouldExist []string
		notExist    []string
	}{
		{
			name:      "Clash订阅-VLESS Reality & gRPC & XHTTP",
			method:    http.MethodGet,
			url:       "http://clash.domain.com/test.html",
			host:      "clash.domain.com",
			userAgent: "Clash.Meta",
			status:    http.StatusOK,
			shouldExist: []string{
				"proxies:",
				`"name":"Test_Reality"`,
				`"type":"vless"`,
				`"server":"1.1.1.1"`,
				`"flow":"xtls-rprx-vision"`,
				`"public-key":"test_pbk"`,
				`"short-id":"test_sid"`,
				// 验证 gRPC 节点
				`"name":"Test_Grpc"`,
				`"network":"grpc"`,
				`"grpc-service-name":"grpc_service"`,
				// 验证 XHTTP 节点
				`"name":"Test_Xhttp"`,
				`"network":"xhttp"`,
				`"download-settings":`,
				`"server":"4.4.4.4"`,
				// 验证 proxy-groups 包含节点名
				`"Test_Reality"`,
				`"Test_Grpc"`,
				`"Test_Xhttp"`,
			},
			notExist: []string{},
		},
		{
			name:      "v2rayN订阅-所有协议",
			method:    http.MethodGet,
			url:       "http://domain.com/test.html",
			host:      "www.domain.com",
			userAgent: "v2rayN",
			status:    http.StatusOK,
			shouldExist: []string{
				"vless://uuid1@1.1.1.1:443",
				"security=reality",
				"pbk=test_pbk",
				"sid=test_sid",
				"flow=xtls-rprx-vision",
				"#Test_Reality",

				"vless://uuid2@2.2.2.2:443",
				"mode=gun",
				"serviceName=grpc_service",
				"authority=grpc.com",
				"#Test_Grpc",

				// 验证 XHTTP 以及转换至 extra 里的内容
				"vless://uuid3@3.3.3.3:443",
				"type=xhttp",
				"mode=auto",
				"path=%2Fxhttp",
				"extra=" + encodedExtra,
				"#Test_Xhttp",
			},
		},
		{
			name:   "404页面测试-不存在的文件",
			method: http.MethodGet,
			url:    "http://domain.com/aa.html",
			host:   "domain.com",
			status: http.StatusMovedPermanently,
			shouldExist: []string{
				"Moved",
			},
		},
		{
			name:   "404页面测试-非html",
			method: http.MethodGet,
			url:    "http://domain.com/aa",
			host:   "domain.com",
			status: http.StatusOK,
			shouldExist: []string{
				"404",
			},
		},
	}

	for i, value := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i+1, value.name), func(t *testing.T) {
			b := time.Now().UnixMilli()

			req, err := http.NewRequest(value.method, value.url, nil)
			if err != nil {
				t.Fatal("构造请求出错:", err)
			}

			if value.host != "" {
				req.Host = value.host
			}
			if value.userAgent != "" {
				req.Header.Set("User-Agent", value.userAgent)
			}

			res := httptest.NewRecorder()
			ginServer.ServeHTTP(res, req)

			result := res.Result()
			defer result.Body.Close()

			fmt.Printf("--- 测试用例 [%s] 耗时: %dms ---\n", value.name, time.Now().UnixMilli()-b)

			assert.Equal(t, value.status, result.StatusCode)

			bodyBytes, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatal(err)
			}
			bodyString := string(bodyBytes)

			checkContent := bodyString
			if strings.Contains(value.name, "v2ray") && !strings.Contains(bodyString, "html") && !strings.Contains(bodyString, "Moved") {
				decoded := utils.Base64Decode(bodyString)
				if decoded != bodyString {
					checkContent = decoded
				}
			}

			for _, expect := range value.shouldExist {
				assert.Contains(t, checkContent, expect, "响应内容应包含: "+expect)
			}

			for _, notExpect := range value.notExist {
				assert.NotContains(t, checkContent, notExpect, "响应内容不应包含: "+notExpect)
			}
		})
	}
}

func injectTestProxies() {
	global.Config.Proxies = []config.Proxies{
		{
			Name:       "Test_Reality",
			Type:       "vless",
			Server:     "1.1.1.1",
			Port:       "443",
			Uuid:       "uuid1",
			Flow:       "xtls-rprx-vision",
			Network:    "tcp",
			Servername: "reality.com",
			RealityOpts: config.RealityOpts{
				PublicKey: "test_pbk",
				ShortId:   "test_sid",
			},
		},
		{
			Name:       "Test_Grpc",
			Type:       "vless",
			Server:     "2.2.2.2",
			Port:       "443",
			Uuid:       "uuid2",
			Network:    "grpc",
			Servername: "grpc.com",
			Tls:        true,
			GrpcOpts: config.GrpcOpts{
				GrpcServiceName: "grpc_service",
			},
		},
		{
			Name:       "Test_Xhttp",
			Type:       "vless",
			Server:     "3.3.3.3",
			Port:       "443",
			Uuid:       "uuid3",
			Network:    "xhttp",
			Servername: "xhttp.com",
			Tls:        true,
			XhttpOpts: config.XhttpOpts{
				Mode: "auto",
				Path: "/xhttp",
				DownloadSettings: &config.DownloadSettings{
					Server: "4.4.4.4",
					Port:   "443",
				},
				Extra: map[string]any{
					"a":    "b",
					"mode": "auto",
				},
			},
		},
		{
			Name: "Test_Ignored_Root",
			Type: "vless",
			Root: true,
		},
	}

	_ = utils.CreateFile(global.Config.ClashPath)
	_ = writeFile(global.Config.ClashPath, []byte(`proxies: [proxies]
proxy-groups:
  - name: test
    proxies: [proxies_name]`), 0644)
}

func writeFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}
