package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/jonasfj/go-localtunnel"
	"github.com/valyala/fasthttp"
)

func main() {
	for {
		err := startAgentServer()
		if err != nil {
			log.Printf("Agent crashed: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}
		break
	}
}

func startAgentServer() error {
	fmt.Println("Welcome to ServConq!")
	listener, err := localtunnel.Listen(localtunnel.Options{})
	if err != nil {
		return fmt.Errorf("localtunnel listen error: %w", err)
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
	app.Post("/run", AgentRunCommandHandler)
	fmt.Println("status check 3")

	server := &fasthttp.Server{
		Handler: app.Handler(),
	}
	fmt.Println("status check 4")

	err = server.Serve(listener)
	if err != nil {
		return fmt.Errorf("fasthttp server error: %w", err)
	}
	fmt.Println("status check 5")
	return nil
}

func AgentRunCommandHandler(c *fiber.Ctx) error {
	var req struct {
		Command string `json:"command"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", req.Command)
	} else {
		cmd = exec.Command("bash", "-c", req.Command)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  err.Error(),
			"output": string(out),
		})
	}
	return c.JSON(fiber.Map{"output": string(out)})
}
