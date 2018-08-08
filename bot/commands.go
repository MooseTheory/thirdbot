package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func runCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string) {
	switch strings.ToLower(command) {
	case "help":
		help(s, m)
	case "leaders":
		leaders(s, m)
	case "last":
		last(s, m)
	case "me":
		me(s, m)
	case "status":
		status(s, m)
	}
}

func help(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp := "**#leaders**\n"
	resp += "Returns the rather unimportant list of third leaders.\n"
	resp += "**#last**\n"
	resp += "Returns the last person to get a third.\n"
	resp += "**#me**\n"
	resp += "Show how many thirds you have, if you're good enough for third.\n"
	resp += "**#status**\n"
	resp += "Show the requestor's status\n"
	resp += "\n\n"
	resp += "I'm not cool enough to make sure a first or second happened yet.\n"
	resp += "I'm also not cool enough to wait until firstbot has reset for the day.\n"
	resp += "For now, you gotta deal with it. So. Deal with it.\n"
	resp += "https://i.imgur.com/mWNumm0.gif"

	sendMessage(s, m, resp)
}

func status(s *discordgo.Session, m *discordgo.MessageCreate) {
	g, err := s.Guild("218131283505709056")
	if err != nil {
		sendMessage(s, m, "I broke attempting to get the guild!")
		fmt.Println(err)
		return
	}
	resp := ""
	for _, p := range g.Presences {
		resp += fmt.Sprintf("%+v\n", p)
		resp += fmt.Sprintf("%+v\n", *p.User)
		user, err := s.User(p.User.ID)
		if err != nil {
			fmt.Println(err)
			continue
		}
		resp += fmt.Sprintf("%+v\n", user)
		if p.Game != nil && p.User.ID == m.Author.ID {
			resp += p.Game.Name + "\n"
		}
		if p.Game != nil {
			resp += p.Game.Name + " not author\n"
		}
	}
	s.ChannelMessageSend(m.ChannelID, resp)
}

func leaders(s *discordgo.Session, m *discordgo.MessageCreate) {
	stmt, err := db.Prepare("SELECT COUNT(*) AS `count`, userid FROM thirds GROUP BY userid ORDER BY count DESC")
	if err != nil {
		sendMessage(s, m, "I broke trying to do this!")
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		sendMessage(s, m, "I broke trying to do this!")
		return
	}
	defer rows.Close()
	resp := "**LEADERS**\n"
	resp += config.Comments.getLeaderHeader() + "\n"
	isFirst := true
	for rows.Next() {
		var count int
		var userID string
		err = rows.Scan(&count, &userID)
		if err != nil {
			continue
		}
		user, err := s.User(userID)
		if err != nil {
			continue
		}
		if !isFirst {
			resp += fmt.Sprintf("%s: %d\n%s\n", user.Username, count, config.Comments.getLeaderComment())
		} else {
			resp += fmt.Sprintf("%s: %d\n%s\n", user.Username, count, config.Comments.getFirstComment())
		}
		isFirst = false
	}
	s.ChannelMessageSend(m.ChannelID, resp)
}

func last(s *discordgo.Session, m *discordgo.MessageCreate) {
	stmt, err := db.Prepare("SELECT `userid`, `date` FROM `thirds` ORDER BY `date` DESC LIMIT 1")
	if err != nil {
		sendMessage(s, m, "I broke trying to do this! "+err.Error())
		fmt.Println(err)
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		sendMessage(s, m, "I broke trying to do this!"+err.Error())
		fmt.Println(err)
		return
	}
	defer rows.Close()
	var resp string
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		sendMessage(s, m, "I broke trying to do this!"+err.Error())
		fmt.Println(err)
		return
	}
	for rows.Next() {
		var userID string
		var date time.Time
		err = rows.Scan(&userID, &date)
		if err != nil {
			fmt.Println(err)
			continue
		}
		user, err := s.User(userID)
		if err != nil {
			sendMessage(s, m, "I broke trying to do this!"+err.Error())
			fmt.Println(err)
			continue
		}
		zonedDate := date.In(tz)
		resp = fmt.Sprintf("The last third was %s on %s, at %s.", user.Username, zonedDate.Format("Jan 02"), zonedDate.Format("3:04PM"))
	}
	if resp == "" {
		resp = "Ain't nobody been third yet! Slackers."
	}
	s.ChannelMessageSend(m.ChannelID, resp)
}

func me(s *discordgo.Session, m *discordgo.MessageCreate) {
	userID := m.Author.ID

	stmt, err := db.Prepare("SELECT COUNT(*) FROM `thirds` WHERE `userid` = ?")
	if err != nil {
		sendMessage(s, m, "I broke trying to do this! "+err.Error())
		return
	}
	var numResults int
	err = stmt.QueryRow(userID).Scan(&numResults)
	if err != nil {
		sendMessage(s, m, "I broke trying to do this!"+err.Error())
		return
	}
	user, err := s.User(userID)
	if err != nil {
		sendMessage(s, m, "I broke trying to do this!"+err.Error())
		return
	}
	var resp string
	if numResults >= 0 {
		var sIfNeeded = ""
		if numResults > 1 {
			sIfNeeded = "s"
		}
		resp = fmt.Sprintf("%s, you have **%d** third%s!", user.Username, numResults, sIfNeeded)
	} else {
		resp = fmt.Sprintf("Oh, I see %s, you think you're too good to get any thirds!", user.Username)
	}
	s.ChannelMessageSend(m.ChannelID, resp)
}

func sendMessage(s *discordgo.Session, m *discordgo.MessageCreate, resp string) {
	s.ChannelMessageSend(m.ChannelID, "DM sent!")
	userChan, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error sending message! I'm broken!")
		return
	}
	s.ChannelMessageSend(userChan.ID, resp)
}
