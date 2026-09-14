package admin

import (
	"context"
	"database/sql"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/db"
	"github.com/twbworld/proxy/service/common"
)

var (
	step3Regex    = regexp.MustCompile(`(.+)@(\d+)`)
	step4Regex    = regexp.MustCompile(`(.+)@(\d+)@(.+)`)
	userNameRegex = regexp.MustCompile("^[a-zA-Z0-9_-]{4,64}$")
)

type BotService interface {
	Webhook(ctx *gin.Context) error
}

type botService struct {
	async    common.AsyncService
	notifier global.NotifierService
	llmSvc   LLMService
}

type handler struct {
	*botService
	update   *tg.Update
}

func NewBotService(async common.AsyncService, notifier global.NotifierService, llmSvc LLMService) BotService {
	return &botService{
		async:    async,
		notifier: notifier,
		llmSvc:   llmSvc,
	}
}

func (t *botService) Webhook(ctx *gin.Context) error {
	update := new(tg.Update)
	if err := json.UnmarshalRead(ctx.Request.Body, update); err != nil {
		return err
	}

	//为避免高并发的数据覆盖, 每个请求都构建新的handler
	handler := &handler{
		botService: t,
		update:   update,
	}
	return handler.handle()
}

func (c *handler) handle() error {
	if c.update.Message != nil {
		// (对话第一步) 和 (对话第五步)
		return c.firstStep()
	} else if c.update.CallbackQuery != nil {
		data := c.update.CallbackQuery.Data
		if data == "" {
			return errors.New("参数错误[gdoigjiod]")
		}

		//操作@用户id
		if params := step4Regex.FindStringSubmatch(data); len(params) > 3 {
			// (对话第四步)
			params = params[:4] //消除边界检查
			intNum, err := strconv.Atoi(params[2])
			if err != nil {
				return errors.New("参数错误[oidfjgoid]")
			}
			return c.input(uint(intNum), params[3])
		} else if params := step3Regex.FindStringSubmatch(data); len(params) > 2 {
			params = params[:3]
			// (对话第三步)
			intNum, err := strconv.Atoi(params[2])
			if err != nil {
				return errors.New("参数错误[opdpp]")
			}
			return c.actionType(params[1], uint(intNum))

		} else {
			// (对话第二步)
			return c.selectUser()
		}
	}
	return nil
}

