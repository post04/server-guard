package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/post04/server-guard/events"

	"github.com/bwmarrin/discordgo"
)

var (
	c *events.Config
)

func init() {
	rand.Seed(time.Now().UnixNano())
	f, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(f, &c)
	if err != nil {
		log.Fatal(err)
	}
	files, err := os.ReadDir("./emojis/")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		c.Emojis = append(c.Emojis, strings.Split(file.Name(), ".")[0])
	}
	c.CaptchaListeners = make(map[string]*events.Captcha)
	c.PreviousJoins = make(map[string][]*events.JoinedUser)
	fmt.Println(c.Emojis)
}

func ready(session *discordgo.Session, evt *discordgo.Ready) {
	fmt.Printf("Logged in under: %s#%s\n", evt.User.Username, evt.User.Discriminator)
	go c.ClearBots(5 * time.Minute)
	go c.CheckBots(5 * time.Second)
}

func main() {
	bot, err := discordgo.New("Bot " + c.Token)
	if err != nil {
		log.Fatal("ERROR LOGGING IN", err)
	}
	c.Session = bot
	bot.Identify.Intents = discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessageReactions | discordgo.IntentsDirectMessageReactions
	bot.AddHandler(ready)
	bot.AddHandler(c.OnMemberJoin)
	bot.AddHandler(c.OnReaction)
	err = bot.Open()
	if err != nil {
		log.Fatal("ERROR OPENING CONNECTION", err)
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc
	bot.Close()
}
