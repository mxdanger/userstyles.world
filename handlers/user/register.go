package user

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"userstyles.world/handlers/jwt"
	"userstyles.world/models"
	"userstyles.world/modules/config"
	"userstyles.world/modules/log"
	"userstyles.world/utils"
)

func RegisterGet(c *fiber.Ctx) error {
	if u, ok := jwt.User(c); ok {
		log.Info.Printf("User %d has set session, redirecting.\n", u.ID)
		return c.Redirect("/account", fiber.StatusSeeOther)
	}

	return c.Render("user/register", fiber.Map{
		"Title":     "Register",
		"Canonical": "register",
	})
}

func RegisterPost(c *fiber.Ctx) error {
	password, confirm := c.FormValue("password"), c.FormValue("confirm")
	if confirm != password {
		return c.Status(fiber.StatusForbidden).
			Render("user/register", fiber.Map{
				"Title": "Register failed",
				"Error": "Your passwords don't match.",
			})
	}

	u := models.User{
		Username: c.FormValue("username"),
		Password: password,
		Email:    c.FormValue("email"),
	}

	err := utils.Validate().StructPartial(u, "Username", "Email", "Password")
	if err != nil {
		var validationError validator.ValidationErrors
		if ok := errors.As(err, &validationError); ok {
			log.Info.Println("Validation errors:", validationError)
		}
		return c.Status(fiber.StatusInternalServerError).
			Render("user/register", fiber.Map{
				"Title": "Register failed",
				"Error": "Failed to register. Make sure your input is correct.",
			})
	}

	token, err := utils.NewJWTToken().
		SetClaim("username", strings.ToLower(u.Username)).
		SetClaim("password", u.Password).
		SetClaim("email", u.Email).
		SetExpiration(time.Now().Add(time.Hour * 2)).
		GetSignedString(utils.VerifySigningKey)
	if err != nil {
		log.Warn.Println("Failed to create a JWT Token:", err.Error())
		return c.Status(fiber.StatusInternalServerError).
			Render("err", fiber.Map{
				"Title": "Internal server error",
			})
	}

	link := c.BaseURL() + "/verify/" + utils.EncryptText(token, utils.AEADCrypto, config.ScrambleConfig)

	partPlain := utils.NewPart().
		SetBody("Verify your UserStyles.world account by clicking the link below.\n" +
			"The link will expire in 2 hours\n\n" +
			link + "\n\n" +
			"You can safely ignore this e-mail if you never made an account for UserStyles.world.")
	partHTML := utils.NewPart().
		SetBody("<p>Verify your UserStyles.world account by clicking the link below.</p>\n" +
			"<b>The link will expire in 2 hours</b>\n" +
			"<br>\n" +
			"<a target=\"_blank\" clicktracking=\"off\" href=\"" + link + "\">Verifcation link</a>\n" +
			"<br><br>\n" +
			"<p>You can safely ignore this e-mail if you never made an account for UserStyles.world.</p>").
		SetContentType("text/html")

	err = utils.NewEmail().
		SetTo(u.Email).
		SetSubject("Verify your email address").
		AddPart(*partPlain).
		AddPart(*partHTML).
		SendEmail(config.IMAPServer)

	if err != nil {
		log.Warn.Println("Failed to send an email:", err.Error())
		return c.Status(fiber.StatusInternalServerError).
			Render("err", fiber.Map{
				"Title": "Internal server error",
				"Error": "Failed to send e-mail.",
			})
	}

	return c.Render("user/email-sent", fiber.Map{
		"Title":  "Email verifcation",
		"Reason": "Verification link has been sent to your e-mail address.",
	})
}
