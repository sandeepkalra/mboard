package main

const (
	pgConnHost     = "localhost"
	pgConnPort     = 5432
	pgConnUser     = "mbadmin"
	pgConnPassword = "kalra"
	pgConnDBName   = "message_board"
)

//DBInterface defined what DB module must offer for users
type DBInterface interface {
	NewDB() error
	InsertNewUser(email, firstname, lastname string) (userID int64, token string, e error)
	UpdateUserByEmail(email, status, password, firstname, lastname, location, phone, token, preference string) (e error)
	SearchUserByEmail(email string) (id int64, status, firstname, lastname, location, phone string, e error)
	RegenerateUserTokenByEmail(email string) (token string, e error)
	Post(userID int64, title, msg string) (msgID int64, e error)
	Delete(userEmail, title string) (e error)
	List(pageIndex int) (m []Message, e error)
	DestroyDB()
}
