package lsp

import (
	"math/rand"
	"os"
	"time"

	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/utils"

	"github.com/alecthomas/kong"
)

func (lgc *LspGroupCommand) DivinationCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	_, output := lgc.parseCommandSyntax(&struct{}{}, lgc.CommandName(), kong.Description("/占卜 赛博算命"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	date := time.Now().Format("20060102")
	var divineKey = localdb.Key("Divination", lgc.uin(), date)
	var divinationSN int64
	var divinationExist = false
	err := localdb.RWCover(func() error {
		var lErr error
		divinationSN, lErr = localdb.GetInt64(divineKey, localdb.IgnoreNotFoundOpt())
		if lErr != nil {
			return lErr
		}
		if localdb.Exist(divineKey) {
			divinationExist = true
			return lErr
		}

		divinationSN = int64(rand.Intn(len(utils.Divinations)))
		lErr = localdb.SetInt64(divineKey, divinationSN, localdb.SetExpireOpt(time.Hour*24*3))
		if lErr != nil {
			return lErr
		}
		return nil
	})
	if err != nil {
		lgc.textSend("占卜失败 - 内部错误")
		log.Errorf("divination error: %v", err)
		return
	}

	divination := utils.Divinations[divinationSN]

	divInscription, err := os.ReadFile(divination.InscriptionPath)
	if err != nil {
		lgc.textSend("占卜失败 - 内部错误")
		log.Errorf("divination open inscription file error: %v", err)
		return
	}
	inscription := string(divInscription)

	// reply the divination to group
	lgc.reply(
		lgc.templateMsg(
			"command.group.divination.tmpl",
			map[string]interface{}{
				"divination_exist": divinationExist,
				"divination_title": divination.Title,
				"divination_image": divination.ImagePath,
				"inscription":      inscription,
			},
		),
	)

	return
}
