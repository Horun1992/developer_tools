package main

import (
	"developers_tools/internal/server"
)

func main() {
	server.Start(func() { server.MainHandler() }, "127.0.0.1:5100")
}
