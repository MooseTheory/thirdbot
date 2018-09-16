package thirddb

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

// ThirdConn is a database connection for thirdbot
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

// FirstRecord holds the information about a first.
type FirstRecord struct {
	UserID    string
	Count     int
	Timestamp time.Time
}

// NewConn returns a new instance of ThirdConn. This connection should already
// be connected, unless there was an error.
func NewConn(username, password, database, net, server string, port int) (conn ThirdConn, err error) {
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
	tc := ThirdConn{
		username: username,
		password: password,
		database: database,
		net:      net,
		server:   server,
		port:     port,
		addr:     addr,
	}
	err = tc.connect()
	if err != nil {
		return ThirdConn{}, err
	}
	err = tc.prepareDatabase()
	if err != nil {
		return ThirdConn{}, err
	}
	return tc, nil
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

// driver.ErrBadConn?

func (tc *ThirdConn) prepareQuery(query string, retries int) (stmt *sql.Stmt, err error) {
	for i := 0; i < retries; i++ {
		stmt, err = tc.db.Prepare(query)
		if err == nil {
			return stmt, err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, err
}

// CheckIfThird determins if the third is available for the day
func (tc *ThirdConn) CheckIfThird(startTime time.Time) (isThird bool, err error) {
	query := "SELECT COUNT(*) FROM thirds WHERE date BETWEEN ? AND ?"
	stmt, err := tc.prepareQuery(query, 5)
	if err != nil {
		return false, err
	}

	now := time.Now()
	if now.Before(startTime) {
		// We haven't reached the offset yet.
		return false, nil
	}
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		return false, err
	}
	now = now.In(tz)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz)
	tomorrow := midnight.Add(time.Hour * 24)
	numResults := 0
	err = stmt.QueryRow(midnight, tomorrow).Scan(&numResults)
	if err != nil {
		return false, err
	}
	if numResults == 0 {
		return true, nil
	}
	return false, nil
}

// AddThird adds a third to the database
func (tc *ThirdConn) AddThird(authorID string) (err error) {
	query := "INSERT INTO thirds(userid, date) VALUES(?, ?)"
	stmt, err := tc.prepareQuery(query, 5)
	if err != nil {
		return err
	}
	now := time.Now()
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		return err
	}
	zonedDate := now.In(tz)
	_, err = stmt.Exec(authorID, zonedDate)
	if err != nil {
		return err
	}
	return nil
}

// GetLeaders returns an array of FirstRecords containing the tracked firsts.
func (tc *ThirdConn) GetLeaders() (firsts []FirstRecord, err error) {
	query := "SELECT COUNT(*) AS `count`, userid FROM thirds GROUP BY userid ORDER BY count DESC"
	stmt, err := tc.prepareQuery(query, 5)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var count int
		var userID string
		err = rows.Scan(&count, &userID)
		if err != nil {
			continue
		}
		rec := FirstRecord{
			UserID: userID,
			Count:  count,
		}
		firsts = append(firsts, rec)
	}
	return firsts, nil
}

// GetLast returns the last first awarded.
func (tc *ThirdConn) GetLast() (rec FirstRecord, err error) {
	query := "SELECT `userid`, `date` FROM `thirds` ORDER BY `date` DESC LIMIT 1"
	stmt, err := tc.prepareQuery(query, 5)
	if err != nil {
		return FirstRecord{}, err
	}
	rows, err := stmt.Query()
	if err != nil {
		return FirstRecord{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var date time.Time
		err = rows.Scan(&userID, &date)
		if err != nil {
			continue
		}
		rec = FirstRecord{
			UserID:    userID,
			Timestamp: date,
		}
	}
	return rec, nil
}

// GetUserLast returns the last third gotten by a specific user.
func (tc *ThirdConn) GetUserLast(userID string) (numFirsts int, err error) {
	query := "SELECT COUNT(*) FROM `thirds` WHERE `userid` = ?"
	stmt, err := tc.prepareQuery(query, 5)
	if err != nil {
		return -1, err
	}
	err = stmt.QueryRow(userID).Scan(&numFirsts)
	return numFirsts, err
}

func (tc *ThirdConn) prepareDatabase() (err error) {
	const createThirdsStatement = "CREATE TABLE IF NOT EXISTS`thirds`(" +
		"`id` int NOT NULL AUTO_INCREMENT," +
		"`userid` varchar(64) NOT NULL," +
		"`date` datetime NOT NULL," +
		"PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Tracks Thirds';"

	_, err = tc.db.Exec(createThirdsStatement)
	if err != nil {
		return err
	}
	return
}
