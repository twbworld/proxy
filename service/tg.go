package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/twbworld/proxy/dao"
	"github.com/twbworld/proxy/global"
	"github.com/twbworld/proxy/model"
)

var IsTgSend bool //只允许发送一次

func TgSend(text string) (err error) {
	if global.Bot == nil || text == "" || IsTgSend {
		return
	}
	msg := tg.NewMessage(global.Config.Env.Telegram.Id, fmt.Sprintf("[%s]%s", global.Config.AppConfig.ProjectName, text))
	_, err = global.Bot.Send(msg)
	if err == nil {
		IsTgSend = true
	}
	return
}

func TgWebhookClear() (err error) {
	if global.Bot == nil {
		return
	}
	_, err = global.Bot.Request(tg.DeleteWebhookConfig{})
	return
}

func TgWebhookHandle(ctx *gin.Context) error {
	var update tg.Update
	if err := json.NewDecoder(ctx.Request.Body).Decode(&update); err != nil {
		return err
	}

	// st, _ := json.Marshal(update)
	// fmt.Println(string(st))

	if update.Message != nil {
		// (对话第一步) 和 (对话第五步)
		return firstStep(&update)
	} else if update.CallbackQuery != nil {
		if update.CallbackQuery.Data == "" {
			return errors.New("参数错误[gdoigjiod]")
		}

		//操作@用户id
		if params := regexp.MustCompile(`(.+)@(\d+)@(.+)`).FindStringSubmatch(update.CallbackQuery.Data); len(params) > 3 {
			// (对话第四步)
			intNum, err := strconv.Atoi(params[2])
			if err != nil {
				return errors.New("参数错误[oidfjgoid]")
			}
			return input(&update, uint(intNum), params[3])
		} else if params := regexp.MustCompile(`(.+)@(\d+)`).FindStringSubmatch(update.CallbackQuery.Data); len(params) > 2 {
			// (对话第三步)
			intNum, err := strconv.Atoi(params[2])
			if err != nil {
				return errors.New("参数错误[opdpp]")
			}
			return actionType(&update, params[1], uint(intNum))

		} else {
			// (对话第二步)
			return selectUser(&update)
		}

	}
	return nil
}

func firstStep(update *tg.Update) error {
	msg := tg.NewMessage(update.Message.Chat.ID, "")
	msg.Text = "命令不存在!!!"
	//ReplyKeyboard位于输入框下的按钮
	msg.ReplyMarkup = tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("/start"),
		),
	)

	var info model.SystemInfo
	errSys := dao.GetSysValByKey(&info, strconv.FormatInt(update.Message.Chat.ID, 10)+"_step")

	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "start", "help":
			msg.Text = "\nHello " + update.Message.From.FirstName + update.Message.From.LastName
			//InlineKeyboard位于对话框下的按钮
			msg.ReplyMarkup = tg.NewInlineKeyboardMarkup(
				tg.NewInlineKeyboardRow(
					tg.NewInlineKeyboardButtonData("查询用户", "user_select"),
					tg.NewInlineKeyboardButtonData("新增用户", "user_insert"),
				),
			)

			if errSys == nil || info.Value != "" {
				//初始化,清空流程码
				tx, err := dao.DB.Beginx()
				if err != nil {
					return errors.New("开启事务失败[ghfgasd]: " + err.Error())
				}
				defer func() {
					if p := recover(); p != nil {
						tx.Rollback()
						panic(p)
					} else if err != nil {
						tx.Rollback()
					} else {
						err = tx.Commit()
						if err != nil {
							tx.Rollback()
							panic(err)
						}
					}
				}()

				err = dao.SaveSysVal(tx, strconv.FormatInt(update.Message.Chat.ID, 10)+"_step", "")
				if err != nil {
					err = errors.New("开启事务失败[oierutgoe]: " + err.Error())
				}
			}

		default:
			msg.ReplyToMessageID = update.Message.MessageID //引用对话
		}
	} else {
		msg.ReplyToMessageID = update.Message.MessageID //引用对话

		if errSys != nil {
			goto SEND
		} else if t, err := time.Parse(time.DateTime, info.UpdateTime); err != nil || time.Now().Unix()-t.Unix() > dao.SysTimeOut {
			if err != nil {
				return errors.New("系统错误[adfaio]" + err.Error())
			}
			//数据过期
			TgSend(`输入超时, 请按"/start"重新设置:`)
			return nil
		}

		if info.Value == "user_insert" {
			return userInsert(update, &info)
		}

		params := regexp.MustCompile(`(.+)@(\d+)@(.+)`).FindStringSubmatch(info.Value)
		if len(params) < 4 {
			return errors.New("参数错误[fijsa]")
		}

		intNum, err := strconv.Atoi(params[2])
		if err != nil {
			return errors.New("错误[jklsd]: " + err.Error())
		}
		var user model.Users
		err = dao.GetUsersByUserId(&user, uint(intNum))
		if err != nil {
			return errors.New("错误[tigfhh]: " + err.Error())
		}

		tx, err := dao.DB.Beginx()
		if err != nil {
			return errors.New("开启事务失败[lkdfty]: " + err.Error())
		}
		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				panic(p)
			} else if err != nil {
				global.Log.Warnln("事务回滚[lghjfn]: ", err)
				tx.Rollback()
			} else {
				err = tx.Commit()
				panic(err)
			}
		}()

		switch params[3] {
		case "quota":
			text, err := strconv.ParseFloat(update.Message.Text, 64)
			if err != nil || (text != -1 && text < 0) {
				msg.Text = fmt.Sprintf("请输入限制\\[`%s`\\]的流量数值, 单位为*G*; 如想限制该用户为2\\.3G流量,则输入\"2\\.3\", 不限则输入\"\\-1\"", user.Username)
				msg.ParseMode = "MarkdownV2"
				dao.SaveSysVal(tx, info.Key, info.Value) //为了更新时间字段update_time
				goto SEND
			}
			if update.Message.Text != "-1" {
				text = text * dao.QuotaMax
			}
			user.Quota = int(text)
		case "expiryDate":
			t, err := time.Parse(time.DateOnly, update.Message.Text)
			if err != nil {
				msg.Text = fmt.Sprintf("请输入限制\\[`%s`\\]的到期时间, 格式为\"2066\\-06\\-06\"", user.Username)
				msg.ParseMode = "MarkdownV2"
				dao.SaveSysVal(tx, info.Key, info.Value)
				goto SEND
			}
			ExpiryDate := t.Format(time.DateOnly)
			user.ExpiryDate = &ExpiryDate
		default:
			return errors.New("错误[kdfhf]: " + params[3])
		}

		err = dao.UpdateUsers(tx, &user)
		if err != nil {
			return errors.New("错误[jkljkjkl]: " + err.Error())
		}

		//修改成功后,清空流程码
		err = dao.SaveSysVal(tx, info.Key, "")
		if err != nil {
			return errors.New("错误[asdsad]: " + err.Error())
		}
		msg.Text = "*修改成功\\!\\!\\!*\n" + getUserMarkdownV2Text(&user)
		msg.ReplyMarkup = tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("修改此用户["+user.Username+"]", fmt.Sprintf("user_update@%d", user.Id)),
			),
		)
		msg.ParseMode = "MarkdownV2"
	}
