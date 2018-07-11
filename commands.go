package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func runCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string) {
	switch strings.ToLower(command) {
	case "help":
		help(s, m)
	case "leaders":
		leaders(s, m)
	}
}

func help(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp := "**#leaders**\n"
	resp += "Returns the rather unimportant list of first leaders.\n"

	sendMessage(s, m, resp)
}

func leaders(s *discordgo.Session, m *discordgo.MessageCreate) {
	stmt, err := db.Prepare("SELECT (SELECT COUNT(DISTINCT userid) FROM thirds) AS count, userid FROM thirds ORDER BY count")
	if err != nil {
		sendMessage(s, m, "I broke trying to do this!")
	}
	rows, err := stmt.Query()
	if err != nil {
		sendMessage(s, m, "I broke trying to do this!")
	}
	defer rows.Close()
	resp := "**LEADERS**\n"
	resp += "All the thirds!\n"
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
		resp += fmt.Sprintf("%s: %d\n%s\n", user.Username, count, "You're the thirdest!")
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
