package global

import (
	"os"
	"reflect"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)


func initConfig() {
	struType := reflect.TypeOf(Config.AppConfig)
	struValue := reflect.ValueOf(&Config.AppConfig).Elem()
	arrN := struType.NumField()
	for i := 0; i < arrN; i++ {
		if struVal := struValue.FieldByName(struType.Field(i).Name); struVal.CanSet() {
			struVal.SetString(struType.Field(i).Tag.Get("default")) //参数初始化
		}
	}
}


func initEnv(configPath string) {
	// 初始化 viper
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("json")
	if err := v.ReadInConfig(); err != nil {
		Log.Fatalln("读取配置失败[u9ij]: ", err)
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

	// 将配置赋值给全局变量
	if err := v.Unmarshal(&Config.Env); err != nil {
		Log.Fatalln(err)
	}

}
func initTrojanGoConfig(configPath string) {
	struType := reflect.TypeOf(Config.TrojanGoConfig.Mysql)
	struValue := reflect.ValueOf(&Config.TrojanGoConfig.Mysql).Elem()
	arrN := struType.NumField()
	for i := 0; i < arrN; i++ {
		if struVal := struValue.FieldByName(struType.Field(i).Name); struVal.CanSet() {
			struVal.SetString(struType.Field(i).Tag.Get("default")) //参数初始化
		}
	}


	// 初始化 viper
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("json")
	if err := v.ReadInConfig(); err != nil {
		Log.Warnln("读取配置文件失败, 将读取环境变量作为参数配置[uihsd]: ", err)
		for i := 0; i < arrN; i++ {
			if struVal := struValue.FieldByName(struType.Field(i).Name); struVal.CanSet() {
				//读取环境变量,作为mysql连接参数
				if v, ok := os.LookupEnv(struType.Field(i).Tag.Get("env")); ok && strings.TrimSpace(v) != "" {
					struVal.SetString(v)
				}
			}
		}
		return
	}

	// 监听配置文件
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		Log.Infoln("配置文件变化[djiads]: ", in.Name)
		// 重载配置
		if err := v.Unmarshal(&Config.TrojanGoConfig); err != nil {
			if err := v.Unmarshal(&Config.TrojanGoConfig); err != nil {
				Log.Warnln(err)
			}
		}
	})

	// 将配置赋值给全局变量
	if err := v.Unmarshal(&Config.TrojanGoConfig); err != nil {
		Log.Fatalln(err)
	}

}
