package main

import (
	"./server"
	"./controllers"
)

func main() {
	controllers.Run()
	server.Start(server.ServerProperties{Address: "/goxchange", Port: "8082"})
}
