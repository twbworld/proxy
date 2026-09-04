package initialize

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/config"
	"golang.org/x/sync/errgroup"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/twbworld/proxy/task"
)

var (
	Conf string
	Act  string
)

func init() {
	flag.StringVar(&Conf, "c", "", "choose config file.")
	flag.StringVar(&Act, "a", "", `行为,默认为空,即启动服务; "clear": 清除上下行流量记录; "expiry": 处理过期用户`)
}

// New 创建一个新的初始化器，并加载配置文件
func New(taskManager *task.Manager) *Initializer {
	if gin.Mode() != gin.TestMode {
		//避免 单元测试(go test)自动加参数, 导致flag报错
		flag.Parse() //解析cli命令参数
	}

	configPath := cmp.Or(Conf, "config.yaml")
	// 将 Initializer 实例创建提前
	i := &Initializer{taskManager: taskManager}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		panic("读取配置失败[u9ij]: " + configPath + err.Error())
	}

	// 初始化载入并绑定到原子指针
	initConfig := &config.Config{}
	if err := v.Unmarshal(initConfig); err != nil {
		panic("出错[dhfal]: " + err.Error())
	}
	handleConfig(initConfig)
	global.InitConfig(initConfig)

	// --- 为热重载引入Debounce(防抖)机制 ---
	var (
		debounceTimer *time.Timer
		debounceMutex sync.Mutex
	)
	const debounceDuration = 500 * time.Millisecond // 增加防抖时间，确保文件写入完成

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		debounceMutex.Lock()
		defer debounceMutex.Unlock()

		// 如果已有计时器，则重置它
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		// 启动一个新的计时器，在持续时间内如果没有新事件，则执行重载
		debounceTimer = time.AfterFunc(debounceDuration, func() {
			fmt.Println("配置文件变化[djiads] (Debounced): ", e.Name)

			oldConfig := global.Config() // 无锁获取配置的快照

			// 反序列化到临时对象，避免直接污染全局变量
			tmpConfig := &config.Config{}
			if err := v.Unmarshal(tmpConfig); err != nil {
				fmt.Println("热加载配置文件反序列化失败:", err)
				return
			}

			handleConfig(tmpConfig)

			// 安全原子替换全局配置 (完全无锁操作)
			global.InitConfig(tmpConfig)

			// 热重载
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			i.HandleConfigChange(ctx, oldConfig, tmpConfig)
		})
	})

	return i
}

// HandleConfigChange 配置热重载
func (i *Initializer) HandleConfigChange(ctx context.Context, oldConfig, newConfig *config.Config) {
	i.reloadLock.Lock()
	defer i.reloadLock.Unlock()

	var restartNeeded []string

	//检查不可热重载的高风险配置
	if !reflect.DeepEqual(oldConfig.Database, newConfig.Database) {
		restartNeeded = append(restartNeeded, "database")
	}
	if oldConfig.GinAddr != newConfig.GinAddr {
		restartNeeded = append(restartNeeded, "gin_addr")
	}
	if oldConfig.GinLogPath != newConfig.GinLogPath || oldConfig.RunLogPath != newConfig.RunLogPath {
		restartNeeded = append(restartNeeded, "log_path")
	}

	//并发执行可安全热重载的任务
	eg, egCtx := errgroup.WithContext(ctx)

	// 时区重载
	if oldConfig.Tz != newConfig.Tz {
		eg.Go(func() error {
			if err := i.InitTz(); err != nil {
				global.Log.Errorf("热重载时区失败: %v", err)
				return err
			}
			global.Log.Info("热重载时区成功")
			return nil
		})
	}

	if !reflect.DeepEqual(oldConfig.Redis, newConfig.Redis) {
		eg.Go(func() error {
			_ = i.redisClose()
			if err := i.initRedis(egCtx); err != nil {
				global.Log.Errorf("热重载Redis客户端失败: %v", err)
				return err
			}
			global.Log.Info("热重载Redis客户端成功")
			return nil
		})
	}

	if !reflect.DeepEqual(oldConfig.Database, newConfig.Database) {
		eg.Go(func() error {
			_ = i.dbClose()
			if err := i.initDB(); err != nil {
				global.Log.Errorf("热重载数据库客户端失败: %v", err)
				return err
			}
			global.Log.Info("热重载数据库客户端成功")
			return nil
		})
	}

	if !reflect.DeepEqual(oldConfig.Telegram, newConfig.Telegram) {
		eg.Go(func() error {
			_ = i.notifyClose()
			if err := i.initNotify(); err != nil {
				global.Log.Errorf("热重载消息客户端失败: %v", err)
				return err
			}
			global.Log.Info("热重载消息客户端成功")
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		global.Log.Errorf("并发热重载过程中发生错误: %v", err)
	}

	if len(restartNeeded) > 0 {
		global.Log.Warnf("检测到存在需要 重启服务 才能生效的配置变更: [%s]。", strings.Join(restartNeeded, ", "))
	}

	global.Log.Info("配置变更处理完成")
}

// handleConfig 处理和设置配置的默认值
func handleConfig(c *config.Config) {
	c.StaticDir = strings.TrimRight(cmp.Or(c.StaticDir, "static"), "/")
	c.ProjectName = cmp.Or(c.ProjectName, "my-project")
	c.GinAddr = cmp.Or(c.GinAddr, ":80")
	c.GinLogPath = cmp.Or(c.GinLogPath, "log/gin.log")
	c.RunLogPath = cmp.Or(c.RunLogPath, "log/run.log")
	c.Tz = cmp.Or(c.Tz, "Asia/Shanghai")
	c.Database.Type = cmp.Or(c.Database.Type, "sqlite")
	c.Database.SqlitePath = cmp.Or(c.Database.SqlitePath, "data.db")
	c.ClashPath = cmp.Or(c.ClashPath, "config.yaml")

	if c.LogRetentionDays == 0 {
		c.LogRetentionDays = 7
	}
	if len(c.Cors) == 0 {
		c.Cors = []string{"*"}
	}
	if c.Redis.LockExpiry == 0 {
		c.Redis.LockExpiry = 30
	}
	if c.Redis.ConversationHistoryTTL == 0 {
		c.Redis.ConversationHistoryTTL = 3600
	}
	if c.Redis.HistoryLockExpiry == 0 {
		c.Redis.HistoryLockExpiry = 10
	}
}
