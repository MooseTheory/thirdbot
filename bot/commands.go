package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
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

	sendWhisper(s, m, resp)
}

func status(s *discordgo.Session, m *discordgo.MessageCreate) {
	g, err := s.Guild("218131283505709056")
	if err != nil {
		sendWhisper(s, m, "I broke attempting to get the guild!")
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
	sendWhisper(s, m, resp)
}

func leaders(s *discordgo.Session, m *discordgo.MessageCreate) {
	firsts, err := conn.GetLeaders()
	if err != nil {
		log.Errorln("error getting leaders", err)
		sendChatError(s, m, "error getting leaders")
		return
	}
	resp := "**LEADERS**\n"
	resp += config.Comments.getLeaderHeader() + "\n"

	for i, rec := range firsts {
		user, err := s.User(rec.UserID)
		if err != nil {
			continue
		}
		if i == 0 {
			resp += fmt.Sprintf("%s: %d\n%s\n", user.Username, rec.Count, config.Comments.getLeaderComment())
		} else {
			resp += fmt.Sprintf("%s: %d\n%s\n", user.Username, rec.Count, config.Comments.getFirstComment())
		}
	}
	sendChatMessage(s, m, resp)
}

func last(s *discordgo.Session, m *discordgo.MessageCreate) {
	var resp string
	rec, err := conn.GetLast()
	if err != nil {
		log.Errorln("error getting last third", err)
		sendChatError(s, m, "error getting last third")
		return
	}
	if rec.UserID == "" {
		resp = "Ain't nobody been third yet! Slackers."
	} else {
		user, err := s.User(rec.UserID)
		if err != nil {
			log.Errorln("error getting username", err)
			sendChatError(s, m, "error getting username")
			return
		}
		tz, err := time.LoadLocation("America/New_York")
		if err != nil {
			log.Errorln("error getting timezone", err)
			sendChatError(s, m, "error getting timezone")
			return
		}
		zonedDate := rec.Timestamp.In(tz)
		resp = fmt.Sprintf("The last third was %s on %s, at %s.", user.Username, zonedDate.Format("Jan 02"), zonedDate.Format("3:04PM"))
	}
	sendChatMessage(s, m, resp)
}

func me(s *discordgo.Session, m *discordgo.MessageCreate) {
	userID := m.Author.ID

	numResults, err := conn.GetUserLast(userID)
	if err != nil {
		log.Errorln("error getting count", err)
		sendChatError(s, m, "error getting count")
		return
	}
	user, err := s.User(userID)
	if err != nil {
		log.Errorln("error getting username", err)
		sendChatError(s, m, "error getting username")
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
	sendChatMessage(s, m, resp)
}
