package thirddb

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

type ThirdConn struct {
	username string
	password string
	database string
	net      string
	server   string
	port     int
	addr     string
	db       *sql.DB
}

func newDb(username, password, database, net, server string, port int) (conn ThirdConn, err error) {
	var addr string
	if server != "" {
		if port != 0 {
			addr = fmt.Sprintf("%s:%d", server, port)
		} else {
			addr = server
		}
	} else {
		return ThirdConn{}, errors.New("you must provide a server name to connect to")
	}
	if net == "" {
		net = "tcp"
	}
	return ThirdConn{
		username: username,
		password: password,
		database: database,
		net:      net,
		server:   server,
		port:     port,
	}, nil
}

func (tc *ThirdConn) connect() (err error) {
	conf := mysql.Config{
		User:                 tc.username,
		Passwd:               tc.password,
		DBName:               tc.database,
		Net:                  tc.net,
		Addr:                 tc.addr,
		AllowNativePasswords: true,
		ParseTime:            true,
	}
	tc.db, err = sql.Open("mysql", conf.FormatDSN())
	return err
}

func (tc *ThirdConn) checkIfThird(startTime time.Time) (err error) {
	stmt, err := tc.db.Prepare("SELECT COUNT(*) FROM thirds WHERE date BETWEEN ? AND ?")
	if err != nil {
		return err
	}
	now := time.Now()
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		return errors.New("error getting time zone")
	}
	now = now.In(tz)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz)
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
