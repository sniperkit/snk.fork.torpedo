package main

import (
        "encoding/base64"
        "fmt"
        "io/ioutil"
        "log"
        "os"
        "strings"

        "github.com/nlopes/slack"
       )


func GetCreateChannelDir(channel string) (channelDir string, err error) {
    wd, err := os.Getwd()
    if err != nil {
        return
    }
    channelDirPath := fmt.Sprintf("%s%s%s%s%s", wd, string(os.PathSeparator), "data", string(os.PathSeparator),  channel)
    err = os.MkdirAll(channelDirPath, 0755)
    if err == nil {
        channelDir = channelDirPath
    }
    return
}


func GetChannelFile(channel, message string) (channelFile, mimetype string, err error) {
    wd, err := GetCreateChannelDir(channel)
    if err != nil {
        return
    }
    // TODO: Add message permutations
    encoded := base64.URLEncoding.EncodeToString([]byte(strings.TrimSpace(message)))
    fname := fmt.Sprintf("%s%s%s", wd, string(os.PathSeparator), encoded)
    if FileExists(fname) {
        mimetype, _, _, err = GetMIMEType(fname)
        if err != nil {
            return
        }
        channelFile = fname
    }
    return
}


func SetChannelFile(channel, message string) (result string, err error) {
    wd, err := GetCreateChannelDir(channel)
    if err != nil {
        return
    }
    url_formatted := strings.Split(message, " ")[0]
    url := UnformatURL(url_formatted)
    if ! (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
        result = "No valid URL found"
        return
    }
    destination := strings.TrimSpace(strings.TrimLeft(message, url_formatted))
    if destination == "" {
        result = "No valid destination found"
        return
    }
    // Check if target already exists before downloading new one
    encoded := base64.URLEncoding.EncodeToString([]byte(destination))
    new_name := fmt.Sprintf("%s%s%s", wd, string(os.PathSeparator), encoded)
    if FileExists(new_name) {
        result = "Destination already exists, set skipped. Use `!rmimg destination` to remove."
        return
    }
    fname, _, is_image, err := DownloadToTmp(url)
    if is_image {
        err = os.Rename(fname, new_name)
        if err != nil {
            result = "There was an issue with setting image"
        } else {
            result = "Image set"
        }
    }
    return
}

func ListChannelFiles(channel string) (files []string, err error) {
    wd, err := GetCreateChannelDir(channel)
    if err != nil {
        return
    }

    file_names, err := ioutil.ReadDir(wd)
    if err != nil {
        log.Fatal(err)
        return
    }

    for _, file := range file_names {
        files = append(files, file.Name())
    }
    return
}

func GetSetImageProcessMessage(api *slack.Client, event *slack.MessageEvent) {
    var params slack.PostMessageParameters
    requestedFeature, command, message := GetRequestedFeature(event.Text)
    if command != "" {
        switch requestedFeature {
        case "!getimg":
            fpath, mimetype, err := GetChannelFile(event.Channel, command)
            if fpath != "" {
                ChannelsUploadImage([]string{event.Channel}, command, fpath, mimetype, api)
                return
            } else {
                message = fmt.Sprintf("%+v", err)
            }
        case "!setimg":
            msg, err := SetChannelFile(event.Channel, command)
            if err != nil {
                message = fmt.Sprintf("%+v", err)
            } else {
                message = msg
            }
        case "!listimg", "!lsimg":
            files, err := ListChannelFiles(event.Channel)
            if err != nil {
                message = "An error occured while retrieving image file list"
            } else {
                message = ""
                for _, fname := range files {
                    msg, err := base64.URLEncoding.DecodeString(fname)
                    if err != nil {
                        continue
                    }
                    message += fmt.Sprintf("`%s`\n", msg)
                }
                if message == "" {
                    message = "No files found, upload using !setimg first"
                } else {
                    message = fmt.Sprintf("Available image files:\n%s", message)
                }
            }
        case "!rmimg":
            fpath, _, _ := GetChannelFile(event.Channel, command)
            if fpath != "" {
                err := os.Remove(fpath)
                if err != nil {
                    message = fmt.Sprintf("An error occured while trying to remove `%s`", command)
                } else {
                    message = fmt.Sprintf("Requested file `%s` was removed", command)
                }
            } else {
                message = fmt.Sprintf("Requested file `%s` was not found", command)
            }
        default:
            // should never get here
            message = "Unknown feature requested"
        }
    }
    postMessage(event.Channel, message, api, params)
}
