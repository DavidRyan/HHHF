package main

import (
	_ "time/tzdata"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

type Video struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func loadVideos(file string) ([]Video, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var videos []Video
	return videos, json.Unmarshal(data, &videos)
}

func post(dg *discordgo.Session, channelID, file, msg string) {
	videos, err := loadVideos(file)
	if err != nil {
		log.Println("error loading videos:", err)
		return
	}
	if len(videos) == 0 {
		log.Println("no videos in dataset:", file)
		return
	}
	v := videos[rand.Intn(len(videos))]
	full := fmt.Sprintf("%s\n**%s**\n%s", msg, v.Title, v.URL)
	if _, err := dg.ChannelMessageSend(channelID, full); err != nil {
		log.Println("error sending message:", err)
	} else {
		log.Printf("posted: %s", v.Title)
	}
}

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	channelID := os.Getenv("DISCORD_CHANNEL_ID")

	if token == "" || channelID == "" {
		log.Fatal("DISCORD_BOT_TOKEN and DISCORD_CHANNEL_ID must be set")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("error creating Discord session:", err)
	}

	if err := dg.Open(); err != nil {
		log.Fatal("error opening Discord connection:", err)
	}
	defer dg.Close()

	log.Println("bot connected to Discord")
	post(dg, channelID, "videos.json", "Happy Harry Hood Friday!")

	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		log.Fatal("error loading timezone:", err)
	}

	c := cron.New(cron.WithLocation(loc))
	_, err = c.AddFunc("0 9 * * 5", func() {
		post(dg, channelID, "videos.json", "Happy Harry Hood Friday!")
	})
	if err != nil {
		log.Fatal("error scheduling Hood cron:", err)
	}
	_, err = c.AddFunc("0 16 * * 5", func() {
		post(dg, channelID, "franklins_tower_videos.json", "HFTF")
	})
	if err != nil {
		log.Fatal("error scheduling Franklin's Tower cron:", err)
	}
	c.Start()
	defer c.Stop()

	log.Println("scheduler running — Hood at 9am, Franklin's Tower at 4pm, every Friday CST")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Println("shutting down")
}
