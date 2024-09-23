package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model/db"
)

type TgService struct{}

type tgConfig struct {
	update *tg.Update
}

var lock sync.RWMutex

// 向Tg发送信息(请用协程执行)
func (t *TgService) TgSend(text string) (err error) {
	if global.Bot == nil {
		return fmt.Errorf("[ertioj98]出错")
	}

	if len(text) < 1 {
		return fmt.Errorf("[sioejn89]出错")
	}

	lock.RLock()
	defer lock.RUnlock()

	var str strings.Builder
	str.WriteString(`[`)
	str.WriteString(global.Config.ProjectName)
	str.WriteString(`]`)
	str.WriteString(text)

	msg := tg.NewMessage(global.Config.Telegram.Id, str.String())
	// msg.ParseMode = "MarkdownV2" //使用Markdown格式, 需要对特殊字符进行转义

	_, err = global.Bot.Send(msg)
	return
}

func (t *TgService) Webhook(ctx *gin.Context) (err error) {
	tc := &tgConfig{
		update: &tg.Update{},
	}

	if err = json.NewDecoder(ctx.Request.Body).Decode(tc.update); err != nil {
		return
	}

	return tc.handle()
}

func (c *tgConfig) handle() error {

	// st, _ := json.Marshal(update)
	// fmt.Println(string(st))

	if c.update.Message != nil {
		// (对话第一步) 和 (对话第五步)
		return c.firstStep()
	} else if c.update.CallbackQuery != nil {
		data := c.update.CallbackQuery.Data
		if data == "" {
			return errors.New("参数错误[gdoigjiod]")
		}

		//操作@用户id
		if params := regexp.MustCompile(`(.+)@(\d+)@(.+)`).FindStringSubmatch(data); len(params) > 3 {
			// (对话第四步)
			params = params[:4] //消除边界检查
			intNum, err := strconv.Atoi(params[2])
			if err != nil {
				return errors.New("参数错误[oidfjgoid]")
			}
			return c.input(uint(intNum), params[3])
		} else if params := regexp.MustCompile(`(.+)@(\d+)`).FindStringSubmatch(data); len(params) > 2 {
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

func (c *tgConfig) firstStep() error {
	msg := tg.NewMessage(c.update.Message.Chat.ID, "")
	//ReplyKeyboard位于输入框下的按钮
	msg.Text, msg.ReplyMarkup = "命令不存在!!!", tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("/start"),
		),
	)

	var info db.SystemInfo
	errSys := dao.App.SystemInfoDb.GetSysValByKey(&info, strconv.FormatInt(c.update.Message.Chat.ID, 10)+"_step")

	if c.update.Message.IsCommand() {
		switch c.update.Message.Command() {
		case "start", "help":
			msg.Text = "\nHello " + c.update.Message.From.FirstName + c.update.Message.From.LastName
			//InlineKeyboard位于对话框下的按钮
			msg.ReplyMarkup = tg.NewInlineKeyboardMarkup(
				tg.NewInlineKeyboardRow(
					tg.NewInlineKeyboardButtonData("查询用户", "user_select"),
					tg.NewInlineKeyboardButtonData("新增用户", "user_insert"),
				),
			)

			if errSys == nil || info.Value != "" {
				//初始化,清空流程码
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

		if errSys != nil {
			goto SEND
		} else if t, err := time.ParseInLocation(time.DateTime, info.UpdateTime, global.Tz); err != nil || time.Now().Unix()-t.Unix() > dao.SysTimeOut {
			if err != nil {
				return errors.New("系统错误[adfaio]" + err.Error())
			}

			//数据过期
			(&TgService{}).TgSend(`输入超时, 请按"/start"重新设置:`)
			return nil
		}

		if info.Value == "user_insert" {
			return c.userInsert(&info)
		}

		params := regexp.MustCompile(`(.+)@(\d+)@(.+)`).FindStringSubmatch(info.Value)
		if len(params) < 4 {
			return errors.New("参数错误[fijsa]")
		}

		params = params[:4]

		intNum, err := strconv.Atoi(params[2])
		if err != nil {
			return errors.New("错误[jklsd]: " + err.Error())
		}
		var user db.Users
		if err = dao.App.UsersDb.GetUsersByUserId(&user, uint(intNum)); err != nil {
			if err == sql.ErrNoRows {
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
					//为了更新时间字段update_time
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
					//为了更新时间字段update_time
					return dao.App.SystemInfoDb.SaveSysVal(info.Key, info.Value, tx)
				})
				if err != nil {
					return err
				}
				goto SEND
			}
			ExpiryDate := t.Format(time.DateOnly)
			user.ExpiryDate = &ExpiryDate
		default:
			return errors.New("错误[kdfhf]: " + params[3])
		}

		err = dao.Tx(func(tx *sqlx.Tx) (e error) {
			if e = dao.App.UsersDb.UpdateUsers(&user, tx); e != nil {
				return errors.New("错误[jkljkjkl]: " + e.Error())
			}
			//修改成功后,清空流程码
			return dao.App.SystemInfoDb.SaveSysVal(info.Key, "", tx)
		})
		if err != nil {
			return err
		}

		msg.Text = "*修改成功\\!\\!\\!*\n" + c.getUserMarkdownV2Text(&user)
		msg.ReplyMarkup = tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("修改此用户["+user.Username+"]", fmt.Sprintf("user_update@%d", user.Id)),
			),
		)
		msg.ParseMode = "MarkdownV2"
	}