func (c *handler) firstStep() error {
	msg := tg.NewMessage(c.update.Message.Chat.ID, "")
	// ReplyKeyboard位于输入框下的按钮
	msg.Text, msg.ReplyMarkup = "命令不存在!!!", tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("/start"),
		),
	)

	info := new(db.SystemInfo)
	errSys := dao.App.GetByCol(info, "key", strconv.FormatInt(c.update.Message.Chat.ID, 10)+"_step")
	if c.update.Message.IsCommand() {
		switch c.update.Message.Command() {
		case "start", "help":
			msg.Text = "\nHello " + c.update.Message.From.FirstName + c.update.Message.From.LastName
			// InlineKeyboard位于对话框下的按钮
			msg.ReplyMarkup = tg.NewInlineKeyboardMarkup(
				tg.NewInlineKeyboardRow(
					tg.NewInlineKeyboardButtonData("查询用户", "user_select"),
					tg.NewInlineKeyboardButtonData("新增用户", "user_insert"),
				),
			)

			if errSys == nil || info.Value != "" {
				// 初始化,清空流程码
				err := dao.Tx(func(tx *sqlx.Tx) (e error) {
					return dao.App.SystemInfoDb.SaveSysVal(strconv.FormatInt(c.update.Message.Chat.ID, 10)+"_step", "", tx)
				})
				if err != nil {
					return err
				}
			}
		default:
			msg.ReplyToMessageID = c.update.Message.MessageID //引用对话
		}
	} else {
		msg.ReplyToMessageID = c.update.Message.MessageID //引用对话

		// 检查用户是否处于状态机流程中
		if errSys != nil || info.Value == "" {
			// 用户不在任何状态机流程中，使用 LLM 智能助理处理自然语言
			c.async.GoSafe(func() {
				timeout := 120 * time.Second
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				reply, err := c.llmSvc.Handle(ctx, c.update.Message.Chat.ID, c.update.Message.Text)
				if err == nil {
					// 容错处理：小模型常忽略提示强制输出双星号，在此强制转为单星号兼容 TG 渲染
					reply = strings.ReplaceAll(reply, "**", "*")
				}

				msg := tg.NewMessage(c.update.Message.Chat.ID, reply)
				if err != nil {
					global.Log.Errorf("LLM 处理失败：%v", err)
					msg.Text = "智能助理处理失败，请稍后重试"
				}

				// 开启基础 Markdown 解析模式
				msg.ParseMode = "Markdown"

				if sendErr := c.notifier.SendRaw(msg); sendErr != nil {
					global.Log.Errorf("发送 LLM 回复到 TG 失败：%v", sendErr)
				}
			})
			return nil
		}

		// 用户在状态机流程中，检查是否超时
		if t := time.Now().Unix() - info.UpdatedAt; t < 0 || info.UpdatedAt == 0 || t > dao.SysTimeOut {
			if t < 0 || info.UpdatedAt == 0 {
				return fmt.Errorf("系统错误[adfaio]%d", info.UpdatedAt)
			}
			c.notifier.Send(`输入超时, 请按"/start"重新设置:`)
			return nil
		}

		if info.Value == "user_insert" {
			return c.userInsert(info)
		}

		params := step4Regex.FindStringSubmatch(info.Value)
		if len(params) < 4 {
			return errors.New("参数错误[fijsa]")
		}
		params = params[:4]

		intNum, err := strconv.Atoi(params[2])
		if err != nil {
			return errors.New("错误[jklsd]: " + err.Error())
		}
		user := new(db.Users)
		if err = dao.App.GetByCol(user, "id", uint(intNum)); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("用户不存在[tigfffhh]")
			}
			return errors.New("错误[tigfhh]: " + err.Error())
		}

		switch params[3] {
		case "quota":
			text, err := strconv.ParseFloat(c.update.Message.Text, 64)
			if err != nil || (text != -1 && text < 0) {
				msg.Text = fmt.Sprintf("请输入限制\\[`%s`\\]的流量数值, 单位为*G*; 如想限制该用户为2\\.3G流量,则输入\"2\\.3\", 不限则输入\"\\-1\"", user.Username)
				msg.ParseMode = "MarkdownV2"

				err := dao.Tx(func(tx *sqlx.Tx) (e error) {
					return dao.App.SystemInfoDb.SaveSysVal(info.Key, info.Value, tx)
				})
				if err != nil {
					return err
				}
				goto SEND
			}
			if c.update.Message.Text != "-1" {
				text = text * dao.QuotaMax
			}
			user.Quota = int(text)
		case "expiryDate":
			t, err := time.ParseInLocation(time.DateOnly, c.update.Message.Text, global.Tz)
			if err != nil {
				msg.Text = fmt.Sprintf("请输入限制\\[`%s`\\]的到期时间, 格式为\"2066\\-06\\-06\"", user.Username)
				msg.ParseMode = "MarkdownV2"
				err := dao.Tx(func(tx *sqlx.Tx) (e error) {
					return dao.App.SystemInfoDb.SaveSysVal(info.Key, info.Value, tx)
				})
				if err != nil {
					return err
				}
				goto SEND
			}
			user.ExpiryDate = t.In(time.UTC).Format(time.DateOnly)
		default:
			return errors.New("错误[kdfhf]: " + params[3])
		}

		err = dao.Tx(func(tx *sqlx.Tx) (e error) {
			if e = dao.App.UsersDb.UpdateUsers(user, tx); e != nil {
				return errors.New("错误[jkljkjkl]: " + e.Error())
			}
			// 修改成功后,清空流程码
			return dao.App.SystemInfoDb.SaveSysVal(info.Key, "", tx)
		})
		if err != nil {
			return err
		}

		msg.Text = "*修改成功\\!\\!\\!*\n" + c.getUserMarkdownV2Text(user)
		msg.ReplyMarkup = tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("修改此用户["+user.Username+"]", fmt.Sprintf("user_update@%d", user.Id)),
			),
		)
		msg.ParseMode = "MarkdownV2"
	}
