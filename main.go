package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

func main() {
	app := fiber.New()

	app.Get("/metrics", monitor.New(monitor.Config{APIOnly: true}))
	app.Listen(":8000")
}
