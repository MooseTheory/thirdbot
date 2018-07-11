package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/pelletier/go-toml"

	"github.com/bwmarrin/discordgo"
	"github.com/go-sql-driver/mysql"
)

var (
	token  string
	config Config
	db     *sql.DB
)

type DiscordInfo struct {
	Token string
}
type DatabaseInfo struct {
	Server       string
	Port         int
	User         string
	Password     string
	DatabaseName string
}
type Config struct {
	Database DatabaseInfo
	Discord  DiscordInfo
	Comments CommentInfo
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
	err = getDbConn()
	if err != nil {
		log.Fatalln("error connecting to database", err)
	}
	defer db.Close()

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		log.Fatalln("error opening connection", err)
	}
	err = prepareDatabase()
	if err != nil {
		log.Fatalln("error connecting to database", err)
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sc)
	//<-sc
	go func() {
		<-sc
		dg.Close()
		done <- true
	}()

	<-done
	fmt.Println("Goodbye!")
	dg.Close()
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
		err := checkIfThird(s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error logging your third! Sucks to be you.")
			return
		}
	}
}

func checkIfThird(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM thirds WHERE date BETWEEN ? AND ?")
	if err != nil {
		return err
	}
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	tomorrow := midnight.Add(time.Hour * 24)
	numResults := 0
	err = stmt.QueryRow(midnight, tomorrow).Scan(&numResults)
	if err != nil {
		return err
	}
	if numResults == 0 {
		err = addThird(s, m)
		s.ChannelMessageSend(m.ChannelID, config.Comments.getThirdComment())
		return err
	}
	return
}

func addThird(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	stmt, err := db.Prepare("INSERT INTO thirds(userid, date) VALUES(?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(m.Author.ID, time.Now())
	if err != nil {
		return err
	}

	return nil
}

var createThirdsStatement = "CREATE TABLE IF NOT EXISTS`thirds`(" +
	"`id` int NOT NULL AUTO_INCREMENT," +
	"`userid` varchar(64) NOT NULL," +
	"`date` datetime NOT NULL," +
	"PRIMARY KEY (`id`)" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Tracks Thirds';"

func prepareDatabase() (err error) {
	_, err = db.Exec(createThirdsStatement)
	if err != nil {
		return err
	}
	return
}

func getDbConn() (err error) {
	var addr string
	if config.Database.Server != "" {
		if config.Database.Port != 0 {
			addr = fmt.Sprintf("%s:%d", config.Database.Server, config.Database.Port)
		} else {
			addr = config.Database.Server
		}
	}
	conf := mysql.Config{
		User:   config.Database.User,
		Passwd: config.Database.Password,
		DBName: config.Database.DatabaseName,
		Net:    "tcp",
		Addr:   addr,
	}
	conf.AllowNativePasswords = true
	db, err = sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		return err
	}
	return nil
}
