package goprivategpt

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type Server struct {
	*fiber.App
	pgpt *PrivateGPT
}

func NewServer(pgpt *PrivateGPT) (*Server, error) {
	app := fiber.New()
	app.Use(logger.New())
	app.Use(requestid.New())
	srv := &Server{app, pgpt}
	return srv, nil
}

func (s *Server) askHandler(c *fiber.Ctx) error {
	c.Accepts("application/json")

	query := c.Queries()
	prompt := query["prompt"]
	if prompt == "" {
		return c.Status(400).SendString("empty prompt")
	}

	res, err := s.pgpt.Predict(context.Background(), prompt)
	if err != nil {
		return err
	}

	println("Prompt: ", prompt)
	println("Response: ", res)
	return c.JSON(map[string]any{"message": res})
}

func (s *Server) router() {
	s.App.Get("/api/ask", s.askHandler)
}

func (s *Server) Shutdown() error {
	return s.App.Shutdown()
}

func (s *Server) Start(addr string) error {
	s.router()
	return s.App.Listen(addr)
}
