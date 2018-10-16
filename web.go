//
// Author: Sandeep Kalra
// Part of code is adopted from my old code here: https://github.com/sandeepkalra/mv-backend
//
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// Message Interface with web!
type Message struct {
	Title    string `json:"title"`
	Msg      string `json:"message"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User Interface with web!
type User struct {
	Email        string `json:"email"`
	FirstName    string `json:"firstname,omitempty"`
	LastName     string `json:"lastname,omitempty"`
	Password     string `json:"password,omitempty"`
	Status       string `json:"status,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Location     string `json:"location,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Preferences  string `json:"preferences,omitempty"`
	OneTimeToken string `json:"token,omitempty"`
}

//JSONObjResp is the struct we fill to send resp
type JSONObjResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Body interface{} `json:"body,omitempty"`
}

//WebDriver single main web struct
type WebDriver struct {
	DB *DBIf
}

//NewWeb new driver
func NewWeb() *WebDriver {
	db, err := NewDB()
	if err != nil {
		panic(err)
	}
	return &WebDriver{DB: db}
}

//DestroyWebDriver destroys all
func (web *WebDriver) DestroyWebDriver() {
	if web != nil && web.DB != nil {
		web.DB.DestroyDB()
	}
}

//GetCryptPassword creates bcrypt password hash
func GetCryptPassword(password string) string {
	pasBytes, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if e != nil {
		//TODO: Properly handle error
		log.Fatal(e)
	}
	return string(pasBytes)
}

//CheckPasswordHashes check if given password belongs to same bcrypt hash or not
func CheckPasswordHashes(userGivenPassword, userDBPassword string) (bool, error) {
	if e := bcrypt.CompareHashAndPassword([]byte(userDBPassword), []byte(userGivenPassword)); e != nil {
		return false, e
	}
	return true, nil
}

//NewJsObj returns rew JSON response object
func NewJsObj() *JSONObjResp {
	return &JSONObjResp{Code: -1, Msg: "Invalid Request"}
}

//Send sends the JSON data back
func (js *JSONObjResp) Send(w http.ResponseWriter) {
	if js == nil {
		fmt.Println("Js obj is still nil ")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	data, err := json.Marshal(*js)
	if err != nil {
		fmt.Println("error marshalling the results", data)
	} else {
		fmt.Fprintf(w, string(data))
	}
	return
}

//SignupUser sign-up new user
func (web *WebDriver) SignupUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := User{}
	js := NewJsObj()
	if js == nil {
		fmt.Fprintf(w, "something went wrong with server")
		return
	}
	defer js.Send(w)

	if e := json.NewDecoder(r.Body).Decode(&user); e != nil {
		js.Msg = " failed to decode incoming msg "
		return
	}

	if user.Email == "" || user.LastName == "" || user.FirstName == "" {
		js.Msg = "mandatory field missing: need firstname, lastname and email"
		js.Body = map[string]interface{}{
			"got": user,
		}
		return
	}

	id, token, err := web.DB.InsertNewUser(user.Email, user.FirstName, user.LastName)
	if err != nil {
		js.Msg = err.Error()
		js.Body = map[string]interface{}{
			"got": user,
		}
		return
	}

	js.Code = 0
	js.Msg = "ok"
	js.Body = map[string]interface{}{
		"id":    id,
		"token": token,
	}
}

//LoginUser login the user
func (web *WebDriver) LoginUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := User{}
	js := NewJsObj()
	if js == nil {
		fmt.Fprintf(w, "something went wrong with server")
		return
	}
	defer js.Send(w)

	if e := json.NewDecoder(r.Body).Decode(&user); e != nil {
		js.Msg = " failed to decode incoming msg "
		return
	}

	if user.Email == "" || user.Password == "" {
		js.Msg = "mandatory field missing: need email, password"
		js.Body = map[string]interface{}{
			"got": user,
		}
		return
	}

	_, _, _, password, _, _, _, _, _, e := web.DB.SearchUserByEmail(user.Email)
	if e != nil {
		js.Msg = "user | password mismatch or does not exist"
		js.Body = map[string]interface{}{
			"got": user,
		}
		return
	}

	if b, _ := CheckPasswordHashes(user.Password, password); b != true {
		js.Msg = "user | password mismatch or does not exist"
		return
	}

	js.Code = 0
	js.Msg = "ok"
}

