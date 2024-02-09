package global

import (
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func initConfig() {
	struType, struValue := reflect.TypeOf(Config.AppConfig), reflect.ValueOf(&Config.AppConfig).Elem()
	for i := range struType.NumField() {
		if struVal := struValue.FieldByName(struType.Field(i).Name); struVal.CanSet() {
			struVal.SetString(struType.Field(i).Tag.Get("default")) //参数初始化
		}
	}
}

func initEnv(configPath string) {
	struType, struValue := reflect.TypeOf(Config.Env.Db), reflect.ValueOf(&Config.Env.Db).Elem()
	for i := range struType.NumField() {
		if struVal := struValue.FieldByName(struType.Field(i).Name); struVal.CanSet() {
			struVal.SetString(struType.Field(i).Tag.Get("default")) //参数初始化
		}
	}

	// 初始化 viper
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("json")
	if err := v.ReadInConfig(); err != nil {
		panic("读取配置失败[u9ij]: " + err.Error())
	}

	// 监听配置文件
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		Log.Infoln("配置文件变化[djiads]: ", in.Name)
		// 重载配置
		if err := v.Unmarshal(&Config.Env); err != nil {
			if err := v.Unmarshal(&Config.Env); err != nil {
				Log.Warnln(err)
			}
		}
	})

	// 将配置赋值给全局变量(结构体需要设置mapstructure的tag)
	if err := v.Unmarshal(&Config.Env); err != nil {
		panic("出错[dhfal]: " + err.Error())
	}

}
