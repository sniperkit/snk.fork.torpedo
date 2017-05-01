package main

import (
	"./pinterest"
	"github.com/nlopes/slack"
	"fmt"
	"flag"
	"strings"
)

var Token = flag.String("pinterest_token", "", "Pinterest Client Token")

func PinterestProcessMessage (api *slack.Client, event *slack.MessageEvent) {
	var params slack.PostMessageParameters
	requestedFeature, command, message := GetRequestedFeature(event.Text, "board")
	command = strings.Split(command, " ")[0]

	switch command {
	case "board":
		board := strings.TrimSpace(strings.TrimPrefix(event.Text, fmt.Sprintf("%s %s", requestedFeature, command)))
		if board != "" {
			api := pinterest.New(*Token)
			images, err := api.GetImagesForBoard(board)
			if err != nil {
				return
			}
			attachment := slack.Attachment{
				Color:   "#36a64f",
				Text:    board,
				Title: board,
				TitleLink: pinterest.PINTEREST_API_BASE + board,
				ImageURL: images[0],
			}
			params.Attachments = []slack.Attachment{attachment}
		}
	default:
		if command != "" {
			message = fmt.Sprintf("Command %s not available yet", command)
		}
	}

	postMessage(event.Channel, message, api, params)
}