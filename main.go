package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
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

	url := listener.Addr().String()
	fmt.Printf("ServConq Agent Connection String: %s\n", url)

	app := fiber.New()
	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Fiber over LocalTunnel!")
	})
	app.Get("/metrics", monitor.New(monitor.Config{APIOnly: true}))
	app.Post("/run", AgentRunCommandHandler)

	server := &fasthttp.Server{
		Handler: app.Handler(),
	}

	err = server.Serve(listener)
	if err != nil {
		return fmt.Errorf("fasthttp server error: %w", err)
	}

	return nil
}

func AgentRunCommandHandler(c *fiber.Ctx) error {
	var req struct {
		Command string `json:"command"`
		Pwd     string `json:"pwd"`
	}
	log.Println("[main.go:63] req = ", req.Pwd)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	var cmd *exec.Cmd
	prefix := ""
	if runtime.GOOS == "windows" {
		if req.Pwd == "||$$$HOME$$$||" {
			prefix = "cd $env:USERPROFILE; "
		} else {
			prefix = "cd " + req.Pwd + "; "
		}
		cmd = exec.Command("powershell", "-Command", prefix+req.Command+"; pwd")
	} else {
		if req.Pwd == "||$$$HOME$$$||" {
			prefix = "cd $HOME; "
		} else {
			prefix = "cd " + req.Pwd + "; "
		}
		commandToBeRun := prefix + req.Command + "; pwd"
		log.Println("[main.go:86] commandToBeRun = ", commandToBeRun)
		cmd = exec.Command("bash", "-c", commandToBeRun)
	}

	out, err := cmd.CombinedOutput()
	outputStr := string(out)
	pwd := "||$$$HOME$$$||"
	cmdOutput := ""
	lines := strings.Split(outputStr, "\n")
	if len(lines) != 0 {
		pwd = strings.TrimSpace(lines[len(lines)-2])
		cmdOutput = strings.Join(lines[:len(lines)-2], "\n")
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  err.Error(),
			"output": cmdOutput,
			"pwd":    pwd,
		})
	}

	return c.JSON(fiber.Map{
		"output": cmdOutput,
		"pwd":    pwd,
		"error":  err,
	})
}
