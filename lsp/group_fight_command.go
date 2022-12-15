package lsp

import (
	"math/rand"

	"github.com/Sora233/DDBOT/lsp/mmsg"

	mirai_client "github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/alecthomas/kong"
)

func (lgc *LspGroupCommand) FightCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	_, output := lgc.parseCommandSyntax(
		&struct{}{}, lgc.CommandName(),
		kong.Description("at 一位群友打ta，或者留空随机打一位无辜群友"),
	)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	var err error

	// find if the message at someone
	var atList []*message.AtElement
	for _, e := range lgc.msg.Elements {
		if e.Type() == message.At {
			switch ae := e.(type) {
			case *message.AtElement:
				atList = append(atList, ae)
			default:
				log.Errorf("cast message element to AtElement failed")
				lgc.textReply("打人失败 - 内部错误 可能是网线过不去")
				return
			}
		}
	}

	var victimInfo *mirai_client.GroupMemberInfo

	var msg *mmsg.MSG
	if len(atList) != 0 {
		msg = mmsg.NewMSG()
		for _, ae := range atList {
			log.WithField("target", ae.Target).
				WithField("display", ae.Display).
				WithField("subtype", ae.SubType).
				Infof("atElement exists")

			victimInfo, err = (*lgc.bot.Bot).GetMemberInfo(lgc.msg.GroupCode, ae.Target)
			if err != nil {
				lgc.textSend("打人失败 - 内部错误")
				log.Errorf("fight GetMemberInfo error: %v", err)
				return
			}

			var victimDisplayName string
			if victimInfo.CardName != "" {
				victimDisplayName = victimInfo.CardName
			} else {
				victimDisplayName = victimInfo.Nickname
			}

			victimMsg := lgc.templateMsg(
				"command.group.fight.tmpl",
				map[string]interface{}{
					"victim_uin":         victimInfo.Uin,
					"victim_displayname": victimDisplayName,
				},
			)
			msg.Append(victimMsg.Elements()...)
		}
	} else {
		// did not at, randomly pick a victim
		var groupInfo *mirai_client.GroupInfo
		groupInfo, err = (*lgc.bot.Bot).GetGroupInfo(lgc.msg.GroupCode)
		if err != nil {
			lgc.textSend("打人失败 - 内部错误")
			log.Errorf("fight GetGroupInfo error: %v", err)
			return
		}
		var groupMembers []*mirai_client.GroupMemberInfo
		groupMembers, err = (*lgc.bot.Bot).GetGroupMembers(groupInfo)
		if err != nil {
			lgc.textSend("打人失败 - 内部错误")
			log.Errorf("fight GetGroupMembers error: %v", err)
			return
		}

		victimInfo = groupMembers[rand.Intn(len(groupMembers))]

		var victimDisplayName string
		if victimInfo.CardName != "" {
			victimDisplayName = victimInfo.CardName
		} else {
			victimDisplayName = victimInfo.Nickname
		}

		msg = lgc.templateMsg(
			"command.group.fight.tmpl",
			map[string]interface{}{
				"victim_uin":         victimInfo.Uin,
				"victim_displayname": victimDisplayName,
			},
		)
	}

	lgc.reply(msg)
	return
}
