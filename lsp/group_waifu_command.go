package lsp

import (
	"math/rand"
	"time"

	localdb "github.com/Sora233/DDBOT/lsp/buntdb"

	mirai_client "github.com/Mrs4s/MiraiGo/client"
	"github.com/alecthomas/kong"
)

func (lgc *LspGroupCommand) WaifuCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	_, output := lgc.parseCommandSyntax(&struct{}{}, lgc.CommandName(), kong.Description("/今日老婆 抽取今天的老婆！"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	date := time.Now().Format("20060102")
	waifuKey := localdb.Key("Waifu", lgc.groupCode(), lgc.uin(), date)
	log.Infof("waifuKey: %s", waifuKey)
	var waifuInfo *mirai_client.GroupMemberInfo
	var waifuExist bool
	// pre-check daily waifu existence
	err := localdb.RCover(func() error {
		var lErr error

		// check database if the user had a waifu in the group today already
		lErr = localdb.GetJson(waifuKey, &waifuInfo, localdb.IgnoreNotFoundOpt())
		if lErr != nil {
			return lErr
		}
		if localdb.Exist(waifuKey) {
			waifuExist = true
		}
		return nil
	})
	if err != nil {
		lgc.textSend("贴贴老婆失败 - 内部错误")
		log.Errorf("daily waifu precheck error: %v", err)
		return
	}

	if waifuExist {
		// if waifu exists, log the old waifu and reply
		log = log.WithField("old_waifu_uin", waifuInfo.Uin).
			WithField("old_waifu_nickname", waifuInfo.Nickname).
			WithField("old_waifu_cardname", waifuInfo.CardName)
		log.Infof("old waifu found")
		var waifu_displayname string
		if waifuInfo.CardName != "" {
			waifu_displayname = waifuInfo.CardName
		} else {
			waifu_displayname = waifuInfo.Nickname
		}
		lgc.reply(
			lgc.templateMsg(
				"command.group.waifu.tmpl",
				map[string]interface{}{
					"waifu_exist":       waifuExist,
					"waifu_uin":         waifuInfo.Uin,
					"waifu_displayname": waifu_displayname,
					// "waifu_icon_url": fmt.Sprintf("http://q2.qlogo.cn/headimg_dl?dst_uin=%d&spec=100", waifuInfo.Uin),
				},
			),
		)
		return
	}

	// no waifu yet today
	// get group member list from the group message
	groupInfo, err := (*lgc.bot.Bot).GetGroupInfo(lgc.msg.GroupCode)
	if err != nil {
		lgc.textSend("贴贴老婆失败 - 内部错误")
		log.Errorf("waifu GetGroupInfo error: %v", err)
		return
	}
	groupMembers, err := (*lgc.bot.Bot).GetGroupMembers(groupInfo)
	if err != nil {
		lgc.textSend("贴贴老婆失败 - 内部错误")
		log.Errorf("waifu GetGroupMembers error: %v", err)
		return
	}

	// get a random user from group member list
	waifuInfo = groupMembers[rand.Intn(len(groupMembers))]
	// Fixme: this is a temporary fix to avoid infinite pointer loop during json Marshal
	//		should use a dedicated structure to store waifu info rather than storing the original GroupMemberInfo
	waifuInfo.Group = nil

	log = log.WithField("new_waifu_uin", waifuInfo.Uin).
		WithField("new_waifu_nickname", waifuInfo.Nickname).
		WithField("new_waifu_cardname", waifuInfo.CardName)
	log.Infof("new waifu rolled!")

	// record daily waifu of the user in this group in database
	err = localdb.RWCover(func() error {
		var lErr error

		if localdb.Exist(waifuKey) {
			waifuExist = true
			// another goroutine selected a waifu between the waifu selection
			lErr = localdb.GetJson(waifuKey, &waifuInfo, localdb.IgnoreNotFoundOpt())
			if lErr != nil {
				return lErr
			}
			// if waifu exists, log waifu and return
			log = log.WithField("old_waifu_uin", waifuInfo.Uin).
				WithField("old_waifu_nickname", waifuInfo.Nickname).
				WithField("old_waifu_cardname", waifuInfo.CardName)
			log.Infof("old waifu found After rolled new waifu")
			return nil
		}

		lErr = localdb.SetJson(
			waifuKey,
			waifuInfo,
			localdb.SetExpireOpt(time.Hour*24*3),
		)
		if lErr != nil {
			return lErr
		}
		return nil
	})
	if err != nil {
		lgc.textSend("贴贴老婆失败 - 内部错误")
		log.Errorf("waifu insert new daily waifu error: %v", err)
		return
	}
	// reply the waifu to group
	var waifu_displayname string
	if waifuInfo.CardName != "" {
		waifu_displayname = waifuInfo.CardName
	} else {
		waifu_displayname = waifuInfo.Nickname
	}
	lgc.reply(
		lgc.templateMsg(
			"command.group.waifu.tmpl",
			map[string]interface{}{
				"waifu_exist":       waifuExist,
				"waifu_uin":         waifuInfo.Uin,
				"waifu_displayname": waifu_displayname,
				// "waifu_icon_url": fmt.Sprintf("http://q2.qlogo.cn/headimg_dl?dst_uin=%d&spec=100", waifuInfo.Uin),
			},
		),
	)
	return
}