SEND:
	_, err := global.Bot.Send(msg)
	return err
}

func selectUser(update *tg.Update) error {
	switch update.CallbackQuery.Data {
	case "user_select":
		var (
			row        []tg.InlineKeyboardButton
			mkrow      [][]tg.InlineKeyboardButton
			usersMysql []model.Users
		)
		err := dao.GetUserNames(&usersMysql)
		if err != nil {
			return errors.New("错误[lkffgdh]" + err.Error())
		}
		l := len(usersMysql)
		for _, v := range usersMysql {
			l--
			row = append(row, tg.NewInlineKeyboardButtonData(v.Username, update.CallbackQuery.Data+"@"+strconv.Itoa(int(v.Id))))
			if len(row) == 2 || l == 0 {
				//每行两个进行展示
				mkrow = append(mkrow, tg.NewInlineKeyboardRow(row...))
				row = []tg.InlineKeyboardButton{}
			}
		}
		msg := tg.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "选择*查询*的用户", tg.NewInlineKeyboardMarkup(mkrow...))
		msg.ParseMode = "MarkdownV2"
		if _, err = global.Bot.Send(msg); err != nil {
			return errors.New("错误[iofgjiosj]" + err.Error())
		}
	case "user_insert":
		tx, err := dao.DB.Beginx()
		if err != nil {
			return errors.New("开启事务失败[sdfgdfjkgh]" + err.Error())
		}
		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				panic(p)
			} else if err != nil {
				tx.Rollback()
			} else {
				err = tx.Commit()
				if err != nil {
					panic(err)
				}
			}
		}()

		err = dao.SaveSysVal(tx, strconv.FormatInt(update.CallbackQuery.Message.Chat.ID, 10)+"_step", "user_insert")
		if err != nil {
			return errors.New("错误[fgrejlk]" + err.Error())
		}

		msg := tg.NewMessage(update.CallbackQuery.Message.Chat.ID, "请输入用户名称, 4\\-64个字符以内的英文/数字/符号\n例:`210606_abc`")
		msg.ParseMode = "MarkdownV2"
		if _, err := global.Bot.Send(msg); err != nil {
			return errors.New("错误[fgrejlk]" + err.Error())
		}

	default:
		msg := tg.NewMessage(update.CallbackQuery.Message.Chat.ID, "命令不存在!!!")
		if _, err := global.Bot.Send(msg); err != nil {
			return errors.New("错误[fdxcvjkh]" + err.Error())
		}
	}

	return nil
}

