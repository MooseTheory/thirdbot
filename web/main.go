package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-sql-driver/mysql"
	toml "github.com/pelletier/go-toml"
)

var (
	config Config
	db     *sql.DB
	dg     *discordgo.Session
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

// Config holds the total configuration information.
type Config struct {
	Database DatabaseInfo
	Discord  DiscordInfo
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
	err = getDbConn()
	if err != nil {
		panic(err)
	}
	dg, err = discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatalln("error creating Discord session", err)
	}
}

func main() {
	http.HandleFunc("/last", handleLast)
	http.HandleFunc("/lasts", handleGetThirds)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleLast(w http.ResponseWriter, r *http.Request) {
	stmt, err := db.Prepare("SELECT `userid`, `date` FROM `thirds` ORDER BY `date` DESC LIMIT 1")
	if err != nil {
		log.Println(err)
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Println(err)
		return
	}
	var last lastInfo
	for rows.Next() {
		var userID string
		var date time.Time
		err = rows.Scan(&userID, &date)
		if err != nil {
			fmt.Println(err)
			continue
		}
		user, err := dg.User(userID)
		if err != nil {
			fmt.Println("dg.User error")
			fmt.Println(err)
			continue
		}
		last.Date = date.In(tz)
		last.UserID = userID
		last.UserName = user.Username
	}
	resp, err := json.Marshal(last)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(resp))
}

func handleGetThirds(w http.ResponseWriter, r *http.Request) {
	stmt, err := db.Prepare("SELECT `userid`, `date` FROM `thirds` ORDER BY `date` DESC")
	if err != nil {
		log.Println(err)
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Println(err)
		return
	}
	//var lasts []lastInfo
	lasts := []lastInfo{}
	for rows.Next() {
		var last lastInfo
		var userID string
		var date time.Time
		err = rows.Scan(&userID, &date)
		if err != nil {
			fmt.Println(err)
			continue
		}
		user, err := dg.User(userID)
		if err != nil {
			fmt.Println("dg.User error")
			fmt.Println(err)
			continue
		}
		last.Date = date.In(tz)
		last.UserID = userID
		last.UserName = user.Username
		lasts = append(lasts, last)
	}
	resp, err := json.Marshal(lasts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(resp))
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
	conf.ParseTime = true
	db, err = sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		return err
	}
	return nil
}