SEND:
	c.async.GoSafe(func() {
		if err := c.notifier.SendRaw(msg); err != nil {
			global.Log.Error(err)
		}
	})
	return nil
}

func (c *handler) selectUser() error {
	var err error
	switch c.update.CallbackQuery.Data {
	case "user_select":
		var usersMysql []db.Users
		if err = dao.App.UsersDb.GetUserNames(&usersMysql); err != nil {
			return errors.New("错误[lkffgdh]" + err.Error())
		}
		row, mkrow := make([]tg.InlineKeyboardButton, 0, 2), make([][]tg.InlineKeyboardButton, 0, len(usersMysql)/2+1)
		for _, v := range usersMysql {
			row = append(row, tg.NewInlineKeyboardButtonData(v.Username, c.update.CallbackQuery.Data+"@"+strconv.Itoa(int(v.Id))))
			if len(row) == 2 {
				mkrow = append(mkrow, tg.NewInlineKeyboardRow(row...))
				row = make([]tg.InlineKeyboardButton, 0, 2)
			}
		}
		if len(row) > 0 {
			mkrow = append(mkrow, tg.NewInlineKeyboardRow(row...))
		}

		msg := tg.NewEditMessageTextAndMarkup(c.update.CallbackQuery.Message.Chat.ID, c.update.CallbackQuery.Message.MessageID, "选择*查询*的用户", tg.NewInlineKeyboardMarkup(mkrow...))
		msg.ParseMode = "MarkdownV2"

		c.async.GoSafe(func() {
			if err := c.notifier.SendRaw(msg); err != nil {
				global.Log.Error("错误[iofgjiosj]" + err.Error())
			}
		})

	case "user_insert":
		err = dao.Tx(func(tx *sqlx.Tx) (e error) {
			return dao.App.SystemInfoDb.SaveSysVal(strconv.FormatInt(c.update.CallbackQuery.Message.Chat.ID, 10)+"_step", "user_insert", tx)
		})
		if err != nil {
			return err
		}

		msg2 := tg.NewMessage(c.update.CallbackQuery.Message.Chat.ID, "请输入用户名称, 4\\-64个字符以内的英文/数字/符号\n例:`210606_abc`")
		msg2.ParseMode = "MarkdownV2"

		c.async.GoSafe(func() {
			if err := c.notifier.SendRaw(msg2); err != nil {
				global.Log.Error("错误[iosdgiosj]" + err.Error())
			}
		})

	default:
		msg3 := tg.NewMessage(c.update.CallbackQuery.Message.Chat.ID, "命令不存在!!!")
		c.async.GoSafe(func() {
			if err := c.notifier.SendRaw(msg3); err != nil {
				global.Log.Error("错误[idgfbiosj]" + err.Error())
			}
		})
	}

	return nil
}

