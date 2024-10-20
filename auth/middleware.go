package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func authErr(err error) error {
	fmt.Println("Error authenticating user", err)
	return echo.NewHTTPError(401)
}

func IsAuthed(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(401)
		}

		tokenString := strings.Split(authHeader, "Bearer ")[1]
		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			fmt.Println("Error parsing token", err)
			return authErr(errors.New("credential error"))
		}

		switch {
		case token.Valid:
			u := User{
				ID:      claims.ID,
				Role:    claims.Role,
				IsAdmin: claims.Role == "admin",
			}

			c.Set("user", &u)

			return next(c)
		case errors.Is(err, jwt.ErrTokenMalformed):
			return authErr(errors.New("malformed token"))
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			// Invalid signature
			return authErr(errors.New("invalid signature"))
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			// Token is either expired or not active yet
			return authErr(errors.New("token expired or not active yet"))
		default:
			return authErr(errors.New("couldn't handle this token >>"))
		}
	}
}
