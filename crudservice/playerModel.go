package main

import (
	"database/sql"
	"errors"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

//Player ...
type Player struct {
	ID          string `json:"id,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	GamesPlayed string `json:"gamesplayed,omitempty"`
	GamesWon    string `json:"gameswon,omitempty"`
	GoalsScored string `json:"goalsscored,omitempty"`
	RankWin     string `json:"rankwin,omitempty"`
	RankScore   string `json:"rankscore,omitempty"`
	LastAvatar  string `json:"lastavatar,omitempty"`
}

//PlayerProfile contains a player struct with its respective apptoken
type PlayerProfile struct {
	Profile  Player
	AppToken string `json:"ApplicationToken,omitempty"`
	Error    string `json:"error-message,omitempty"`
}

//QueryDeletePlayer Clears Player in database
func QueryDeletePlayer(db *sql.DB, p *Player) error {
	result, err := db.Exec(`DELETE FROM Players WHERE ID = ?`, p.ID)
	if err != nil {
		return err
	}
	affected, err2 := result.RowsAffected()
	if err2 != nil {
		return errors.New("Rows Affected")
	}
	if affected != int64(1) {
		return errors.New("Player Not Found")
	}
	return nil
}

//QueryCreatePlayer inserts new Player in database
func QueryCreatePlayer(db *sql.DB, p *Player) error {
	result, err := db.Exec(`INSERT INTO Players (Nickname)
	VALUES (?)`, p.Nickname)
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
		return errors.New("Invalid User ID")
	}

	rows, err := db.Query(`SELECT * FROM Players WHERE ID = ?`, p.ID)
	if err != nil {
		return errors.New("Select Failed" + err.Error())
	}
	defer rows.Close()

	var ID, nickname, lastAvatar string
	var results, gamesplayed, gameswon, goalsscored, rankwin, rankscore int

	for rows.Next() {
		err2 := rows.Scan(&ID, &nickname, &gamesplayed, &gameswon, &goalsscored, &rankwin, &rankscore, &lastAvatar)
		if err2 != nil {
			return errors.New("Scan Rows Failed" + err2.Error())
		}

		results++
		p.Nickname = nickname
		p.GamesPlayed = strconv.Itoa(gamesplayed)
		p.GamesWon = strconv.Itoa(gameswon)
		p.GoalsScored = strconv.Itoa(goalsscored)
		p.RankWin = strconv.Itoa(rankwin)
		p.RankScore = strconv.Itoa(rankscore)
		p.LastAvatar = lastAvatar
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
		return errors.New("Invalid User ID")
	}
	var mods []string //Declaring slice of values to change

	//any value of the struct that is non nil is updated
	if p.Nickname != "" {
		mods = append(mods, "Nickname", p.Nickname)
	}

	if p.GamesPlayed != "" {
		mods = append(mods, "GamesPlayed", p.GamesPlayed)
	}

	if p.GoalsScored != "" {
		mods = append(mods, "GoalsScored", p.GoalsScored)
	}

	if p.GamesWon != "" {
		mods = append(mods, "GamesWon", p.GamesWon)
	}

	if p.LastAvatar != "" {
		mods = append(mods, "LastAvatar", p.LastAvatar)
	}

	//Easily can add more

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

	/*_, rankErr := db.Exec(rankTrigger)
	if rankErr != nil {
		return errors.New("Ranking Error" + rankErr.Error())
	}*/

	return nil
}

//QueryCreateFBData inserts player information obtained from graph API
func QueryCreateFBData(db *sql.DB, u *AppUser) error {
	result, err := db.Exec(`INSERT INTO FacebookData (FacebookID, PlayerID, FullName, Email)
	VALUES (?, ?, ?, ?)`, u.FacebookID, u.ID, u.FullName, u.Email)
	if err != nil {
		return errors.New("Get FB Data ID Error - Execution" + err.Error())
	}
	affected, err2 := result.RowsAffected()
	if err2 != nil {
		return errors.New("Get FB Data ID Error - Result" + err2.Error())
	}
	if affected == int64(0) {
		return errors.New("Create Fail")
	}
	return nil
}

//QueryGetFBDataID Looks in the db if there is an ID corresponding the AppUser's FB ID
func QueryGetFBDataID(db *sql.DB, u *AppUser) error {
	row, err := db.Query("SELECT PlayerID FROM FacebookData WHERE FacebookID = ?", u.FacebookID)
	if err != nil {
		return err
	}
	row.Next()
	err2 := row.Scan(&u.ID)
	if err2 != nil {
		return errors.New(u.FacebookID)
	}
	return nil
}

//QueryRankTrigger trigger for score and win rank
func QueryRankTrigger(db *sql.DB) error {

	var rankTrigger = `
	UPDATE Players AS MainT
	JOIN
	(SELECT @rownum := @rownum+1 as Rank, ID
	 FROM (
			SELECT *
			FROM Players
			ORDER BY Players.GamesWon DESC) AS P2, (
			SELECT @rownum := 0
		   ) r
	) AS Ranked
	ON MainT.ID = Ranked.ID
	SET MainT.RankMostWins = Ranked.Rank`

	var rankTrigger2 = `
	UPDATE Players AS MainT
	JOIN
	(SELECT @rownum := @rownum+1 as Rank, ID
	 FROM (
			SELECT *
			FROM Players
			ORDER BY Players.GoalsScored DESC) AS P2, (
			SELECT @rownum := 0
		  ) r
	) AS Ranked
	ON MainT.ID = Ranked.ID
	SET MainT.RankMostScored = Ranked.Rank`

	_, rankErr := db.Exec(rankTrigger)
	if rankErr != nil {
		return rankErr
	}
	_, rankErr = db.Exec(rankTrigger2)
	if rankErr != nil {
		return rankErr
	}
	return nil
}