func (c *handler) actionType(act string, userId uint) error {
	var err error
	user := new(db.Users)
	if err = dao.App.GetByCol(user, "id", userId); err != nil {
		// 调用通用接口通知
		c.notifier.Send("找不到用户")
		return errors.New("错误[jfdsgsd]" + err.Error())
	}

	msg := tg.NewEditMessageTextAndMarkup(c.update.CallbackQuery.Message.Chat.ID, c.update.CallbackQuery.Message.MessageID, c.getUserMarkdownV2Text(user), tg.NewInlineKeyboardMarkup([]tg.InlineKeyboardButton{}))
	msg.ParseMode = "MarkdownV2"

	switch act {
	case "user_select":
		msg.ReplyMarkup = new(tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("修改此用户["+user.Username+"]", fmt.Sprintf("user_update@%d", user.Id)),
			),
		))
		goto SEND3
	case "user_update":
		msg.ReplyMarkup = new(tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("到期", fmt.Sprintf("user_update@%d@%s", user.Id, "expiryDate")),
			),
		))
		msg.Text = fmt.Sprintf("选择修改\\[`%s`\\]的设置\n", user.Username) + msg.Text
		goto SEND3
	default:
		c.notifier.Send("参数错误[goidjgd]")
		return nil
	}

SEND3:
	c.async.GoSafe(func() {
		if err := c.notifier.SendRaw(msg); err != nil {
			global.Log.Error("错误[isefjiosj]" + err.Error())
		}
	})
	return nil
}

func (c *handler) input(userId uint, value string) error {
	user := new(db.Users)
	if dao.App.GetByCol(user, "id", userId) != nil {
		c.notifier.Send("找不到用户")
		return errors.New("找不到用户[kysafd]")
	}

	msg := tg.NewEditMessageTextAndMarkup(c.update.CallbackQuery.Message.Chat.ID, c.update.CallbackQuery.Message.MessageID, "", tg.NewInlineKeyboardMarkup([]tg.InlineKeyboardButton{}))
	msg.ParseMode = "MarkdownV2"

	switch value {
	case "quota":
		msg.Text = fmt.Sprintf("请输入限制\\[`%s`\\]的流量数值, 单位为*G*; 如想限制该用户为2\\.3G流量,则输入\"2\\.3\", 不限则输入\"\\-1\"\n", user.Username) + c.getUserMarkdownV2Text(user)
	case "expiryDate":
		msg.Text = fmt.Sprintf("请输入限制\\[`%s`\\]的到期时间, 格式为\"2066\\-06\\-06\"\n", user.Username) + c.getUserMarkdownV2Text(user)
	default:
		msg.Text = "命令不存在!!!"
	}

	c.async.GoSafe(func() {
		if err := c.notifier.SendRaw(msg); err != nil {
			global.Log.Error("错误[ser]" + err.Error())
		}
	})

	return dao.Tx(func(tx *sqlx.Tx) (e error) {
		return dao.App.SystemInfoDb.SaveSysVal(strconv.FormatInt(c.update.CallbackQuery.Message.Chat.ID, 10)+"_step", c.update.CallbackQuery.Data, tx)
	})
}

