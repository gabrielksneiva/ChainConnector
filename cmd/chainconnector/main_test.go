package main

import (
	"testing"

	httpPkg "ChainConnector/internal/adapters/http"

	"github.com/gofiber/fiber/v2"
)

func TestRun_Success(t *testing.T) {
	// Replace listen function to avoid actually binding a port.
	called := false
	httpPkg.SetListenFunc(func(_ *fiber.App, _ string) error {
		called = true
		return nil
	})
	defer httpPkg.ResetListenFunc()

	// Call the StartServer which should invoke the mocked listen function.
	app := httpPkg.CreateFiberServer()
	httpPkg.StartServer(app, ":3000")

	if !called {
		t.Fatalf("expected listen func to be called")
	}
}

func TestRun_ErrorReturned(t *testing.T) {
	// Replace listen function to return an error and ensure run returns it.
	testErr := errorString("listen failed")
	httpPkg.SetListenFunc(func(_ *fiber.App, _ string) error {
		return testErr
	})
	defer httpPkg.ResetListenFunc()

	// Call StartServerError and expect the error propagated.
	err := httpPkg.StartServerError(httpPkg.CreateFiberServer(), ":3000")
	if err != testErr {
		t.Fatalf("expected error %v, got %v", testErr, err)
	}
}

// simple error type to avoid importing errors package
type errorString string

func (e errorString) Error() string { return string(e) }

func TestMainFunction_WithMockedListen(t *testing.T) {
	called := false
	httpPkg.SetListenFunc(func(_ *fiber.App, _ string) error {
		called = true
		return nil
	})
	defer httpPkg.ResetListenFunc()

	// Call main() directly; with listen mocked it should return quickly
	main()

	if !called {
		t.Fatalf("expected mocked listen to be called from main()")
	}
}
