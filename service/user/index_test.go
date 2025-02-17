package user

import (
	"fmt"
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
	assert.IsType(t, &xray{}, b.SetProtocol("xray"))
	assert.IsType(t, &xray{}, b.SetProtocol("unknown"))
}

func TestClashHandle(t *testing.T) {
	global.Tz, _ = time.LoadLocation("Asia/Shanghai")
	ti := time.Now().In(global.Tz).AddDate(0, 1, 0).Format(time.DateOnly)

	user := &db.Users{Quota: -1, ExpiryDate: &ti}
	proxy := config.Proxy{Type: "vless", Server: "server1", Port: "443", Flow: "xtls-rprx-vision", RealityOpts: config.RealityOpts{PublicKey: "xxx"}}
	global.Config.Proxy = []config.Proxy{proxy}
	global.Config.ClashPath = "test_clash.yaml"
	os.WriteFile(global.Config.ClashPath, []byte(`[proxies_name] [proxies]`), 0644)
	defer os.Remove(global.Config.ClashPath)

	c := &clash{}
	result := c.Handle(user)
	assert.Contains(t, result, fmt.Sprintf(`["外网信息复杂_理智分辨真假_%s_%s"]
  - {`, proxy.Server, proxy.Port))
	assert.Contains(t, result, fmt.Sprintf(`"flow":"%s"`, proxy.Flow))
	assert.Contains(t, result, fmt.Sprintf(`"reality-opts":{"public-key":"%s"`, proxy.RealityOpts.PublicKey))
}

func TestXrayHandle(t *testing.T) {
	global.Tz, _ = time.LoadLocation("Asia/Shanghai")
	ti := time.Now().In(global.Tz).AddDate(0, 1, 0).Format(time.DateOnly)

	user := &db.Users{Quota: -1, ExpiryDate: &ti}
	proxy := config.Proxy{Type: "vless", Server: "server1", Port: "443", Uuid: "xxx", Network: "ws", WsOpts: config.WsOpts{Path: "xx"}}
	global.Config.Proxy = []config.Proxy{proxy}

	x := &xray{}
	result := x.Handle(user)
	assert.Contains(t, utils.Base64Decode(result), fmt.Sprintf("%s://%s@%s:%s?", proxy.Type, proxy.Uuid, proxy.Server, proxy.Port))
	assert.Contains(t, utils.Base64Decode(result), fmt.Sprintf("type=%s", proxy.Network))
}

func TestClashGetConfig(t *testing.T) {
	c := &clash{}
	proxy := config.Proxy{Type: "vless", Network: "ws", WsOpts: config.WsOpts{Path: "/ws"}}
	result := c.getConfig(&proxy)
	assert.IsType(t, common.ClashVlessWs{}, result)
}

func TestXrayGetConfig(t *testing.T) {
	x := &xray{}
	proxy := config.Proxy{Type: "vless", Server: "server1", Port: "443", Uuid: "uuid1"}
	result := x.getConfig(&proxy)
	assert.Contains(t, result, "vless://uuid1@server1:443")
}

func TestCheckUser(t *testing.T) {
	global.Tz, _ = time.LoadLocation("Asia/Shanghai")
	ti := time.Now().In(global.Tz)
	t1 := ti.Format(time.DateOnly)
	t2 := ti.AddDate(0, -1, 0).Format(time.DateOnly)

	user := &db.Users{ExpiryDate: &t1}
	assert.True(t, checkUser(user))

	user.ExpiryDate = &t2
	assert.False(t, checkUser(user))
}