//ResetUser sets the user to blocked state till he goes back and unblocks itself with new info.
func (web *WebDriver) ResetUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := User{}
	js := NewJsObj()
	if js == nil {
		fmt.Fprintf(w, "something went wrong with server")
		return
	}
	defer js.Send(w)

	if e := json.NewDecoder(r.Body).Decode(&user); e != nil {
		js.Msg = " failed to decode incoming msg "
		return
	}

	if user.Email == "" || user.OneTimeToken == "" {
		js.Msg = "mandatory field missing: need token and email"
		js.Body = map[string]interface{}{
			"got00": user,
		}
		return
	}

	_, _, preferences, password, _, firstname, lastname, location, phone, e := web.DB.SearchUserByEmail(user.Email)
	if e != nil {
		js.Msg = "error:" + e.Error()
		js.Body = map[string]interface{}{
			"got0": user,
		}
		return
	}

	user.Status = "blocked"
	user.OneTimeToken = uuid.Must(uuid.NewV4()).String()
	user.FirstName = firstname
	user.LastName = lastname
	user.Password = password
	user.Location = location
	user.Phone = phone
	user.Preferences = preferences
	err := web.DB.UpdateUserByEmail(user.Email, user.Status, user.Password, user.FirstName, user.LastName, user.Location, user.Phone, user.OneTimeToken, user.Preferences)

	if err != nil {
		js.Msg = err.Error()
		js.Body = map[string]interface{}{
			"got2": user,
		}
		return
	}

	js.Code = 0
	js.Msg = "ok"
	js.Body = map[string]interface{}{
		"email": user.Email,
		"token": user.OneTimeToken,
	}

}

//UpdateUser user wants to update its profile
func (web *WebDriver) UpdateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := User{}
	js := NewJsObj()
	if js == nil {
		fmt.Fprintf(w, "something went wrong with server")
		return
	}
	defer js.Send(w)

	if e := json.NewDecoder(r.Body).Decode(&user); e != nil {
		js.Msg = " failed to decode incoming msg "
		return
	}

	if user.Email == "" || user.OneTimeToken == "" {
		js.Msg = "mandatory field missing: need token and email"
		js.Body = map[string]interface{}{
			"got00": user,
		}
		return
	}

	id, token, preferences, password, status, firstname, lastname, location, phone, e := web.DB.SearchUserByEmail(user.Email)
	if e != nil {
		js.Msg = "error:" + e.Error()
		js.Body = map[string]interface{}{
			"got0": user,
		}
		return
	}
	if token != user.OneTimeToken {
		js.Msg = "token mismatch"
		js.Body = map[string]interface{}{
			"got1": user,
		}
	}

	user.OneTimeToken = ""
	// copy over null values with values from DB
	if status == "blocked" {
		user.Status = "active"
	} else {
		user.Status = status
	}

	if user.FirstName == "" && firstname != "" {
		user.FirstName = firstname
	}

	if user.LastName == "" && lastname != "" {
		user.LastName = lastname
	}

	if user.Password != "" {
		user.Password = GetCryptPassword(user.Password)
	} else {
		user.Password = password
	}

	if user.Location == "" {
		user.Location = location
	}

	if user.Phone == "" {
		user.Phone = phone
	}

	if user.Preferences == "" {
		user.Preferences = preferences
	}

	err := web.DB.UpdateUserByEmail(user.Email, user.Status, user.Password, user.FirstName, user.LastName, user.Location, user.Phone, user.OneTimeToken, user.Preferences)
	if err != nil {
		js.Msg = err.Error()
		js.Body = map[string]interface{}{
			"got2": user,
		}
		return
	}

	js.Code = 0
	js.Msg = "ok"
	js.Body = map[string]interface{}{
		"id":    id,
		"token": token,
	}
}

