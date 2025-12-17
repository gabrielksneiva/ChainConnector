package main

import "ChainConnector/internal/adapters/http"

func main() {
	app := http.CreateFiberServer()
	http.StartServer(app, ":3000")
}

// run is the entry point separated for testing. It returns any error
// produced when starting the server.
func run() error {
	app := http.CreateFiberServer()
	return http.StartServerError(app, ":3000")
}
