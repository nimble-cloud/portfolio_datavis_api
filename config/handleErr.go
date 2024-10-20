package config

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

func HandleErr(err error, msg string) error {
	errMsg := fmt.Sprint(msg, " >> ", err.Error())
	fmt.Println(errMsg)
	return echo.NewHTTPError(500, errMsg)
}
