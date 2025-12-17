package http

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func CreateFiberServer() *fiber.App {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("criando servidor fiber")
	})

	setupHealthcheckRoute(app)

	return app
}

func StartServer(app *fiber.App, port string) {
	logServerStart(port)
	if err := listenFunc(app, port); err != nil {
		log.Fatal(err)
	}
}

func StartServerError(app *fiber.App, port string) error {
	logServerStart(port)
	return listenFunc(app, port)
}

func logServerStart(port string) {
	log.Println("Servidor rodando na porta " + port + "...")
}

// listenFunc is the function used to start the fiber app. It is replaceable
// in tests via SetListenFunc / ResetListenFunc.
var listenFunc = defaultListenFunc

// defaultListenImpl is the underlying implementation used by defaultListenFunc.
// Tests can replace this via SetDefaultListenImpl to avoid calling the real
// `app.Listen` which would block.
var defaultListenImpl = func(app *fiber.App, port string) error {
	return app.Listen(port)
}

func defaultListenFunc(app *fiber.App, port string) error {
	return defaultListenImpl(app, port)
}

// SetDefaultListenImpl replaces the internal implementation used by
// defaultListenFunc. Only intended for tests.
func SetDefaultListenImpl(f func(*fiber.App, string) error) {
	defaultListenImpl = f
}

// ResetDefaultListenImpl restores the original implementation.
func ResetDefaultListenImpl() {
	defaultListenImpl = func(app *fiber.App, port string) error {
		return app.Listen(port)
	}
}

// SetListenFunc sets a custom listen function (for tests).
func SetListenFunc(f func(*fiber.App, string) error) {
	listenFunc = f
}

// ResetListenFunc restores the default listen function.
func ResetListenFunc() {
	listenFunc = defaultListenFunc
}

func setupHealthcheckRoute(app *fiber.App) {
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
}
