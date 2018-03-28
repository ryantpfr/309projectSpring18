package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//Player ...
type Player struct {
	ID          string `json:"id,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	GamesPlayed string `json:"gamesplayed,omitempty"`
	GoalsScored string `json:"goalsscored,omitempty"`
}

//PlayerProfile contains a player struct with its respective apptoken
type PlayerProfile struct {
	Profile  Player
	AppToken string `json:"ApplicationToken,omitempty"`
	Error    string `json:"error-message,omitempty"`
}

//QueryDeletePlayer Clears Player in database
func QueryDeletePlayer(db *sql.DB, p *Player) error {
	request := fmt.Sprintf(`DELETE FROM Players WHERE ID = '%s'`, p.ID)
	result, err := db.Exec(request)
	if err != nil {
		return err
	}
	affected, err2 := result.RowsAffected()
	if err2 != nil {
		return err2
	}
	if affected != int64(1) {
		return errors.New("Player Not Found")
	}
	return nil
}

//QueryCreatePlayer inserts new Player in database
func QueryCreatePlayer(db *sql.DB, p *Player) error {
	request := fmt.Sprintf(`INSERT INTO Players (Nickname, GamesPlayed, GamesWon, GoalsScored, Active)
	VALUES ('%s', '0', '0', '0', '0')`, p.Nickname)
	result, err := db.Exec(request)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == int64(0) {
		return errors.New("Create Fail")
	}
	db.QueryRow("SELECT ID FROM Players WHERE Nickname = ?", p.Nickname).Scan(&p.ID)
	p.GamesPlayed = "0"
	p.GoalsScored = "0"
	return nil
}

//QuerySearchPlayer Looks for Player in database
func QuerySearchPlayer(db *sql.DB, p *Player) error {
	if p.ID == "" {
		return errors.New("Invalid user ID")
	}
	request := fmt.Sprintf("SELECT * FROM Players WHERE ID = '%s'", p.ID)
	rows, err := db.Query(request)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ID, Nickname string
	var results, a, b, c, d int

	for rows.Next() {
		err := rows.Scan(&ID, &Nickname, &a, &b, &c, &d)
		if err != nil {
			return sql.ErrNoRows
		}
		results++
		p.Nickname = Nickname
		p.GamesPlayed = string(a)
		p.GoalsScored = string(c)
	}
	if results == 0 { //Diego from the future, you idiot, dont move this from here
		return sql.ErrNoRows
	}

	return nil
}

//QueryUpdatePlayer Searchs for a matching ID and updates based on player values given
//Returns errors in case of not finding the correct ID or getting a wrong value
//MODIFIES Player object to overrwrite
func QueryUpdatePlayer(db *sql.DB, p *Player) error {
	if p.ID == "" {
		return errors.New("Invalid user ID")
	}
	var mods []string //Declaring slie of values to change

	//any value of the struct that is non nil is updated
	if p.Nickname != "" {
		mods = append(mods, "Nickname", p.Nickname)
	}

	if p.GamesPlayed != "" {
		mods = append(mods, "GamesPlayed", p.GamesPlayed)
	}

	if p.GoalsScored != "" {
		mods = append(mods, "GoalsScored", p.GoalsScored)
	} //Easily can add more

	mods = append(mods, p.ID) //Appending ID which is gonna be modified

	effect, execErr := db.Exec(prepUpdate(mods))
	if execErr != nil {
		return execErr
	}
	i, err := effect.RowsAffected()
	if err != nil {
		return err
	}
	if int(i) == 0 {
		return errors.New("Not Modified")
	}

	return nil
}

//QueryCreateFBData inserts player information obtained from graph API
func QueryCreateFBData(db *sql.DB, u *AppUser) error {
	request := fmt.Sprintf(`INSERT INTO FacebookData (FacebookID, PlayerID, FullName, Email)
	VALUES ('%s', '%s', '%s', '%s')`, u.FacebookID, u.ID, u.FullName, u.Email)
	result, err := db.Exec(request)
	if err != nil {
		return err
	}
	affected, err2 := result.RowsAffected()
	if err2 != nil {
		return err2
	}
	if affected == int64(0) {
		return errors.New("Create Fail")
	}
	return nil
}

//QueryGetFBDataID Looks in the db if there is an ID corresponding the AppUser's FB ID
func QueryGetFBDataID(db *sql.DB, u *AppUser) error {
	row, err := db.Query("SELECT PlayerID FROM FacebookData WHERE FacebookID = ?", u.FacebookID)
	row.Scan(&u.ID)
	if strings.Compare(u.ID, "") == 0 || err != nil {
		return errors.New("Invalid user ID")
	}
	return nil
}

//QuerySetToken creates a new entry on the applicationToken table
//giving the set ID a corresponding appToken and assigning an expiration
func QuerySetToken(db *sql.DB, ID string, appToken string, tokenLife int) error {
	request := fmt.Sprintf(`INSERT INTO TokenTable (applicationToken, playerID, expiration)
	VALUES ('%s', '%s', '%d')`, appToken, ID, getExpiration(tokenLife))
	result, err := db.Exec(request)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected != int64(1) {
		return errors.New("Create Fail")
	}
	return nil
}

//QueryGetUpdateToken query to return the token assigned to a specific ID
func QueryGetUpdateToken(db *sql.DB, ID string) (string, error) {
	res, err := db.Exec(`UPDATE TokenTable SET expiration = ? WHERE playerID = ?`, getExpiration(2), ID)
	if err != nil {
		return "", err
	}
	affected, _ := res.RowsAffected()
	if int(affected) == 0 {
		return "", sql.ErrNoRows
	}

	row, err := db.Query("SELECT applicationToken FROM TokenTable WHERE playerID = ?",
		ID)
	if err != nil {
		return "", err
	}
	var tok string
	row.Next()
	err = row.Scan(&tok)
	if err != nil {
		return "", errors.New("Player Not Found")
	}
	return tok, nil
}

//QueryGetToken query to return the token assigned to a specific ID
func QueryGetToken(db *sql.DB, ID string) (string, error) {
	row, err := db.Query("SELECT applicationToken, expiration FROM TokenTable WHERE playerID = ?", ID)
	if err != nil {
		return "", err
	}
	var tok string
	var exp int64
	row.Next()
	err = row.Scan(&tok, &exp)
	if err != nil {
		return "", errors.New("Player Not Found")
	}
	if exp < time.Now().Unix() {
		return "", errors.New("Application Token Expired")
	}
	return tok, nil
}

//QueryAssertToken returns the nickname of the given apptoken or 404
func QueryAssertToken(db *sql.DB, AppToken string) (string, error) {
	row, err := db.Query(`SELECT Nickname, expiration FROM TokenTable 
	JOIN Players ON TokenTable.playerID = Players.ID WHERE applicationToken = ?`, AppToken)
	if err != nil {
		return "", err
	}
	var Nickname string
	var exp int64
	row.Next()
	err = row.Scan(&Nickname, &exp)
	if err != nil {
		return "", errors.New("Player Not Found")
	}
	if exp < time.Now().Unix() {
		return "", errors.New("Application Token Expired")
	}
	return Nickname, nil
}
