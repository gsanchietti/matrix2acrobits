package main

import (
	"log"
	"os"

	"github.com/gsanchietti/matrix2acrobits/internal/api"
	"github.com/gsanchietti/matrix2acrobits/internal/matrix"
	"github.com/gsanchietti/matrix2acrobits/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = defaultPort
	}

	e := echo.New()
	e.HideBanner = true
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	homeserver := os.Getenv("MATRIX_HOMESERVER_URL")
	if homeserver == "" {
		log.Fatal("MATRIX_HOMESERVER_URL is required")
	}

	adminToken := os.Getenv("SUPER_ADMIN_TOKEN")
	if adminToken == "" {
		log.Fatal("SUPER_ADMIN_TOKEN is required")
	}

	matrixClient, err := matrix.NewClient(matrix.Config{HomeserverURL: homeserver})
	if err != nil {
		log.Fatalf("initialize matrix client: %v", err)
	}

	svc := service.NewMessageService(matrixClient)
	api.RegisterRoutes(e, svc, adminToken)

	log.Printf("listening on :%s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
