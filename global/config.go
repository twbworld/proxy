package global

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)


func initConfig(configPath string) {
	// 初始化 viper
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("json")
	if err := v.ReadInConfig(); err != nil {
		log.Fatalln("读取配置失败[dopfido]: ", err)
	}

	// 监听配置文件
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		log.Println("配置文件变化[foiod]: ", in.Name)
		// 重载配置
		if err := v.Unmarshal(&Config.AppConfig); err != nil {
			if err := v.Unmarshal(&Config.AppConfig); err != nil {
				log.Println(err)
			}
		}
	})
	// 将配置赋值给全局变量
	if err := v.Unmarshal(&Config.AppConfig); err != nil {
		log.Fatalln(err)
	}
}

func initEnv(configPath string) {
	// 初始化 viper
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("json")
	if err := v.ReadInConfig(); err != nil {
		log.Fatalln("读取配置失败[u9ij]: ", err)
	}

	// 监听配置文件
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		log.Println("配置文件变化[djiads]: ", in.Name)
		// 重载配置
		if err := v.Unmarshal(&Config.Env); err != nil {
			if err := v.Unmarshal(&Config.Env); err != nil {
				log.Println(err)
			}
		}
	})

	// 将配置赋值给全局变量
	if err := v.Unmarshal(&Config.Env); err != nil {
		log.Fatalln(err)
	}

}