//PostMessage user wants to post a message
func (web *WebDriver) PostMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	message := Message{}
	js := NewJsObj()
	if js == nil {
		fmt.Fprintf(w, "something went wrong with server")
		return
	}
	defer js.Send(w)

	if e := json.NewDecoder(r.Body).Decode(&message); e != nil {
		js.Msg = " failed to decode incoming msg "
		return
	}

	if message.Email == "" || message.Password == "" {
		js.Msg = "mandatory field missing: need email, password"
		js.Body = map[string]interface{}{
			"got": message,
		}
		return
	}

	_, _, _, password, _, _, _, _, _, e := web.DB.SearchUserByEmail(message.Email)
	if e != nil {
		js.Msg = "user | password mismatch or does not exist"
		js.Body = map[string]interface{}{
			"got": message,
		}
		return
	}

	if b, _ := CheckPasswordHashes(message.Password, password); b != true {
		js.Msg = "user | password mismatch or does not exist"
		return
	}

	messageID, e := web.DB.Post(message.Email, message.Title, message.Msg)
	if e != nil {
		js.Msg = e.Error()
		return
	}

	js.Code = 0
	js.Msg = "ok"
	js.Body = map[string]interface{}{
		"message_id": messageID,
		"title":      message.Title,
	}
}

//DeletePost user wants to delete his/her post
func (web *WebDriver) DeletePost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	message := Message{}
	js := NewJsObj()
	if js == nil {
		fmt.Fprintf(w, "something went wrong with server")
		return
	}
	defer js.Send(w)

	if e := json.NewDecoder(r.Body).Decode(&message); e != nil {
		js.Msg = " failed to decode incoming msg "
		return
	}

	if message.Email == "" || message.Password == "" {
		js.Msg = "mandatory field missing: need email, password"
		js.Body = map[string]interface{}{
			"got": message,
		}
		return
	}

	_, _, _, password, _, _, _, _, _, e := web.DB.SearchUserByEmail(message.Email)
	if e != nil {
		js.Msg = "user | password mismatch or does not exist"
		js.Body = map[string]interface{}{
			"got": message,
		}
		return
	}

	if b, _ := CheckPasswordHashes(message.Password, password); b != true {
		js.Msg = "user | password mismatch or does not exist"
		return
	}

	e = web.DB.Delete(message.Email, message.Title)
	if e != nil {
		js.Msg = e.Error()
		return
	}

	js.Code = 0
	js.Msg = "ok"
}

//ReadPost user(anonymous) can read post
func (web *WebDriver) ReadPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	message := Message{}
	js := NewJsObj()
	if js == nil {
		fmt.Fprintf(w, "something went wrong with server")
		return
	}
	defer js.Send(w)

	if e := json.NewDecoder(r.Body).Decode(&message); e != nil {
		js.Msg = " failed to decode incoming msg "
		return
	}

	if message.Email == "" || message.Password == "" {
		js.Msg = "mandatory field missing: need email, password"
		js.Body = map[string]interface{}{
			"got": message,
		}
		return
	}

	_, _, _, password, _, _, _, _, _, e := web.DB.SearchUserByEmail(message.Email)
	if e != nil {
		js.Msg = "user | password mismatch or does not exist"
		js.Body = map[string]interface{}{
			"got": message,
		}
		return
	}

	if b, _ := CheckPasswordHashes(message.Password, password); b != true {
		js.Msg = "user | password mismatch or does not exist"
		return
	}

	m, e := web.DB.List(0)
	if e != nil {
		js.Msg = e.Error()
		return
	}

	js.Code = 0
	js.Msg = "ok"
	js.Body = map[string]interface{}{
		"messages": m,
	}
}

//main get the ball rolling
func main() {
	web := NewWeb()
	defer web.DestroyWebDriver()

	r := httprouter.New()
	r.POST("/api/signup", web.SignupUser) // done
	r.POST("/api/login", web.LoginUser)   // done
	r.POST("/api/reset", web.ResetUser)   // done
	r.POST("/api/update", web.UpdateUser) // done
	r.POST("/api/post", web.PostMessage)
	r.POST("/api/read", web.ReadPost)
	r.POST("/api/delete", web.DeletePost)
	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8080",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	fmt.Println("listening on :8080")
	log.Fatal(srv.ListenAndServe())
}
