package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
	"k8s.io/klog/v2"
)

func deleteFile(fileName string) {

	if err := os.Remove(fileName); err != nil {
		klog.Fatal("Error when deleting converted file", err)
	}
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

type NullPoller struct {
}

func (p *NullPoller) Poll(b *tele.Bot, updates chan tele.Update, stop chan struct{}) {
}

func main() {

	klog.InitFlags(nil)

	if ex, err := exists(".env"); ex && err == nil {
		klog.Info("LOADING ENV")
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	TOKEN := os.Getenv("TOKEN")
	WEBHOOK_URL := os.Getenv("WEBHOOK_URL")
	PORT := os.Getenv("PORT")
	klog.Info("TOKEN", TOKEN)
	klog.Info("WEBHOOK_URL", WEBHOOK_URL)
	klog.Info("PORT", PORT)

	config := tele.Settings{
		Token:       TOKEN,
		Poller:      &NullPoller{},
		Synchronous: true,
		Verbose:     true,
		Offline:     true,
	}

	b, err := tele.NewBot(config)

	// b.ProcessUpdate()
	if err != nil {
		klog.Fatal(err)
		return
	}

	klog.Info("Bot started")

	now := time.Now().Format(time.RFC3339)
	nowFmt := now[:len(now)-6]

	b.Handle(tele.OnVoice, func(c tele.Context) error {
		klog.Infof("Handling voice message from %s", c.Chat().Username)

		if file, err := b.FileByID(c.Message().Voice.File.FileID); err != nil {
			klog.Fatal("Error when getting file by id", err)
		} else {
			if len(file.FilePath) > 0 {
				fileDownloadUrl := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", TOKEN, file.FilePath)
				downloadFileCmd := exec.Command("curl", fileDownloadUrl, "-o", file.FileID)
				if err := downloadFileCmd.Run(); err != nil {
					klog.Fatal("Error when downloading file", err)
				}
				convertedFileName := fmt.Sprintf("tg_%s.mp3", nowFmt)
				converFileCmd := exec.Command("ffmpeg", "-i", file.FileID, "-acodec", "libmp3lame", convertedFileName)
				if err := converFileCmd.Run(); err != nil {
					klog.Fatal("Error when converting file", err)
				}

				deleteFile(file.FileID)

				defer deleteFile(convertedFileName)

				retFile := tele.Audio{File: tele.FromDisk(convertedFileName), Title: convertedFileName, FileName: convertedFileName}
				klog.Infof("Responding to %s", c.Chat().Username)
				return c.Send(&retFile)
			} else {
				klog.Fatalf("File Path not defined for %s", file.FileID)

			}
		}
		return c.Send("C'è stato un problema con questo file, sorry :(")
	})

	b.Handle(tele.OnText, func(ctx tele.Context) error {
		return ctx.Send("I'm alive!")
	})

	go b.Start()

	server := gin.Default()

	server.POST("/", func(c *gin.Context) {
		jsonData, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			klog.Fatalf("Error when reading request body: %v", err)
		}

		var update tele.Update

		if err := json.Unmarshal(jsonData, &update); err != nil {
			klog.Fatalf("Error when unmarshaling update %v", err)
		}
		b.ProcessUpdate(update)
		c.Status(200)
	})

	server.Run()
}
