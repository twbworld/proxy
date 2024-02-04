package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/router"
	"github.com/twbworld/proxy/utils"
)

func TestMain(t *testing.T) {
	global.Init()

	testCases := []struct{
		name string
		status int
		input string
		res string
	}{
		{name: "successTest", status: http.StatusOK, input: "test", res: "vless://"},
		{name: "failTest", status: http.StatusMovedPermanently, input: "aa", res: "<a href="},
	}

	defer func() {
		if p := recover(); p != nil {
			global.Log.Println(p)
		}
		if dao.DB != nil {
			err := dao.Close()
			if err != nil {
				global.Log.Println("数据库关闭出错[fyhkg]", err)
			}
		}
	}()

	dao.Init()

	ginServer := gin.Default()
	router.Init(ginServer)

	gin.SetMode(gin.TestMode)

	for _, value := range testCases{

		t.Run(value.name, func(t *testing.T){
			req, err := http.NewRequest(http.MethodGet, "http://localhost:80/" + value.input + ".html", nil)
			if err != nil {
				log.Fatalf("报错: %v", err)
			}

			res := httptest.NewRecorder()
			ginServer.ServeHTTP(res, req)

			result := res.Result()
			assert.Equal(t, value.status, result.StatusCode)

			resB, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatal(err)
			}

			defer result.Body.Close()

			assert.Equal(t, value.res, utils.Base64Decode(string(resB))[:8])

		})

	}



}
