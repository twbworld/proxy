package global

import (
	"flag"
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/config"
)

type GlobalInit struct {
}

var (
	Conf string
	Act  string
)

func init() {
	flag.StringVar(&Conf, "c", "", "choose config file.")
	flag.StringVar(&Act, "a", "", `行为,默认为空,即启动服务; "clear": 清除上下行流量记录; "expiry": 处理过期用户`)
}

func New(configFile ...string) *GlobalInit {
	var config string
	if gin.Mode() != gin.TestMode {
		//避免 单元测试(go test)自动加参数, 导致flag报错
		flag.Parse() //解析cli命令参数
		if Conf != "" {
			config = Conf
		}
	}
	if config == "" && len(configFile) > 0 {
		config = configFile[0]
	}
	if config == "" {
		config = `config.yaml`
	}

	// 初始化 viper
	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		panic("读取配置失败[u9ij]: " + config + err.Error())
	}

	// 监听配置文件
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("配置文件变化[djiads]: ", e.Name)
		if err := v.Unmarshal(global.Config); err != nil {
			if err := v.Unmarshal(global.Config); err != nil {
				fmt.Println(err)
			}
		}
		handleConfig(global.Config)
	})
	// 将配置赋值给全局变量(结构体需要设置mapstructure的tag)
	if err := v.Unmarshal(global.Config); err != nil {
		panic("出错[dhfal]: " + err.Error())
	}

	handleConfig(global.Config)

	return &GlobalInit{}
}

func (g *GlobalInit) Start() {
	if err := g.initLog(); err != nil {
		panic(err)
	}
	if err := g.initTz(); err != nil {
		panic(err)
	}
}

func handleConfig(c *config.Config) {
	c.StaticDir = strings.TrimRight(c.StaticDir, "/")
}
