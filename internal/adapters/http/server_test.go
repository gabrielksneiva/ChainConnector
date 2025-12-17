package http

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCreateFiberServer(t *testing.T) {
	app := CreateFiberServer()

	if app == nil {
		t.Fatal("esperado app não-nil, obteve nil")
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET / deve retornar mensagem de criação",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody:   "criando servidor fiber",
		},
		{
			name:           "GET /healthz deve retornar OK",
			path:           "/healthz",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req)

			if err != nil {
				t.Fatalf("erro ao fazer requisição: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("esperado status %d, obteve %d", tt.expectedStatus, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("erro ao ler resposta: %v", err)
			}

			if string(body) != tt.expectedBody {
				t.Errorf("esperado body '%s', obteve '%s'", tt.expectedBody, string(body))
			}
		})
	}
}

func TestSetupHealthcheckRoute(t *testing.T) {
	app := fiber.New()
	setupHealthcheckRoute(app)

	req, _ := http.NewRequest("GET", "/healthz", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("erro ao testar rota: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("esperado status %d, obteve %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Errorf("esperado body 'OK', obteve '%s'", string(body))
	}
}

func TestRootRoute(t *testing.T) {
	app := CreateFiberServer()

	req, _ := http.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("erro ao testar rota raiz: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("esperado status %d, obteve %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	expectedBody := "criando servidor fiber"
	if string(body) != expectedBody {
		t.Errorf("esperado body '%s', obteve '%s'", expectedBody, string(body))
	}
}

func TestHealthzRoute(t *testing.T) {
	app := CreateFiberServer()

	req, _ := http.NewRequest("GET", "/healthz", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("erro ao testar rota de saúde: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("esperado status %d, obteve %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Errorf("esperado body 'OK', obteve '%s'", string(body))
	}
}

func TestStartServerError(t *testing.T) {
	app := CreateFiberServer()

	// Usar uma porta inválida para testar o erro sem bloquear
	// Porém, Listen não testa em background, apenas retorna erro
	// Este teste valida que StartServerError retorna um erro quando não consegue bindar
	_ = app

	// Nota: Não é possível testar este completamente sem bloquear
	// porque app.Listen() é blocking. Um teste real exigiria mocking.
}

func TestStartServer_UsesListenFunc(t *testing.T) {
	called := false
	SetListenFunc(func(_ *fiber.App, _ string) error {
		called = true
		return nil
	})
	defer ResetListenFunc()

	app := CreateFiberServer()
	StartServer(app, ":0")

	if !called {
		t.Fatalf("expected listen func to be called")
	}
}

func TestStartServerError_ReturnsError(t *testing.T) {
	SetListenFunc(func(_ *fiber.App, _ string) error {
		return errors.New("listen failed")
	})
	defer ResetListenFunc()

	app := CreateFiberServer()
	if err := StartServerError(app, ":0"); err == nil {
		t.Fatalf("expected StartServerError to return error")
	}
}

func TestDefaultListenFunc_UsesImpl(t *testing.T) {
	called := false
	SetDefaultListenImpl(func(_ *fiber.App, _ string) error {
		called = true
		return nil
	})
	defer ResetDefaultListenImpl()

	// Call defaultListenFunc which delegates to defaultListenImpl
	if err := defaultListenFunc(nil, ":0"); err != nil {
		t.Fatalf("expected no error from defaultListenFunc stub, got %v", err)
	}
	if !called {
		t.Fatalf("expected defaultListenImpl to be called")
	}
}

func TestLogServerStart(t *testing.T) {
	// Este teste apenas valida que a função existe
	// A função logServerStart não retorna nada e apenas registra um log
	logServerStart(":3000")
}

func TestCreateFiberServerInitialization(t *testing.T) {
	app := CreateFiberServer()

	if app == nil {
		t.Fatal("esperado app não-nil")
	}

	// Testar que o app tem rotas configuradas
	routes := app.GetRoutes()
	if len(routes) < 2 {
		t.Errorf("esperado pelo menos 2 rotas, obteve %d", len(routes))
	}
}
