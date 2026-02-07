package user

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/common"
	"github.com/twbworld/proxy/model/config"
	"github.com/twbworld/proxy/model/db"
	"github.com/twbworld/proxy/utils"
)

func TestSetProtocol(t *testing.T) {
	b := &BaseService{}
	assert.IsType(t, &clash{}, b.SetProtocol("clash"))
	assert.IsType(t, &v2ray{}, b.SetProtocol("v2ray"))
	assert.IsType(t, &v2ray{}, b.SetProtocol("unknown"))
}

func TestClashHandle(t *testing.T) {
	// 设置时区，避免 checkUser 中的时间解析错误
	global.Tz, _ = time.LoadLocation("Asia/Shanghai")
	// 设置一个未来的过期时间
	ti := time.Now().In(time.UTC).AddDate(0, 1, 0).Format(time.DateOnly)

	user := &db.Users{Quota: -1, ExpiryDate: &ti}

	// 构造一个 Reality 类型的配置
	proxy := config.Proxies{
		Name:        "test_node",
		Type:        "vless",
		Server:      "1.1.1.1",
		Port:        "443",
		Flow:        "xtls-rprx-vision",
		Network:     "tcp",
		RealityOpts: config.RealityOpts{PublicKey: "test_pbk", ShortId: "test_sid"},
	}
	global.Config.Proxies = []config.Proxies{proxy}
	global.Config.ClashPath = "test_clash.yaml"

	// 模拟 clash.yaml 模板，注意这里的占位符要和 index.go 中一致
	err := os.WriteFile(global.Config.ClashPath, []byte(`proxies: [proxies]
proxy-groups:
  - name: test
    proxies: [proxies_name]
  - name: test2
    proxies: [proxies_name]`), 0644)
	assert.NoError(t, err)
	defer os.Remove(global.Config.ClashPath)

	c := &clash{}
	result := c.Handle(user)

	// 验证结果包含节点名称列表 (JSON格式)
	assert.Contains(t, result, `proxies: ["test_node"]`)

	// 验证结果包含节点配置详情 (Clash Meta支持的JSON内嵌写法)
	// 注意：index.go 中使用的是 json.Marshal，所以 key 会有引号
	assert.Contains(t, result, fmt.Sprintf(`"server":"%s"`, proxy.Server))
	assert.Contains(t, result, fmt.Sprintf(`"flow":"%s"`, proxy.Flow))
	assert.Contains(t, result, fmt.Sprintf(`"public-key":"%s"`, proxy.RealityOpts.PublicKey))
	assert.Contains(t, result, fmt.Sprintf(`"short-id":"%s"`, proxy.RealityOpts.ShortId))

	// 验证 YAML 列表符号
	assert.Contains(t, result, "\n  - {")
}

func TestV2rayHandle(t *testing.T) {
	global.Tz, _ = time.LoadLocation("Asia/Shanghai")
	ti := time.Now().In(time.UTC).AddDate(0, 1, 0).Format(time.DateOnly)

	user := &db.Users{Quota: -1, ExpiryDate: &ti}
	// 构造一个 TCP 类型的配置
	proxy := config.Proxies{
		Type:       "vless",
		Server:     "server1",
		Port:       "443",
		Uuid:       "xxx",
		Network:    "tcp",
		Servername: "example.com",
		Tls:        true,
	}
	global.Config.Proxies = []config.Proxies{proxy}

	x := &v2ray{}
	result := x.Handle(user)
	decodedResult := utils.Base64Decode(result)

	// 验证基础链接格式
	expectedPrefix := fmt.Sprintf("%s://%s@%s:%s?", proxy.Type, proxy.Uuid, proxy.Server, proxy.Port)
	assert.Contains(t, decodedResult, expectedPrefix)

	// 验证参数
	assert.Contains(t, decodedResult, "encryption=none")
	assert.Contains(t, decodedResult, "security=tls")
	assert.Contains(t, decodedResult, "sni=example.com")
	assert.Contains(t, decodedResult, "type=tcp")
	assert.Contains(t, decodedResult, "headerType=none")
}

