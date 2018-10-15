//
// Author: Sandeep Kalra
// Part of code is adopted from my old code here: https://github.com/sandeepkalra/mv-backend
//
package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

type WebDriver struct {
	DB *DBIf
}

func NewWeb() *WebDriver {
	db, err := NewDB()
	if err != nil {
		panic(err)
	}
	return &WebDriver{DB: db}
}

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

// Message Interface with web!
type Message struct {
	Title     string `json:"title"`
	Msg       string `json:"message"`
	FromEmail string `json:"from_email"`
	Password  string `json:"user_password"`
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

//JsonObjResp is the struct we fill to send resp
type JsonObjResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Body interface{} `json:"body,omitempty"`
}

func NewJsObj() *JsonObjResp {
	return &JsonObjResp{Code: -1, Msg: "Invalid Request"}
}

func (js *JsonObjResp) Send(w http.ResponseWriter) {
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

func (web *WebDriver) LoginUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func (web *WebDriver) UpdateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

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

func (web *WebDriver) PostMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func (web *WebDriver) DeletePost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func (web *WebDriver) ReadPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}

func main() {
	web := NewWeb()
	defer web.DestroyWebDriver()

	r := httprouter.New()
	r.POST("/api/signup", web.SignupUser)
	r.POST("/api/login", web.LoginUser)
	r.POST("/api/reset", web.ResetUser)
	r.POST("/api/update", web.UpdateUser)
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
