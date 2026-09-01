package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/twbworld/proxy/model/enum"
)

type Response struct {
	Code  enum.ResCode `json:"code"`
	Data  any          `json:"data"`
	Msg   enum.Msg     `json:"msg"`
	Token string       `json:"token,omitzero"`
}

func (r *Response) WithData[T any](data T) *Response {
	r.Data = data
	return r
}

func (r *Response) WithToken(token string) *Response {
	r.Token = token
	return r
}

// 带data
func Success[T any](ctx *gin.Context, data T) {
	resp := &Response{Code: enum.SuccessCode, Msg: enum.DefaultSuccessMsg}
	ctx.JSON(http.StatusOK, resp.WithData(data))
}

func SuccessWs[T any](c chan *Response, data T) {
	resp := &Response{Code: enum.SuccessCode, Msg: enum.DefaultSuccessMsg}
	c <- resp.WithData(data)
}

// 带msg,不带data
func SuccessOk(ctx *gin.Context, message string) {
	resp := &Response{Code: enum.SuccessCode, Msg: enum.Msg(message)}
	ctx.JSON(http.StatusOK, resp.WithData(map[string]any{}))
}

func SuccessAuth[T any](ctx *gin.Context, token string, data T) {
	resp := &Response{Code: enum.SuccessCode, Msg: enum.DefaultSuccessMsg}
	ctx.JSON(http.StatusOK, resp.WithData(data).WithToken(token))
}

func SuccessAuthWs(c chan *Response, token string) {
	resp := &Response{Code: enum.SuccessCode, Msg: enum.DefaultSuccessMsg}
	c <- resp.WithData(map[string]any{}).WithToken(token)
}

func Fail(ctx *gin.Context, message string) {
	resp := &Response{Code: enum.ErrorCode, Msg: enum.Msg(message)}
	ctx.JSON(http.StatusOK, resp.WithData(map[string]any{}))
}

func FailWs(c chan *Response, message string) {
	resp := &Response{Code: enum.ErrorCode, Msg: enum.Msg(message)}
	c <- resp.WithData(map[string]any{})
}

func FailNotFound(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, Response{
		Code: enum.ErrorCode,
		Msg:  enum.DefaultFailMsg,
	})
}

func FailAuth(ctx *gin.Context, message string) {
	resp := &Response{Code: enum.AuthErrorCode, Msg: enum.Msg(message)}
	ctx.AbortWithStatusJSON(http.StatusOK, resp.WithData(map[string]any{}))
}

func FailAuthWs(c chan *Response, message string) {
	resp := &Response{Code: enum.AuthErrorCode, Msg: enum.Msg(message)}
	c <- resp.WithData(map[string]any{})
}
