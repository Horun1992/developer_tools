package main

import (
	"developers_tools/internal/server"
)

func main() {
	server.Start(func() { server.RegisterRoutes() }, "0.0.0.0:5100")
}
