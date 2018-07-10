package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pelletier/go-toml"

	"github.com/bwmarrin/discordgo"
)

var (
	token  string
	config Config
)

type DiscordInfo struct {
	Token string
}
type DatabaseInfo struct {
	Server   string
	User     string
	Password string
}
type Config struct {
	Database DatabaseInfo
	Discord  DiscordInfo
}

func init() {
	contents, err := ioutil.ReadFile("config.toml")
	if err != nil {
		panic(err)
	}
	config = Config{}
	err = toml.Unmarshal(contents, &config)
	if err != nil {
		panic(err)
	}
}

func main() {
	dg, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatalln("error creating Discord session", err)
	}
	dg.Close()
	dg, err = discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatalln("error creating Discord session", err)
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		log.Fatalln("error opening connection", err)
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	//<-sc
	go func() {
		sig := <-sc
		log.Printf("captured %v.\n", sig)
		dg.Close()
		done <- true
	}()

	fmt.Println("Waiting for CTRL-C")
	<-done
	fmt.Println("Goodbye!")
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Got a message!")
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	if strings.Contains(strings.ToLower(m.Content), "third") {
		s.ChannelMessageSend(m.ChannelID, "You're third! Woo?")
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}
