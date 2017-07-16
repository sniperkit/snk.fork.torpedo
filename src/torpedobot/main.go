package main

import (
	"flag"
	"fmt"

	"github.com/tb0hdan/torpedo_registry"
	"torpedobot/multibot"
)

const ProjectURL = "https://github.com/tb0hdan/torpedo"

// Global vars for versioning
var (
	BUILD      = "Not available"
	BUILD_DATE = "Not available"
	VERSION    = "Not available"
)

func main() {
	// Help handlers
	help_msg := "Get help using this command"
	torpedo_registry.Config.RegisterHandler("?", HelpProcessMessage)
	torpedo_registry.Config.RegisterHelp("?", help_msg)
	torpedo_registry.Config.RegisterHandler("h", HelpProcessMessage)
	torpedo_registry.Config.RegisterHelp("h", help_msg)
	torpedo_registry.Config.RegisterHandler("help", HelpProcessMessage)
	torpedo_registry.Config.RegisterHelp("help", help_msg)
	torpedo_registry.Config.RegisterHandler("stats", StatsProcessMessage)
	torpedo_registry.Config.RegisterHelp("stats", "Just system stats, nothing interesting")

	bot := multibot.New()
	bot.SetBuildInfo(BUILD, BUILD_DATE, VERSION, ProjectURL)
	// bot cfg
	torpedo_registry.Config.RegisterPreParser("slack", bot.ConfigureSlackBot)
	torpedo_registry.Config.RegisterPreParser("telegram", bot.ConfigureTelegramBot)
	torpedo_registry.Config.RegisterPreParser("jabber", bot.ConfigureJabberBot)
	torpedo_registry.Config.RegisterPreParser("skype", bot.ConfigureSkypeBot)
	torpedo_registry.Config.RegisterPreParser("kik", bot.ConfigureKikBot)
	torpedo_registry.Config.RegisterPreParser("line", bot.ConfigureLineBot)
	torpedo_registry.Config.RegisterPreParser("matrix", bot.ConfigureMatrixBot)
	torpedo_registry.Config.RegisterPreParser("facebook", bot.ConfigureFacebookBot)
	torpedo_registry.Config.RegisterPreParser("mongodb", bot.ConfigureMongoDBPlugin)

	bot.RunPreParsers()

	flag.Parse()

	torpedo_registry.Config.RegisterPostParser("facebook", bot.ParseFacebookBot)
	torpedo_registry.Config.RegisterPostParser("jabber", bot.ParseJabberBot)
	torpedo_registry.Config.RegisterPostParser("kik", bot.ParseKikBot)
	torpedo_registry.Config.RegisterPostParser("line", bot.ParseLineBot)
	torpedo_registry.Config.RegisterPostParser("matrix", bot.ParseMatrixBot)
	torpedo_registry.Config.RegisterPostParser("skype", bot.ParseSkypeBot)
	torpedo_registry.Config.RegisterPostParser("slack", bot.ParseSlackBot)
	torpedo_registry.Config.RegisterPostParser("telegram", bot.ParseTelegramBot)
	torpedo_registry.Config.RegisterPostParser("mongodb", bot.ParseMongoDBPlugin)

	bot.RunPostParsers()

	// Command handlers and help
	bot.RegisterHandlers(torpedo_registry.Config.GetHandlers())
	bot.RegisterHelp(torpedo_registry.Config.GetHelp())

	fmt.Println(torpedo_registry.Config.GetConfig())
	bot.RunBotsCSV(bot.RunSlackBot, torpedo_registry.Config.GetConfig()["slackapikey"], "!")
	bot.RunBotsCSV(bot.RunTelegramBot, torpedo_registry.Config.GetConfig()["telegramapikey"], "/")
	bot.RunBotsCSV(bot.RunJabberBot, torpedo_registry.Config.GetConfig()["jabberapikey"], "!")
	bot.RunBotsCSV(bot.RunSkypeBot, torpedo_registry.Config.GetConfig()["skypeapikey"], "!")
	bot.RunBotsCSV(bot.RunKikBot, torpedo_registry.Config.GetConfig()["kikapikey"], "!")
	bot.RunBotsCSV(bot.RunLineBot, torpedo_registry.Config.GetConfig()["lineapikey"], "!")
	bot.RunBotsCSV(bot.RunMatrixBot, torpedo_registry.Config.GetConfig()["matrixapikey"], "!")
	bot.RunBotsCSV(bot.RunFacebookBot, torpedo_registry.Config.GetConfig()["facebookapikey"], "!")

	// start plugin coroutines (if any) after connecting to accounts
	bot.RunCoroutines()
	bot.RunLoop()
}
