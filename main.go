package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

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

	var config tele.Settings

	if WEBHOOK_URL == "" {
		// delete bot webhook
		_, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook", TOKEN))

		if err != nil {
			log.Fatal(err)
		}
		config = tele.Settings{
			Token:  TOKEN,
			Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		}
	} else {

		config = tele.Settings{
			Token: TOKEN,
			Poller: &tele.Webhook{
				Endpoint:       &tele.WebhookEndpoint{PublicURL: WEBHOOK_URL},
				AllowedUpdates: []string{"callback_query", "message"},
				Listen:         fmt.Sprintf(":%s", PORT),
			},
		}
	}

	b, err := tele.NewBot(config)
	if err != nil {
		klog.Fatal(err)
		return
	}

	klog.Info("Bot started")

	now := time.Now().Format(time.RFC3339)
	nowFmt := now[:len(now)-6]

	b.Handle(tele.OnVoice, func(c tele.Context) error {
		c.Send(fmt.Sprintf("All right! Processing %ds", c.Message().Voice.Duration))
		klog.Infof("Handling voice message from %s", c.Chat().Username)

		if file, err := b.FileByID(c.Message().Voice.File.FileID); err != nil {
			klog.Fatal("Error when getting file by id", err)
		} else {
			if len(file.FilePath) > 0 {
				fileDownloadUrl := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", TOKEN, file.FilePath)

				if err := DownloadFile(file.FileID, fileDownloadUrl); err != nil {
					klog.Fatal("Error when downloading file", err)
				}
				c.Send("File downloaded")

				convertedFileName := fmt.Sprintf("tg_%s.mp3", nowFmt)
				converFileCmd := exec.Command("ffmpeg", "-i", file.FileID, "-acodec", "libmp3lame", convertedFileName)
				if err := converFileCmd.Run(); err != nil {
					klog.Fatal("Error when converting file", err)
				}
				c.Send("File processed")

				deleteFile(file.FileID)

				defer deleteFile(convertedFileName)

				retFile := tele.Audio{File: tele.FromDisk(convertedFileName), Title: convertedFileName, FileName: convertedFileName}
				klog.Infof("Responding to %s", c.Chat().Username)
				return c.Send(&retFile)
			} else {
				klog.Fatalf("File Path not defined for %s", file.FileID)

			}
		}
		return c.Send("Sorry! There has been an issue with this voice message 😔")
	})

	b.Handle(tele.OnText, func(ctx tele.Context) error {
		return ctx.Send("Hi! I'm here, send me a voice message and I'll hosw you magic!")
	})

	b.Handle("/help", func(ctx tele.Context) error {
		return ctx.Send("Hello👋\nI'm bot that can turn any voice message into file to be shared outside of telegram.\n\nIf you like me, please make a little donation to my creator, thank you!")
	})

	b.Handle("/privacy", func(ctx tele.Context) error {
		return ctx.Send("Hello👋\nI take privacy very seriously:\n- I do not store any of the messages and files we exchange\n- I do not store any information about you.\nYou simply write, I respond, that's it!")
	})

	b.Handle("/donate", func(ctx tele.Context) error {
		return ctx.Send("Hello👋\nIf you would like to thank my creator for my services, please donate using paypal to this email gabriele.filaferro@gmail.com!")
	})

	b.Start()
}
