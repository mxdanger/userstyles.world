package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/template/html"

	"userstyles.world/config"
	"userstyles.world/handlers/core"
	"userstyles.world/handlers/style"
	"userstyles.world/handlers/user"
)

func Initialize() {
	app := fiber.New(fiber.Config{
		Views:                 html.New("./views", ".html"),
		DisableStartupMessage: true,
	})

	app.Static("/", "./static")

	app.Get("/", core.Home)

	app.Get("/login", user.LoginGet)
	app.Post("/login", user.LoginPost)

	app.Post("/logout", user.Logout)
	app.Get("/account", user.Account)

	app.Get("/register", user.RegisterGet)
	app.Post("/register", user.RegisterPost)

	app.Get("/user/:name", user.Profile)

	app.Get("/explore", style.GetExplore)
	app.Get("/style/:id", style.GetStyle)

	app.Get("/monitor", monitor.New())

	app.Use(core.NotFound)

	log.Fatal(app.Listen(config.PORT))
}
