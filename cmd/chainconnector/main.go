package main

import "ChainConnector/internal/adapters/http"

func main() {
	app := http.CreateFiberServer()
	http.StartServer(app, ":3000")
}
