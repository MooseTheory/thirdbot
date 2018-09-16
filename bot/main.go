package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/moosetheory/thirdbot/thirddb"
	"github.com/pelletier/go-toml"

	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"
	"gopkg.in/gemnasium/logrus-graylog-hook.v2"
)

var (
	config     Config
	conn       thirddb.ThirdConn
	dg         *discordgo.Session
	offsetTime time.Time
)

// DiscordInfo holds the config token for the discord bot.
type DiscordInfo struct {
	Token string
}

// DatabaseInfo holds the config information for the mariadb connection
type DatabaseInfo struct {
	Server       string
	Port         int
	User         string
	Password     string
	DatabaseName string
}

// GraylogInfo holds configuration information for graylog.
type GraylogInfo struct {
	Host string
	Port int
}

// Config holds the total configuration information.
type Config struct {
	Database DatabaseInfo
	Discord  DiscordInfo
	Comments CommentInfo
	Graylog  GraylogInfo
}

type lastInfo struct {
	UserID   string
	UserName string
	Date     time.Time
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
	hook := graylog.NewGraylogHook(fmt.Sprintf("%s:%d", config.Graylog.Host, config.Graylog.Port), map[string]interface{}{})
	log.AddHook(hook)
	log.Info("Starting")
}

func main() {
	dg, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatalln("error creating Discord session", err)
	}
	conn, err = thirddb.NewConn(config.Database.User, config.Database.Password,
		config.Database.DatabaseName, "tcp", config.Database.Server,
		config.Database.Port)
	if err != nil {
		log.Fatalln("error connecting to database", err)
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		log.Fatalln("error opening connection", err)
	}

	sc := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sc)
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				// TIMER!
				setStatus(dg)
			case <-sc:
				dg.Close()
				done <- true
			}
		}
	}()
	setStatus(dg)
	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	<-done
	fmt.Println("Goodbye!")
	dg.Close()
}

func setStatus(dg *discordgo.Session) {
	dg.UpdateStatus(0, "#help for a list of commands")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// These are commands, since I don't want to step on firstbot's ! command
	if strings.HasPrefix(m.Content, "#") {
		runCommand(s, m, strings.TrimLeft(m.Content, "#"))
	}

	if strings.Contains(strings.ToLower(m.Content), "third") {
		isThird, err := conn.CheckIfThird(offsetTime)
		if err != nil {
			log.Errorln("error detecting third", err)
			sendChatError(s, m, "error detecting if you are third")
			return
		}
		if isThird {
			err = pickNewOffset()
			if err != nil {
				log.Errorln("error getting time zone for third", err)
				sendChatError(s, m, "error detecting if you are third")
				return
			}
			err = conn.AddThird(m.Author.ID)
			if err != nil {
				log.Errorln("error adding third", err)
				sendChatError(s, m, "error saving your third. Ha ha")
				return
			}
			sendChatMessage(s, m, config.Comments.getThirdComment())
		}
	}
}

func pickNewOffset() (err error) {
	now := time.Now()
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		return
	}
	now = now.In(tz)
	r := rand.New(rand.NewSource(now.UnixNano()))
	hourOffset := r.Intn(3)
	minuteOffset := r.Intn(60)
	offsetTime = time.Date(now.Year(), now.Month(), now.Day(), hourOffset, minuteOffset, 0, 0, tz).AddDate(0, 0, 1)
	return nil
}

func sendChatError(s *discordgo.Session, m *discordgo.MessageCreate, msg string) (err error) {
	// Send errors
	errMsg := "**ERROR:**\n"
	errMsg += "`" + msg + "`"
	_, err = s.ChannelMessageSend(m.ChannelID, msg)
	return nil
}

func sendChatMessage(s *discordgo.Session, m *discordgo.MessageCreate, msg string) (err error) {
	// Send a message to chat
	_, err = s.ChannelMessageSend(m.ChannelID, msg)
	return err
}

func sendWhisper(s *discordgo.Session, m *discordgo.MessageCreate, msg string) (err error) {
	// Send a whisper
	s.ChannelMessageSend(m.ChannelID, "DM sent!")
	userChan, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error sending message! I'm broken!")
		return err
	}
	s.ChannelMessageSend(userChan.ID, msg)
	return nil
}
