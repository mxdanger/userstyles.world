package user

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"userstyles.world/handlers/jwt"
	"userstyles.world/models"
	"userstyles.world/modules/database"
	"userstyles.world/modules/log"
	"userstyles.world/utils"
)

func Account(c *fiber.Ctx) error {
	u, _ := jwt.User(c)

	styles, err := models.GetStylesByUser(u.Username)
	if err != nil {
		return c.Render("err", fiber.Map{
			"Title": "Server error",
			"User":  u,
		})
	}

	user, err := models.FindUserByName(u.Username)
	if err != nil {
		return c.Render("err", fiber.Map{
			"Title": "User not found",
			"User":  u,
		})
	}

	return c.Render("user/account", fiber.Map{
		"Title":  "Account",
		"User":   u,
		"Params": user,
		"Styles": styles,
	})
}

func EditAccount(c *fiber.Ctx) error {
	u, _ := jwt.User(c)

	user, err := models.FindUserByName(u.Username)
	if err != nil {
		return c.Render("err", fiber.Map{
			"Title": "User not found",
			"User":  u,
		})
	}

	record := map[string]interface{}{"id": user.ID}

	form := c.Params("form")
	switch form {
	case "name":
		name := strings.TrimSpace(c.FormValue("name"))
		prev := user.DisplayName
		user.DisplayName = name

		if name == "" {
			log.Info.Printf("User %v cleared their name\n", u.ID)
			record["display_name"] = name
			break
		}

		if err := utils.Validate().StructPartial(user, "DisplayName"); err != nil {
			var validationError validator.ValidationErrors
			if ok := errors.As(err, &validationError); ok {
				log.Info.Printf("Validation errors for user %d: %v\n", u.ID, validationError)
			}
			user.DisplayName = prev

			l := len(name)
			var e string
			switch {
			case l < 3 || l > 32:
				e = "Display name must be between 3 and 32 characters."
			default:
				e = "Make sure your input contains valid characters."
			}

			return c.Render("user/account", fiber.Map{
				"Title":  "Validation Error",
				"User":   u,
				"Params": user,
				"Error":  e,
			})
		}

		record["display_name"] = name

	case "password":
		current := c.FormValue("current")
		if err := utils.CompareHashedPassword(user.Password, current); err != nil {
			return c.Status(fiber.StatusForbidden).Render("err", fiber.Map{
				"Title": "Failed to match current password",
				"User":  u,
			})
		}

		newPassword, confirmPassword := c.FormValue("new_password"), c.FormValue("confirm_password")
		if confirmPassword != newPassword {
			return c.Status(fiber.StatusForbidden).Render("err", fiber.Map{
				"Title": "Failed to match new passwords",
				"User":  u,
			})
		}

		user.Password = newPassword
		if err := utils.Validate().StructPartial(u, "Password"); err != nil {
			var validationError validator.ValidationErrors
			if ok := errors.As(err, &validationError); ok {
				log.Info.Println("Password change error:", validationError)
			}
			return c.Status(fiber.StatusForbidden).Render("err", fiber.Map{
				"Title": "Register failed",
				"User":  u,
			})
		}

		user.Password = utils.GenerateHashedPassword(newPassword)
		record["password"] = user.Password

	case "bio":
		bio := strings.TrimSpace(c.FormValue("bio"))
		prev := user.Biography
		user.Biography = bio

		if err := utils.Validate().StructPartial(user, "Biography"); err != nil {
			var validationError validator.ValidationErrors
			if ok := errors.As(err, &validationError); ok {
				log.Info.Println("Validation errors:", validationError)
			}
			user.Biography = prev

			return c.Render("user/account", fiber.Map{
				"Title":  "Validation Error",
				"User":   u,
				"Params": user,
				"Error":  "Biography must be shorter than 512 characters.",
			})
		}

		record["biography"] = bio

	case "socials":
		record["github"] = strings.TrimSpace(c.FormValue("github"))
		record["gitlab"] = strings.TrimSpace(c.FormValue("gitlab"))
		record["codeberg"] = strings.TrimSpace(c.FormValue("codeberg"))

	default:
		return c.Render("err", fiber.Map{
			"Title": "Invalid form",
			"User":  u,
		})
	}

	dbErr := database.Conn.
		Model(models.User{}).
		Where("id", record["id"]).
		Updates(record).
		Error

	if dbErr != nil {
		log.Warn.Println("Updating user profile failed, err:", err)
		return c.Render("err", fiber.Map{
			"Title": "Internal server error.",
			"User":  u,
		})
	}

	return c.Status(fiber.StatusSeeOther).Redirect("/account#" + form)
}
