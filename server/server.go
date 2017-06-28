package server

import (
	"net/http"
	"log"
	"fmt"
	"github.com/gomodels"
	"encoding/json"
	"../db"
	"../controllers"
	"github.com/bmizerany/pat"
	"time"
	"gopkg.in/mgo.v2/bson"
	"github.com/rs/cors"
)

type ServerProperties struct {
	Port    string
	Address string
}

var mc = controllers.NewMessageController()

func validAuthHeader(req *http.Request) (bool,models.User) {
	auth := req.Header.Get("Authorization")
	var user models.User
	if len(auth) <= 6 {
		return false,user
	}
	user.Token = auth[6:]
	if mc.Validate(&user){
		return true, user
	}else{
		return false, user
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
}

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(json)
}

func SendMessage(w http.ResponseWriter, req *http.Request) {
	validou,user := validAuthHeader(req)
	if !validou{
		unauthorized(w)
		return
	}
	var message models.Message
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&message)
	if err != nil || len(message.Id.Hex()) > 0 {
		badRequest(w, err)
		return
	}
	id := req.URL.Query().Get(":id")
	message.Room = bson.ObjectIdHex(id)
	message.CreatedIn = time.Now()
	message.CreatedBy = user.Id
	controllers.PublishChan <- message
	err = message.Persist(db.GetCollection())
	if err != nil {
		badRequest(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", req.URL.Path)
	w.WriteHeader(http.StatusCreated)
}

func GetMessages(w http.ResponseWriter, req *http.Request) {
	validou,_ := validAuthHeader(req)
	if !validou{
		unauthorized(w)
		return
	}
	id := req.URL.Query().Get(":id")
	var messages models.Messages
	messages, err := messages.FindAll(db.GetCollection(),bson.ObjectIdHex(id))
	if err != nil {
		badRequest(w, err)
		return
	}
	resp, _ := json.Marshal(messages)
	ResponseWithJSON(w, resp, http.StatusOK)
}

func badRequest(w http.ResponseWriter, err error) {
	log.Println(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
}

func Start(properties ServerProperties) {
	startDb()
	m := pat.New()
	handler := cors.AllowAll().Handler(m)
	mapEndpoints(*m, properties)
	http.Handle("/", handler)
	fmt.Println("servidor iniciado no endereço localhost:" + properties.Port + properties.Address)
	err := http.ListenAndServe(":"+properties.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func mapEndpoints(m pat.PatternServeMux, properties ServerProperties) {
	m.Post(properties.Address+"/:id/messages", http.HandlerFunc(SendMessage))
	m.Get(properties.Address+"/:id/messages", http.HandlerFunc(GetMessages))
}

func startDb() {
	db.Start()
}
