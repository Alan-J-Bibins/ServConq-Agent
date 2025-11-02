package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/jonasfj/go-localtunnel"
	"github.com/valyala/fasthttp"
)

func main() {
	// Setup localtunnel listener
	listener, err := localtunnel.Listen(localtunnel.Options{})
	if err != nil {
		log.Fatalf("localtunnel listen error: %v", err)
	}
	defer listener.Close()

	url := listener.Addr().String()
	fmt.Printf("Public tunnel URL: %s\n", url)

	app := fiber.New()
	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Fiber over LocalTunnel!")
	})
	app.Get("/metrics", monitor.New(monitor.Config{APIOnly: true}))

	// Wrap connection with fasthttp.Server
	server := &fasthttp.Server{
		Handler: app.Handler(), // this is a fasthttp.RequestHandler
	}

	// Serve using fasthttp.Server on the localtunnel listener
	if err := server.Serve(listener); err != nil {
		log.Fatalf("fasthttp server error: %v", err)
	}
}