func (c *handler) userInsert(info *db.SystemInfo) (err error) {
	msg := tg.NewMessage(c.update.Message.Chat.ID, "")
	msg.ReplyToMessageID = c.update.Message.MessageID //引用对话

	user := new(db.Users)
	// 1. 用户名格式不合法：刷新会话时间，提示重试
	if !userNameRegex.MatchString(c.update.Message.Text) {
		msg.Text = "请输入用户名称, 64个字符以内的英文/数字/字符\n例:`210606_abc`"
		msg.ParseMode = "MarkdownV2"

		c.async.GoSafe(func() {
			_ = dao.Tx(func(tx *sqlx.Tx) (e error) {
				return dao.App.SystemInfoDb.SaveSysVal(info.Key, info.Value, tx)
			})
		})
		goto SEND2
	}

	// 2. 用户名已存在：刷新会话时间，展示已有用户信息
	if dao.App.GetByCol(user, "username", c.update.Message.Text) == nil {
		msg.Text = strings.Replace(fmt.Sprintf("*用户`%s`已存在\\!\\!\\!*\n", c.update.Message.Text), `-`, `\-`, -1) + c.getUserMarkdownV2Text(user)
		msg.ParseMode = "MarkdownV2"

		c.async.GoSafe(func() {
			_ = dao.Tx(func(tx *sqlx.Tx) (e error) {
				return dao.App.SystemInfoDb.SaveSysVal(info.Key, info.Value, tx)
			})
		})
		goto SEND2
	}

	// 3. 执行核心原子事务：创建用户 -> 查询新建用户 -> 清空流程状态码
	err = dao.Tx(func(tx *sqlx.Tx) (e error) {
		// 插入新用户
		if e = dao.App.UsersDb.InsertEmptyUsers(c.update.Message.Text, tx); e != nil {
			return fmt.Errorf("插入用户失败[iosdhja]: %w", e)
		}
		// 获取新建的用户信息
		if e = dao.App.GetByCol(user, "username", c.update.Message.Text, tx); e != nil {
			return fmt.Errorf("获取新建用户失败: %w", e)
		}
		// 新增成功后，在当前事务内直接清空步骤状态码
		if e = dao.App.SystemInfoDb.SaveSysVal(info.Key, "", tx); e != nil {
			return fmt.Errorf("清空流程状态失败: %w", e)
		}
		return nil
	})
	// 4. 事务失败时的容错处理
	if err != nil {
		global.Log.Errorf("创建用户事务失败[userInsert]: %v", err)
		msg.Text = `系统错误, 请按"/start"重新设置`

		// 事务已完全回滚，如果需要保留并刷新步骤状态，使用全新的独立事务处理
		c.async.GoSafe(func() {
			_ = dao.Tx(func(newTx *sqlx.Tx) error {
				return dao.App.SystemInfoDb.SaveSysVal(info.Key, info.Value, newTx)
			})
		})
		goto SEND2
	}

	// 5. 成功后构建响应内容与操作菜单
	msg.Text = "*新增成功\\!\\!\\!*\n" + c.getUserMarkdownV2Text(user) + generateLinksText(user.Username, true)
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("修改此用户["+user.Username+"]", fmt.Sprintf("user_update@%d", user.Id)),
		),
	)

SEND2:
	c.async.GoSafe(func() {
		if err := c.notifier.SendRaw(msg); err != nil {
			global.Log.Error("错误[iofgjijghj]" + err.Error())
		}
	})
	return nil
}

func (c *handler) getUserMarkdownV2Text(user *db.Users) string {
	expiryDateStr := user.ExpiryDate
	if t, err := time.Parse(time.DateOnly, user.ExpiryDate); err == nil {
		expiryDateStr = t.In(global.Tz).Format(time.DateOnly)
	}
	text := fmt.Sprintf("账号: `%s`\nid: %d\n到期: %s", user.Username, user.Id, expiryDateStr)
	return strings.Replace(strings.Replace(text, `-`, `\-`, -1), `.`, `\.`, -1)
}

// generateLinksText 统筹生成两条订阅链接并可选是否对 MarkdownV2 特殊符号转义
func generateLinksText(username string, escape bool) string {
	cfg := global.Config()
	if cfg == nil || cfg.Domain == "" {
		return ""
	}
	domain := strings.TrimRight(cfg.Domain, "/")

	v2rayLink := fmt.Sprintf("%s/%s.html", domain, username)
	clashLink := v2rayLink

	if u, err := url.Parse(domain); err == nil && u.Host != "" {
		hostParts := strings.Split(u.Host, ".")
		if len(hostParts) >= 2 {
			if hostParts[0] == "www" {
				hostParts[0] = "clash"
			} else if hostParts[0] != "clash" {
				hostParts = append([]string{"clash"}, hostParts...)
			}
			u.Host = strings.Join(hostParts, ".")
			clashLink = fmt.Sprintf("%s/%s.html", u.String(), username)
		}
	}

	text := fmt.Sprintf("\n• 订阅链接:\n  1. Clash: %s\n  2. V2ray: %s", clashLink, v2rayLink)
	if escape {
		replacer := strings.NewReplacer(
			`.`, `\.`, `-`, `\-`, `_`, `\_`, `=`, `\=`,
		)
		text = replacer.Replace(text)
	}
	return text
}
