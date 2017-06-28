package controllers

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gomodels"
	"encoding/json"
)

var client *redis.Client
var PublishChan = make(chan models.Message)

func Run(){
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	if err == nil{
		go PublishMessage()
	}
}

func PublishMessage() {
	for{
		var message = <- PublishChan
		pubsub := client.Subscribe(message.Room.Hex())
		defer pubsub.Close()
		json,_ := json.Marshal(message)
		err := client.Publish(message.Room.Hex(), string(json)).Err()
		if err != nil {
			panic(err)
		}
	}
}