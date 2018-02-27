package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

//Player ...
type Player struct {
	ID          string `json:"id,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	GamesPlayed string `json:"gamesplayed,omitempty"`
	GoalsScored string `json:"goalsscored,omitempty"`
}

//QueryDeletePlayer Clears Player in database
func QueryDeletePlayer(db *sql.DB, p *Player) error {
	request := fmt.Sprintf(`DELETE FROM Players WHERE ID = '%s'`, p.ID)
	result, err := db.Exec(request)
	if err != nil {
		return errors.New("Query Error")
	}
	affected, err2 := result.RowsAffected()
	if err2 != nil {
		return errors.New("Resulting Rows Error")
	}
	if affected != int64(1) {
		return errors.New("None Found")
	}
	return nil
}

//QueryCreatePlayer inserts new Player in database
func QueryCreatePlayer(db *sql.DB, p *Player) error {
	request := fmt.Sprintf(`INSERT INTO Players (Nickname, GamesPlayed, GamesWon, GoalsScored, Active)
	VALUES ('%s', '0', '0', '0', '0')`, p.Nickname)
	result, err := db.Exec(request)
	if err != nil {
		return errors.New("Query Error")
	}
	affected, err2 := result.RowsAffected()
	if err2 != nil {
		return errors.New("Create Failed fam")
	}
	if affected == int64(1) {
		db.QueryRow("SELECT ID FROM Players WHERE Nickname = ?", p.Nickname).Scan(&p.ID)
		p.GamesPlayed = "0"
		p.GoalsScored = "0"
		return nil
	}
	return errors.New("Abnormal number of creates")
}

//QueryAllPlayers Returns all the Players stored in the Players table
func QueryAllPlayers(db *sql.DB) error {
	return errors.New("Not ready")
	/*var rows, err = db.Query("SELECT * FROM Players")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var ID string
	var Nickname string
	var results, a, b, c, d int
	for rows.Next() {
		err := rows.Scan(&ID, &Nickname, &a, &b, &c, &d)
		if err != nil {
			fmt.Println(err)
		}
		results++
		fmt.Println("ID = ", ID, "Nickname = ", Nickname)
	}
	fmt.Println("Players in total:", results)*/
}

//QuerySearchPlayer Looks for Player in database
func QuerySearchPlayer(db *sql.DB, p *Player) error {
	if p.ID == "" {
		return errors.New("Empty")
	}
	request := fmt.Sprintf("SELECT * FROM Players WHERE ID = '%s'", p.ID)
	rows, err := db.Query(request)
	if err != nil {
		return errors.New("Query Error")
	}
	defer rows.Close()

	var ID, Nickname string
	var results, a, b, c, d int

	for rows.Next() {
		err := rows.Scan(&ID, &Nickname, &a, &b, &c, &d)
		if err != nil {
			fmt.Println(err)
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
func QueryUpdatePlayer(db *sql.DB, p *Player) error {
	if p.ID == "" {
		return errors.New("Empty ID")
	}
	request := fmt.Sprintf("SELECT * FROM Players WHERE ID = '%s'", p.ID)
	rows, errReq := db.Exec(request) //Initially look if the user exists in the db
	found, errRow := rows.RowsAffected()
	if errReq != nil || errRow != nil || found != 1 {
		return sql.ErrNoRows
	}

	//Prepare an update statement based on the information we have
	updateMask, prepErr := db.Prepare("UPDATE Players SET ? = ? WHERE ID = ?")
	if prepErr != nil {
		return errors.New("Statement Error")
	}

	var fieldsUpdated int64
	var rowaff int64
	var aff sql.Result
	var execErr error
	if p.Nickname != "" {
		aff, execErr = updateMask.Exec("Nickname", p.Nickname, p.ID)
		rowaff, execErr = aff.RowsAffected()
		fieldsUpdated += rowaff
		rowaff = 0
	}

	if p.GamesPlayed != "" {
		aff, execErr = updateMask.Exec("GamesPlayed", p.GamesPlayed, p.ID)
		rowaff, execErr = aff.RowsAffected()
		fieldsUpdated += rowaff
		rowaff = 0
	}

	if p.GoalsScored != "" {
		aff, execErr = updateMask.Exec("Nickname", p.GoalsScored, p.ID)
		rowaff, execErr = aff.RowsAffected()
		fieldsUpdated += rowaff
		rowaff = 0
	}

	if execErr != nil {
		return execErr
	}
	p.ID = strconv.Itoa(int(fieldsUpdated)) //Returning in ID number of fields affected

	return nil
}
