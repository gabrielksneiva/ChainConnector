package main

import (
	"ChainConnector/internal/app"

	"go.uber.org/fx"
)

func main() {
	fx.New(app.Modules).Run()
}
