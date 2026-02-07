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
	// 注意：在CI/CD环境中可能需要mock数据库，这里假设本地环境可用
	if err := system.DbStart(); err != nil {
		t.Log("警告: 数据库连接失败, 部分依赖数据库的测试可能无法通过:", err)
	} else {
		defer system.DbClose()
	}

	ginServer := gin.Default()
	router.Start(ginServer)

	// 构造 XHTTP extra 参数的预期编码字符串
	extraData := map[string]interface{}{"a": "b", "mode": "auto"}
	extraBytes, _ := json.Marshal(extraData)
	encodedExtra := url.QueryEscape(string(extraBytes))

	// 定义测试用例
	testCases := []struct {
		name        string
		method      string
		url         string // 请求 URL
		host        string // 模拟的 Host Header (用于触发 controller 的协议判断)
		userAgent   string // 模拟 User-Agent
		status      int
		shouldExist []string // 响应中应该包含的字符串
		notExist    []string // 响应中不应该包含的字符串
	}{
		{
			name:      "Clash订阅-VLESS Reality & gRPC",
			method:    http.MethodGet,
			url:       "http://clash.domain.com/test.html",
			host:      "clash.domain.com", // 触发 Clash 逻辑
			userAgent: "Clash.Meta",
			status:    http.StatusOK,
			shouldExist: []string{
				// 验证 YAML 结构
				"proxies:",
				// 验证 Reality 节点 (Clash配置中使用了JSON嵌入)
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
				// 验证 proxy-groups 包含节点名
				`"Test_Reality"`,
				`"Test_Grpc"`,
			},
			notExist: []string{
				// XHTTP 不被 Clash 支持，应该被过滤
				`"name":"Test_Xhttp"`,
				`"network":"xhttp"`,
				"Test_Xhttp", // 组里也不应该有
			},
		},
		{
			name:      "v2rayN订阅-所有协议",
			method:    http.MethodGet,
			url:       "http://domain.com/test.html",
			host:      "www.domain.com", // 触发 Xray 逻辑 (默认)
			userAgent: "v2rayN",
			status:    http.StatusOK,
			shouldExist: []string{
				// Reality
				"vless://uuid1@1.1.1.1:443",
				"security=reality",
				"pbk=test_pbk",
				"sid=test_sid",
				"flow=xtls-rprx-vision",
				"#Test_Reality",
				// gRPC
				"vless://uuid2@2.2.2.2:443",
				"mode=gun",
				"serviceName=grpc_service",
				"authority=grpc.com",
				"#Test_Grpc",
				// XHTTP
				"vless://uuid3@3.3.3.3:443",
				"type=xhttp",
				"mode=auto",
				"path=/xhttp",
				"extra=" + encodedExtra, // 验证 JSON 序列化 + URL 编码
				"#Test_Xhttp",
			},
		},
		{
			name:   "404页面测试-不存在的文件",
			method: http.MethodGet,
			url:    "http://domain.com/aa.html",
			host:   "domain.com",
			status: http.StatusMovedPermanently, // 路由中定义的行为
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

			// 关键：设置 Host 以触发 base.go 中的 SetProtocol 逻辑
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

			// 1. 验证状态码
			assert.Equal(t, value.status, result.StatusCode)

			// 2. 读取响应体
			bodyBytes, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatal(err)
			}
			bodyString := string(bodyBytes)

			// 3. 特殊处理：如果是 v2ray 订阅，且看起来是 Base64，则先解码再验证
			// 注意：404 页面不是 Base64
			checkContent := bodyString
			if strings.Contains(value.name, "v2ray") && !strings.Contains(bodyString, "html") && !strings.Contains(bodyString, "Moved") {
				decoded := utils.Base64Decode(bodyString)
				// 简单的判断解码是否成功
				if decoded != bodyString {
					checkContent = decoded
				}
			}

			// 4. 验证包含内容
			for _, expect := range value.shouldExist {
				assert.Contains(t, checkContent, expect, "响应内容应包含: "+expect)
			}

			// 5. 验证不包含内容
			for _, notExpect := range value.notExist {
				assert.NotContains(t, checkContent, notExpect, "响应内容不应包含: "+notExpect)
			}
		})
	}
}

// injectTestProxies 注入模拟的节点数据
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
				Extra: map[string]interface{}{
					"a":    "b",
					"mode": "auto",
				},
			},
		},
		{
			Name: "Test_Ignored_Root",
			Type: "vless",
			Root: true, // Root 节点，非管理员不应显示
		},
	}

	// 使用 CreateFile 创建文件 (此时 global.Config.ClashPath 已经被修改为 clash.test.yaml)
	_ = utils.CreateFile(global.Config.ClashPath)
	_ = writeFile(global.Config.ClashPath, []byte(`proxies: [proxies]
proxy-groups:
  - name: test
    proxies: [proxies_name]`), 0644)
}

// 辅助函数：简单的写文件，避免引入过多依赖
func writeFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}
