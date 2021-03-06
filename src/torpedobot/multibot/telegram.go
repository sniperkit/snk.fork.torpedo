package multibot

import (
	"os"
	"time"

	"flag"
	"regexp"

	common "github.com/tb0hdan/torpedo_common"

	"fmt"

	"github.com/tb0hdan/torpedo_registry"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var TelegramAPIKey *string

func ToTelegramAttachment(rm torpedo_registry.RichMessage, channel int64) (msg tgbotapi.Chattable, fname string) {
	cu := &common.Utils{}
	fname, _, is_image, err := cu.DownloadToTmp(rm.ImageURL)
	if is_image && err == nil {
		msg = tgbotapi.NewPhotoUpload(channel, fname)
	}
	return
}

func HandleTelegramMessage(channel interface{}, message string, tba *TorpedoBotAPI, richmsgs []torpedo_registry.RichMessage) {
	switch api := tba.API.(type) {
	case *tgbotapi.BotAPI:
		var msg tgbotapi.Chattable
		var tmp string
		if len(richmsgs) > 0 && !richmsgs[0].IsEmpty() {
			msg, tmp = ToTelegramAttachment(richmsgs[0], channel.(int64))
			api.Send(tgbotapi.NewMessage(channel.(int64), richmsgs[0].Text))
		} else {
			msg = tgbotapi.NewMessage(channel.(int64), message)
		}
		api.Send(msg)
		if tmp != "" {
			os.Remove(tmp)
		}
	}
}

func (tb *TorpedoBot) ConfigureTelegramBot(cfg *torpedo_registry.ConfigStruct) {
	TelegramAPIKey = flag.String("telegram", "", "Comma separated list of Telegram bot keys")
}

func (tb *TorpedoBot) ParseTelegramBot(cfg *torpedo_registry.ConfigStruct) {
	cfg.SetConfig("telegramapikey", *TelegramAPIKey)
	if cfg.GetConfig()["telegramapikey"] == "" {
		cfg.SetConfig("telegramapikey", common.GetStripEnv("TELEGRAM"))
	}
}

func (tb *TorpedoBot) RunTelegramBot(apiKey, cmd_prefix string) {
	account := &torpedo_registry.Account{
		APIKey:        apiKey,
		CommandPrefix: cmd_prefix,
	}
	torpedo_registry.Accounts.AppendAccounts(account)
	tb.RunTelegramBotAccount(account)
}

func (tb *TorpedoBot) RunTelegramBotAccount(account *torpedo_registry.Account) {
	tb.Stats.ConnectedAccounts += 1

	cu := &common.Utils{}

	logger := cu.NewLog("telegram-bot")

	api, err := tgbotapi.NewBotAPI(account.APIKey)
	if err != nil {
		logger.Panic(err)
	}

	if torpedo_registry.Config.GetConfig()["debug"] == "yes" {

		api.Debug = true
	}

	logger.Printf("Authorized on account %s", api.Self.UserName)
	account.Connection.ReconnectCount += 1
	account.API = api

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := api.GetUpdatesChan(u)

	tb.RegisteredProtocols["*tgbotapi.BotAPI"] = HandleTelegramMessage

	for update := range updates {
		if update.Message == nil {
			continue
		}

		jitter := int64(time.Now().Unix()) - int64(update.Message.Date)

		if jitter > 10 {
			continue
		}

		// handle multible bot presence
		r := regexp.MustCompile(`(?i)@(.+)bot`)
		message := r.ReplaceAllString(update.Message.Text, "")

		logger.Printf("[%s] %s\n", update.Message.From.UserName, message)

		botApi := &TorpedoBotAPI{}
		botApi.API = api
		botApi.Bot = tb
		botApi.CommandPrefix = account.CommandPrefix
		botApi.UserProfile = &torpedo_registry.UserProfile{ID: fmt.Sprintf("%v", update.Message.From.ID), Nick: update.Message.From.UserName}
		botApi.Me = "torpedobot"

		go tb.processChannelEvent(botApi, update.Message.Chat.ID, message)

	}
	tb.Stats.ConnectedAccounts -= 1
}