SEND:
	go func() {
		if _, err := global.Bot.Send(msg); err != nil {
			global.Log.Error(err)
		}
	}()
	return nil
}

func (c *tgConfig) selectUser() error {
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
				//每行两个进行展示
				mkrow = append(mkrow, tg.NewInlineKeyboardRow(row...))
				row = make([]tg.InlineKeyboardButton, 0, 2)
			}
		}
		if len(row) > 0 {
			mkrow = append(mkrow, tg.NewInlineKeyboardRow(row...))
		}

		msg := tg.NewEditMessageTextAndMarkup(c.update.CallbackQuery.Message.Chat.ID, c.update.CallbackQuery.Message.MessageID, "选择*查询*的用户", tg.NewInlineKeyboardMarkup(mkrow...))
		msg.ParseMode = "MarkdownV2"

		go func() {
			if _, err := global.Bot.Send(msg); err != nil {
				global.Log.Error("错误[iofgjiosj]" + err.Error())
			}
		}()

	case "user_insert":
		err = dao.Tx(func(tx *sqlx.Tx) (e error) {
			return dao.App.SystemInfoDb.SaveSysVal(strconv.FormatInt(c.update.CallbackQuery.Message.Chat.ID, 10)+"_step", "user_insert", tx)
		})
		if err != nil {
			return err
		}

		msg := tg.NewMessage(c.update.CallbackQuery.Message.Chat.ID, "请输入用户名称, 4\\-64个字符以内的英文/数字/符号\n例:`210606_abc`")
		msg.ParseMode = "MarkdownV2"

		go func() {
			if _, err := global.Bot.Send(msg); err != nil {
				global.Log.Error("错误[iofgjfiosj]" + err.Error())
			}
		}()

	default:
		msg := tg.NewMessage(c.update.CallbackQuery.Message.Chat.ID, "命令不存在!!!")
		go func() {
			if _, err := global.Bot.Send(msg); err != nil {
				global.Log.Error("错误[iofgsjiosj]" + err.Error())
			}
		}()
	}

	return nil
}

func (c *tgConfig) actionType(act string, userId uint) error {
	var (
		user db.Users
		err  error
	)
	if err = dao.App.UsersDb.GetUsersByUserId(&user, userId); err != nil {
		(&TgService{}).TgSend("找不到用户")
		return errors.New("错误[jfdsgsd]" + err.Error())
	}

	msg := tg.NewEditMessageTextAndMarkup(c.update.CallbackQuery.Message.Chat.ID, c.update.CallbackQuery.Message.MessageID, c.getUserMarkdownV2Text(&user), tg.NewInlineKeyboardMarkup([]tg.InlineKeyboardButton{}))
	msg.ParseMode = "MarkdownV2"

	switch act {
	case "user_select":
		ikb := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("修改此用户["+user.Username+"]", fmt.Sprintf("user_update@%d", user.Id)),
			),
		)
		msg.ReplyMarkup = &ikb
		goto SEND3
	case "user_update":
		ikb := tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				// tg.NewInlineKeyboardButtonData("限流", fmt.Sprintf("user_update@%d@%s", user.Id, "quota")),
				tg.NewInlineKeyboardButtonData("到期", fmt.Sprintf("user_update@%d@%s", user.Id, "expiryDate")),
			),
		)
		msg.ReplyMarkup = &ikb
		msg.Text = fmt.Sprintf("选择修改\\[`%s`\\]的设置\n", user.Username) + msg.Text
		goto SEND3
	default:
		return (&TgService{}).TgSend("参数错误[goidjgd]")
	}

SEND3:
	go func() {
		if _, err := global.Bot.Send(msg); err != nil {
			global.Log.Error(err)
		}
	}()
	return nil
}