func actionType(update *tg.Update, act string, userId uint) error {
	var user model.Users
	err := dao.GetUsersByUserId(&user, userId)
	if err != nil {
		TgSend("找不到用户")
		return errors.New("错误[jfdsgsd]" + err.Error())
	}
	msg := tg.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, getUserMarkdownV2Text(&user), tg.NewInlineKeyboardMarkup([]tg.InlineKeyboardButton{}))
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
		return TgSend("参数错误[goidjgd]")
	}

SEND3:
	_, err = global.Bot.Send(msg)
	return err
}

func input(update *tg.Update, userId uint, value string) error {
	var user model.Users
	err := dao.GetUsersByUserId(&user, userId)
	if err != nil {
		TgSend("找不到用户")
		return errors.New("找不到用户[kysafd]")
	}

	ikb := tg.NewInlineKeyboardMarkup(
		[]tg.InlineKeyboardButton{},
	)
	msg := tg.NewEditMessageTextAndMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "", ikb)

	msg.ParseMode = "MarkdownV2"
	switch value {
	case "quota":
		msg.Text = fmt.Sprintf("请输入限制\\[`%s`\\]的流量数值, 单位为*G*; 如想限制该用户为2\\.3G流量,则输入\"2\\.3\", 不限则输入\"\\-1\"\n", user.Username) + getUserMarkdownV2Text(&user)
	case "expiryDate":
		msg.Text = fmt.Sprintf("请输入限制\\[`%s`\\]的到期时间, 格式为\"2066\\-06\\-06\"\n", user.Username) + getUserMarkdownV2Text(&user)
	default:
		msg.Text = "命令不存在!!!"
	}
	if _, err := global.Bot.Send(msg); err != nil {
		return errors.New("错误[kysafd]" + err.Error())
	}

	tx, err := dao.DB.Beginx()
	if err != nil {
		return errors.New("开启事务失败[fdsgafsd]" + err.Error())
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				panic(err)
			}
		}
	}()

	return dao.SaveSysVal(tx, strconv.FormatInt(update.CallbackQuery.Message.Chat.ID, 10)+"_step", update.CallbackQuery.Data)
}

func userInsert(update *tg.Update, info *model.SystemInfo) (err error) {
	msg := tg.NewMessage(update.Message.Chat.ID, "")
	msg.ReplyToMessageID = update.Message.MessageID //引用对话
	tx, err := dao.DB.Beginx()
	if err != nil {
		return errors.New("开启事务失败[glyasd]: " + err.Error())
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			err = errors.New("错误[podghj]: " + err.Error())
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var user model.Users
	if ok, _ := regexp.MatchString("^[a-zA-Z0-9_-]{4,64}$", update.Message.Text); !ok {
		msg.Text = "请输入用户名称, 64个字符以内的英文/数字/字符\n例:`210606_abc`"
		msg.ParseMode = "MarkdownV2"
		dao.SaveSysVal(tx, info.Key, info.Value) //为了更新时间字段update_time
		goto SEND2
	}

	err = dao.GetUsersByUserName(&user, update.Message.Text)
	if err == nil {
		msg.Text = strings.Replace(fmt.Sprintf("*用户`%s`已存在\\!\\!\\!*\n", update.Message.Text), `-`, `\-`, -1) + getUserMarkdownV2Text(&user)
		msg.ParseMode = "MarkdownV2"
		dao.SaveSysVal(tx, info.Key, info.Value)
		goto SEND2
	}

	err = dao.InsertEmptyUsers(tx, update.Message.Text)
	if err != nil {
		msg.Text = `系统错误, 请按"/start"重新设置4`
		dao.SaveSysVal(tx, info.Key, info.Value)
		goto SEND2
	}

	err = dao.GetUsersByUserNameTx(tx, &user, update.Message.Text)
	if err != nil {
		msg.Text = `系统错误, 请按"/start"重新设置5`
		dao.SaveSysVal(tx, info.Key, info.Value)
		goto SEND2
	}

	msg.Text = "*新增成功\\!\\!\\!*\n" + getUserMarkdownV2Text(&user)
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("修改此用户["+user.Username+"]", fmt.Sprintf("user_update@%d", user.Id)),
		),
	)

SEND2:
	_, err = global.Bot.Send(msg)
	return
}

func getUserMarkdownV2Text(user *model.Users) string {
	// quota := "不限"
	// if user.Quota != -1 {
	// 	quota = fmt.Sprintf("%.1f", float64(user.Quota)/dao.QuotaMax) + "G"
	// }
	// download := fmt.Sprintf("%.1f", float64(user.Download)/dao.QuotaMax)
	// upload := fmt.Sprintf("%.1f", float64(user.Upload)/dao.QuotaMax)
	// text := fmt.Sprintf("账号: `%s`\nid: %d\n限流: %s\n上行: %sG\n下行: %sG\n到期: %s", user.Username, user.Id, quota, upload, download, *user.ExpiryDate)
	text := fmt.Sprintf("账号: `%s`\nid: %d\n到期: %s", user.Username, user.Id, *user.ExpiryDate)
	return strings.Replace(strings.Replace(text, `-`, `\-`, -1), `.`, `\.`, -1)
}
