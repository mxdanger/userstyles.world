package user

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"userstyles.world/database"
	"userstyles.world/handlers/jwt"
	"userstyles.world/models"
)

func Profile(c *fiber.Ctx) error {
	u, _ := jwt.User(c)
	p := c.Params("name")

	user, err := models.FindUserByName(database.DB, strings.ToLower(p))
	if err != nil {
		return c.Render("err", fiber.Map{
			"Title": "User not found",
			"User":  u,
		})
	}

	// Always redirect to correct URL.
	if p != user.Username {
		c.Redirect("/user/"+strings.ToLower(p), fiber.StatusSeeOther)
	}

	styles, err := models.GetStylesByUser(database.DB, p)
	if err != nil {
		return c.Render("err", fiber.Map{
			"User":  u,
			"Title": "Server error",
		})
	}

	// Render Account template if current user matches active session.
	/*if u.Username == p {
		return c.Render("account", fiber.Map{
			"Title":  "Account",
			"User":   u,
			"Styles": styles,
		})
	}*/

	return c.Render("profile", fiber.Map{
		"Title":  user.Username + "'s profile",
		"User":   u,
		"Params": user,
		"Styles": styles,
	})
}
