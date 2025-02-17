package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	initGlobal "github.com/twbworld/proxy/initialize/global"
	"github.com/twbworld/proxy/initialize/system"
	"github.com/twbworld/proxy/model/common"
	"github.com/twbworld/proxy/router"
	"github.com/twbworld/proxy/utils"
)

func TestMain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initGlobal.New("config.example.yaml").Start()
	if err := system.DbStart(); err != nil {
		t.Fatal("数据库连接失败[fsj09]", err)
	}
	defer func() {
		time.Sleep(time.Second * 1) //给足够时间处理数据
		system.DbClose()
	}()

	ginServer := gin.Default()
	router.Start(ginServer)

	//以下是有执行顺序的, 并且库提前有必要数据
	testCases := [...]struct {
		method      string
		postRes     common.Response
		getRes      string
		url         string
		status      int
		postData    interface{}
		contentType string
	}{
		{method: http.MethodGet, url: "http://clash.domain.com/test.html", getRes: `proxies:
  - {"name":"外网信息复杂_理智分辨真假_www.domain.com_443","type":"vless"`},
		{method: http.MethodGet, url: "http://domain.com/test.html", getRes: utils.Base64Encode(`vless://xxxx@www.domain.com:443?encryption=none&headerType=none&sni=www.domain.com&fp=chrome&type=tcp&flow=xtls-rprx-vision&pbk=xxxx&sid=&security=reality#外网信息复杂_理智分辨真假_www.domain.com_443
vless://xxxx@x.x.x.x:443?encryption=none&headerType=none&sni=www.domain.com&fp=chrome&type=ws&alpn=h2,http/1.1&host=www.domain.com&path=/vless-ws&security=tls#外网信息复杂_理智分辨真假_x.x.x.x_443
trojan://password@www.domain.com:443?encryption=none&headerType=none&sni=www.domain.com&fp=chrome&type=ws&alpn=h2,http/1.1&host=www.domain.com&path=/trojan-go-ws/&security=tls#外网信息复杂_理智分辨真假_www.domain.com_443`)},
		{method: http.MethodGet, url: "http://domain.com/aa.html", status: http.StatusMovedPermanently, getRes: `<a href="/404.html">`},
	}

	for k, value := range testCases {
		t.Run(strconv.FormatInt(int64(k+1), 10)+value.url, func(t *testing.T) {
			if value.method == "" {
				value.method = http.MethodPost
			}
			if value.status == 0 {
				value.status = 200
			}
			if value.method == http.MethodPost {
				if value.contentType == "" {
					value.contentType = "application/json"
				}
				if value.postRes == (common.Response{}) {
					value.postRes.Code = 0
				}
			}

			requestBody := new(bytes.Buffer)
			if value.postData != nil {
				if v, ok := value.postData.(*bytes.Buffer); ok {
					requestBody = v
				} else {
					jsonVal, err := json.Marshal(value.postData)
					if err != nil {
						t.Fatal("json出错[godjg]", err)
					}
					requestBody = bytes.NewBuffer(jsonVal)
				}
			}

			b := time.Now().UnixMilli()

			//向注册的路有发起请求
			req, err := http.NewRequest(value.method, value.url, requestBody)
			if err != nil {
				t.Fatal("请求出错[godkojg]", err)
			}
			if value.method == http.MethodPost {
				req.Header.Set("content-type", value.contentType)
			}

			res := httptest.NewRecorder() // 构造一个记录
			ginServer.ServeHTTP(res, req) //模拟http服务处理请求

			result := res.Result() //response响应

			fmt.Printf("^^^^^^处理用时%d毫秒^^^^^^\n", time.Now().UnixMilli()-b)

			assert.Equal(t, value.status, result.StatusCode)

			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer result.Body.Close()

			switch value.method {
			case http.MethodPost:
				var response common.Response
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatal("返回错误", err, string(body))
				}
				assert.Equal(t, value.postRes.Code, response.Code)
			case http.MethodGet:
				assert.Contains(t, string(body), value.getRes)
			}

			// fmt.Println("request!!!!!!!!!!", string(jsonVal))
			// fmt.Println("response!!!!!!!!!!", utils.Base64Decode(string(body)))
			// fmt.Println("response!!!!!!!!!!", string(body))

			time.Sleep(time.Millisecond * 500) //!!!!!!!!!!!!!!!!!!

		})

	}

}
