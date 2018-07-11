package main

import (
	"math/rand"
)

type CommentInfo struct {
	GotThird       []string
	LeaderHeader   []string
	FirstComments  []string
	LeaderComments []string
}

func (ci *CommentInfo) getThirdComment() string {
	i := rand.Intn(len(ci.GotThird))
	return ci.GotThird[i]
}

func (ci *CommentInfo) getLeaderHeader() string {
	i := rand.Intn(len(ci.LeaderHeader))
	return ci.LeaderHeader[i]
}

func (ci *CommentInfo) getFirstComment() string {
	i := rand.Intn(len(ci.FirstComments))
	return ci.FirstComments[i]
}

func (ci *CommentInfo) getLeaderComment() string {
	i := rand.Intn(len(ci.LeaderComments))
	return ci.LeaderComments[i]
}
