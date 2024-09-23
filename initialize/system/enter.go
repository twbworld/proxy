package system

type systemRes struct{}

// 启动系统资源
func Start() *systemRes {
	if err := tgStart(); err != nil {
		panic(err)
	}
	if err := timerStart(); err != nil {
		panic(err)
	}
	return &systemRes{}
}

// 关闭系统资源
func (*systemRes) Stop() {
	if err := tgClear(); err != nil {
		panic(err)
	}
	if err := timerStop(); err != nil {
		panic(err)
	}
}
