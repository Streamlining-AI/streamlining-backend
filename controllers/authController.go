package controllers

import "github.com/gofiber/fiber/v2"

func SendHello(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}

func Register(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return err
	}

	return c.JSON(data)
}