func (c *tgConfig) input(userId uint, value string) error {
	var user db.Users
	if dao.App.UsersDb.GetUsersByUserId(&user, userId) != nil {
		(&TgService{}).TgSend("找不到用户")
		return errors.New("找不到用户[kysafd]")
	}

	msg := tg.NewEditMessageTextAndMarkup(c.update.CallbackQuery.Message.Chat.ID, c.update.CallbackQuery.Message.MessageID, "", tg.NewInlineKeyboardMarkup([]tg.InlineKeyboardButton{}))
	msg.ParseMode = "MarkdownV2"

	switch value {
	case "quota":
		msg.Text = fmt.Sprintf("请输入限制\\[`%s`\\]的流量数值, 单位为*G*; 如想限制该用户为2\\.3G流量,则输入\"2\\.3\", 不限则输入\"\\-1\"\n", user.Username) + c.getUserMarkdownV2Text(&user)
	case "expiryDate":
		msg.Text = fmt.Sprintf("请输入限制\\[`%s`\\]的到期时间, 格式为\"2066\\-06\\-06\"\n", user.Username) + c.getUserMarkdownV2Text(&user)
	default:
		msg.Text = "命令不存在!!!"
	}

	go func() {
		if _, err := global.Bot.Send(msg); err != nil {
			global.Log.Error("错误[iodiosj]" + err.Error())
		}
	}()

	return dao.Tx(func(tx *sqlx.Tx) (e error) {
		return dao.App.SystemInfoDb.SaveSysVal(strconv.FormatInt(c.update.CallbackQuery.Message.Chat.ID, 10)+"_step", c.update.CallbackQuery.Data, tx)
	})
}

func (c *tgConfig) userInsert(info *db.SystemInfo) (err error) {
	msg := tg.NewMessage(c.update.Message.Chat.ID, "")
	msg.ReplyToMessageID = c.update.Message.MessageID //引用对话

	var user db.Users
	if ok, _ := regexp.MatchString("^[a-zA-Z0-9_-]{4,64}$", c.update.Message.Text); !ok {
		msg.Text = "请输入用户名称, 64个字符以内的英文/数字/字符\n例:`210606_abc`"
		msg.ParseMode = "MarkdownV2"

		go dao.Tx(func(tx *sqlx.Tx) (e error) {
			return dao.App.SystemInfoDb.SaveSysVal(info.Key, info.Value, tx)
		})

		goto SEND2
	}

	if dao.App.UsersDb.GetUsersByUserName(&user, c.update.Message.Text) == nil {
		msg.Text = strings.Replace(fmt.Sprintf("*用户`%s`已存在\\!\\!\\!*\n", c.update.Message.Text), `-`, `\-`, -1) + c.getUserMarkdownV2Text(&user)
		msg.ParseMode = "MarkdownV2"

		go dao.Tx(func(tx *sqlx.Tx) (e error) {
			return dao.App.SystemInfoDb.SaveSysVal(info.Key, info.Value, tx)
		})

		goto SEND2
	}

	err = dao.Tx(func(tx *sqlx.Tx) (e error) {
		if e = dao.App.UsersDb.InsertEmptyUsers(c.update.Message.Text, tx); e != nil {
			msg.Text = `系统错误, 请按"/start"重新设置4`
			go dao.App.SystemInfoDb.SaveSysVal(info.Key, info.Value, tx)
			return e
		}
		if e = dao.App.UsersDb.GetUsersByUserName(&user, c.update.Message.Text, tx); e != nil {
			msg.Text = `系统错误, 请按"/start"重新设置5`
			go dao.App.SystemInfoDb.SaveSysVal(info.Key, info.Value, tx)
			return e
		}
		return e
	})
	if err != nil {
		goto SEND2
	}

	msg.Text = "*新增成功\\!\\!\\!*\n" + c.getUserMarkdownV2Text(&user)
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("修改此用户["+user.Username+"]", fmt.Sprintf("user_update@%d", user.Id)),
		),
	)

SEND2:
	go func() {
		if _, err := global.Bot.Send(msg); err != nil {
			global.Log.Error(err)
		}
	}()
	return nil
}

func (c *tgConfig) getUserMarkdownV2Text(user *db.Users) string {
	// quota := "不限"
	// if user.Quota != -1 {
	// 	quota = fmt.Sprintf("%.1f", float64(user.Quota)/dao.App.UsersDb.QuotaMax) + "G"
	// }
	// download := fmt.Sprintf("%.1f", float64(user.Download)/dao.App.UsersDb.QuotaMax)
	// upload := fmt.Sprintf("%.1f", float64(user.Upload)/dao.App.UsersDb.QuotaMax)
	// text := fmt.Sprintf("账号: `%s`\nid: %d\n限流: %s\n上行: %sG\n下行: %sG\n到期: %s", user.Username, user.Id, quota, upload, download, *user.ExpiryDate)
	text := fmt.Sprintf("账号: `%s`\nid: %d\n到期: %s", user.Username, user.Id, *user.ExpiryDate)
	return strings.Replace(strings.Replace(text, `-`, `\-`, -1), `.`, `\.`, -1)
}
