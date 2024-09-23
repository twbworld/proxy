package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code  int8        `json:"code"`
	Data  interface{} `json:"data"`
	Msg   string      `json:"msg"`
	Token string      `json:"token,omitempty"`
}

const (
	successCode       = 0
	errorCode         = 1
	defaultSuccessMsg = `ok`
	defaultFailMsg    = `错误`
)

func result(ctx *gin.Context, code int8, msg string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

// 带data
func Success(ctx *gin.Context, data interface{}) {
	result(ctx, successCode, defaultSuccessMsg, data)
}

// 带msg,不带data
func SuccessOk(ctx *gin.Context, message string) {
	result(ctx, successCode, message, map[string]interface{}{})
}

func Fail(ctx *gin.Context, message string) {
	result(ctx, errorCode, message, map[string]interface{}{})
}

func FailNotFound(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, Response{
		Code: errorCode,
		Msg:  defaultFailMsg,
	})
}
