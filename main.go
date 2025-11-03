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
	fmt.Println("Welcome to ServConq!\n\nTo get started with connecting your server copy the connection string given below and paste it in the form for creating a new server");
	listener, err := localtunnel.Listen(localtunnel.Options{})
	if err != nil {
		log.Fatalf("localtunnel listen error: %v", err)
	}
	defer listener.Close()
	fmt.Println("status check 1")

	url := listener.Addr().String()
	fmt.Printf("ServConq Agent Connection String: %s\n", url)

	app := fiber.New()
	app.Use(logger.New())
	fmt.Println("status check 2")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Fiber over LocalTunnel!")
	})
	app.Get("/metrics", monitor.New(monitor.Config{APIOnly: true}))

	fmt.Println("status check 3")
	// Wrap connection with fasthttp.Server
	server := &fasthttp.Server{
		Handler: app.Handler(), // this is a fasthttp.RequestHandler
	}
	fmt.Println("status check 4")

	// Serve using fasthttp.Server on the localtunnel listener
	if err := server.Serve(listener); err != nil {
		log.Fatalf("fasthttp server error: %v", err)
	}
	fmt.Println("status check 5")
}
