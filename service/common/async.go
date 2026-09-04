package common

import (
	"runtime/debug"

	"github.com/twbworld/proxy/global"
)

type AsyncService interface {
	GoSafe(func())
}

type asyncService struct {
}

func NewAsyncService() AsyncService {
	return new(asyncService)
}

// GoSafe 安全地开启异步协程，内置 panic 恢复与全局日志记录
func (a *asyncService) GoSafe(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.Log.Errorf("异步协程执行发生崩溃[panic]: %v\n堆栈信息:\n%s", r, debug.Stack())
			}
		}()
		fn()
	}()
}
