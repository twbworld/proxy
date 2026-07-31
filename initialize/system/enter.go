package system

import "errors"

type systemRes struct{}

// 启动系统资源
func Start() *systemRes {
	var errs []error
	if err := tgStart(); err != nil {
		errs = append(errs, err)
	}
	if err := timerStart(); err != nil {
		errs = append(errs, err)
	}
	if err := errors.Join(errs...); err != nil {
		panic(err)
	}
	return &systemRes{}
}

// 关闭系统资源
func (*systemRes) Stop() {
	var errs []error
	if err := tgClear(); err != nil {
		errs = append(errs, err)
	}
	if err := timerStop(); err != nil {
		errs = append(errs, err)
	}
	if err := errors.Join(errs...); err != nil {
		panic(err)
	}
}
