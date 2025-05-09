package utils

import (
	"github.com/gofiber/contrib/fiberi18n/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func Translate(c *fiber.Ctx, param *string, MessageID string) string {
	var translate string
	var err error
	if param != nil {
		translate, err = fiberi18n.Localize(c, &i18n.LocalizeConfig{
			MessageID: MessageID,
			TemplateData: map[string]interface{}{
				"name": param,
			},
		})
	} else {
		translate, err = fiberi18n.Localize(c, &i18n.LocalizeConfig{
			MessageID: MessageID,
		})
	}

	if err != nil {
		return "translate not found"
	}

	return translate

}
