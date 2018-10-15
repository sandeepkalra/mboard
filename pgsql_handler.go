package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"time"
)

type DBIf struct {
	db *sql.DB
}

//InitDB checks psql and connects to it
func NewDB() (*DBIf, error) {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		pgConnHost, pgConnPort, pgConnUser, pgConnPassword, pgConnDBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		db.Close()
		return nil, err
	}
	fmt.Println("PGSql:Connection Created")
	dbif := &DBIf{db: db}
	return dbif, nil
}

func (dbIf *DBIf) DestroyDB() {
	if dbIf != nil && dbIf.db != nil {
		dbIf.db.Close()
	}
}

//InsertNewUser is when new user signup
func (dbIf *DBIf) InsertNewUser(email, firstname, lastname string) (user_id int64, token string, e error) {
	if dbIf == nil {
		e = fmt.Errorf("Invalid pointer to DB")
		return
	}
	var id int64
	e = dbIf.db.QueryRow("select user_id,firstname,lastname from users where email=$1", email).Scan(&id)
	if e == nil {
		e = fmt.Errorf("user exists")
		return
	}
	token_uuid := uuid.Must(uuid.NewV4())
	token = token_uuid.String()
	e = dbIf.db.QueryRow("insert into users (firstname, lastname, email,  status, one_time_token) "+
		" values ($1, $2, $3, $4, $5) returning user_id",
		firstname, lastname, email, "blocked", token_uuid.String()).Scan(&id)
	if e != nil {
		return
	}
	return id, token_uuid.String(), nil
}

//SearchUserByEmail search user by email.

func (dbIf *DBIf) SearchUserByEmail(email string) (id int64, token, prefs, password, status, firstname, lastname, location, phone string, e error) {
	var pgtoken, pgprefs, pgpassword, pgstatus, pgfirstname, pglastname, pglocation, pgphone sql.NullString

	if dbIf == nil {
		e = fmt.Errorf("Invalid pointer to DB")
		return
	}

	e = dbIf.db.QueryRow("select user_id, firstname, lastname,preferences, password, one_time_token, status,  location, phone from users where email=$1",
		email).Scan(&id, &pgfirstname, &pglastname, &pgprefs, &pgpassword, &pgtoken, &pgstatus, &pglocation, &pgphone)

	token, prefs, password, status, firstname, lastname, location, phone = pgtoken.String, pgprefs.String, pgpassword.String, pgstatus.String, pgfirstname.String, pglastname.String, pglocation.String, pgphone.String

	return
}

//UpdateUserByEmail updates user by email.
func (dbIf *DBIf) UpdateUserByEmail(email, status, password, firstname, lastname, location, phone, token, preferences string) (e error) {
	if dbIf == nil {
		e = fmt.Errorf("Invalid pointer to DB")
		return
	}

	_, e = dbIf.db.Exec("update users set status=$1, firstname=$2, lastname=$3, location =$4,"+
		" phone=$5, password=$6, preferences=$7, one_time_token = $8 where email = $9 ",
		status, firstname, lastname, location, phone, password, preferences, token, email)
	return
}

// RegenerateUserTokenByEmail generated a token for user, and blocks his profile till he confirms the token.
func (dbIf *DBIf) RegenerateUserTokenByEmail(email string) (token string, e error) {
	if dbIf == nil {
		e = fmt.Errorf("Invalid pointer to DB")
		return
	}
	token_uuid := uuid.Must(uuid.NewV4())
	token = token_uuid.String()
	_, e = dbIf.db.Exec("update users set status=$1, one_time_token=$2 where email = $3",
		"blocked", token, email)
	return token_uuid.String(), e
}

func (dbIf *DBIf) Post(userID int64, title, msg string) (msgID int64, e error) {
	if dbIf == nil {
		e = fmt.Errorf("Invalid pointer to DB")
		return
	}
	t := time.Now()
	e = dbIf.db.QueryRow("insert into messages (created_by_user, title, message, created_on_date) "+
		" values ($1, $2, $3, $4, $5) returning message_id",
		userID, title, msg, t.Format("01-01-2006")).Scan(&msgID)
	return
}

//Delete the message; only by user who created it;
func (dbIf *DBIf) Delete(userEmail, title string) (e error) {
	if dbIf == nil {
		e = fmt.Errorf("Invalid pointer to DB")
		return
	}
	_, e = dbIf.db.Exec("delete from messages where created_by_user=$1 and title = $2", userEmail, title)
	return
}

// List messages in chunck of 50 !
func (dbIf *DBIf) List(pageIndex int) (m []Message, e error) {
	if dbIf == nil {
		e = fmt.Errorf("Invalid pointer to DB")
		return
	}
	if pageIndex < 0 {
		pageIndex = 0
	}
	rows, _ := dbIf.db.Query("select created_by_user, title, messages from messages limit 50 offset $1", pageIndex)
	for rows.Next() {
		var msg Message
		e = rows.Scan(&msg.FromEmail, &msg.Title, &msg.Msg)
		if e == nil {
			m = append(m, msg)
		} else {
			break
		}
	}
	rows.Close()
	return
}