func TestClashGetConfig(t *testing.T) {
	c := &clash{}

	// Case 1: VLESS Reality
	proxyReality := config.Proxies{
		Name:        "RealityNode",
		Type:        "vless",
		Network:     "tcp",
		Flow:        "xtls-rprx-vision",
		RealityOpts: config.RealityOpts{PublicKey: "pk"},
	}
	resReality, nameReality := c.getConfig(&proxyReality)
	assert.Equal(t, "RealityNode", nameReality)
	assert.IsType(t, common.ClashVlessReality{}, resReality)

	// Case 2: VLESS gRPC
	proxyGrpc := config.Proxies{
		Name:    "GrpcNode",
		Type:    "vless",
		Network: "grpc",
	}
	resGrpc, nameGrpc := c.getConfig(&proxyGrpc)
	assert.Equal(t, "GrpcNode", nameGrpc)
	assert.IsType(t, common.ClashVlessGrpc{}, resGrpc)

	// Case 3: VLESS TCP (Base)
	proxyTcp := config.Proxies{
		Name:    "TcpNode",
		Type:    "vless",
		Network: "tcp",
	}
	resTcp, nameTcp := c.getConfig(&proxyTcp)
	assert.Equal(t, "TcpNode", nameTcp)
	assert.IsType(t, common.ClashVlessBase{}, resTcp)

	// Case 4: XHTTP (Clash 暂不支持，应返回 nil)
	proxyXhttp := config.Proxies{
		Name:    "XhttpNode",
		Type:    "vless",
		Network: "xhttp",
	}
	resXhttp, nameXhttp := c.getConfig(&proxyXhttp)
	assert.Nil(t, resXhttp)
	assert.Empty(t, nameXhttp)
}

func TestV2rayGetConfig(t *testing.T) {
	x := &v2ray{}

	// Case 1: Standard VLESS
	proxy := config.Proxies{
		Type:              "vless",
		Server:            "server1",
		Port:              "443",
		Uuid:              "uuid1",
		ClientFingerprint: "chrome",
		Name:              "test",
		Network:           "tcp",
	}
	result := x.getConfig(&proxy)
	assert.Contains(t, result, "vless://uuid1@server1:443")
	assert.Contains(t, result, "fp=chrome")
	assert.Contains(t, result, "#test")

	// Case 2: XHTTP with Extra params
	extraData := map[string]interface{}{
		"downloadSettings": map[string]interface{}{"address": "1.1.1.1"},
		"xhttpSettings":    map[string]interface{}{"path": "/v7"},
	}
	proxyXhttp := config.Proxies{
		Type:    "vless",
		Server:  "server2",
		Port:    "443",
		Uuid:    "uuid2",
		Network: "xhttp",
		XhttpOpts: config.XhttpOpts{
			Mode:  "auto",
			Path:  "/path",
			Extra: extraData,
		},
	}
	resultXhttp := x.getConfig(&proxyXhttp)

	assert.Contains(t, resultXhttp, "type=xhttp")
	assert.Contains(t, resultXhttp, "mode=auto")
	assert.Contains(t, resultXhttp, "path=/path")

	// 验证 Extra 字段是否被正确 JSON 序列化并 URL 编码
	extraJson, _ := json.Marshal(extraData)
	encodedExtra := url.QueryEscape(string(extraJson))
	assert.Contains(t, resultXhttp, "&extra="+encodedExtra)
}

func TestCheckUser(t *testing.T) {
	global.Tz, _ = time.LoadLocation("Asia/Shanghai")
	ti := time.Now().In(time.UTC)
	t1 := ti.AddDate(0, 1, 0).Format(time.DateOnly) // 未来
	t2 := ti.AddDate(0, -1, 0).Format(time.DateOnly) // 过去

	// 未过期
	user := &db.Users{ExpiryDate: &t1}
	assert.True(t, checkUser(user))

	// 已过期
	user.ExpiryDate = &t2
	assert.False(t, checkUser(user))

	// 空字符串视作不过期 (根据 checkUser 逻辑)
	empty := ""
	user.ExpiryDate = &empty
	assert.True(t, checkUser(user))

	zero := "0"
	user.ExpiryDate = &zero
	assert.True(t, checkUser(user))
}
