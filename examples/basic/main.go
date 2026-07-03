// Command basic demonstrates a small express application in Go.
package main

import (
	"log"

	"github.com/malcolmston/express"
)

func main() {
	app := express.New()

	// Global middleware.
	app.Use(express.Logger())
	app.Use(express.Recover())
	app.Use(express.JSON())

	// A simple route.
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("Hello from express-go!")
	})

	// Route parameters.
	app.Get("/users/:id", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON(map[string]any{
			"id":   req.Params("id"),
			"name": "User " + req.Params("id"),
		})
	})

	// Reading a JSON body.
	app.Post("/echo", func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(201).JSON(map[string]any{"youSent": req.Body()})
	})

	// A mounted sub-router.
	api := express.NewRouter()
	api.Get("/status", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON(map[string]string{"status": "ok"})
	})
	app.Use("/api", api)

	log.Println("listening on :3000")
	log.Fatal(app.Listen(":3000"))
}
