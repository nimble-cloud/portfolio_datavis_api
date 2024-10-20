package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"nimblecloud/portfolio_datavis/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
	// Companies []string `json:"companies"`

	IsAdmin bool
}

func MakeYat(id string, companies []string) ([]byte, error) {
	idMap := map[string]any{
		"id":        id,
		"companies": companies,
	}

	idByte, err := json.Marshal(idMap)
	if err != nil {
		fmt.Println("Error marshalling idMap", err)
		return nil, err
	}

	return idByte, nil
}

func CreateUser(e echo.Context) error {

	u := User{}
	if err := e.Bind(&u); err != nil {
		fmt.Println("Error binding user", err.Error())
		return err
	}

	if u.Role == "" || u.Email == "" || u.Password == "" {
		return echo.NewHTTPError(400, "Role, email, and password are required") //TODO make this less of a specific error
	}

	if u.Role != "admin" {
		return echo.NewHTTPError(400, "User must have at least one company")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		fmt.Println("Error hashing password", err)
		return echo.NewHTTPError(500, "Error hashing password") //TODO make this less of a specific error
	}

	fmt.Println("Hashed password: ", string(hashedPassword))
	u.Password = string(hashedPassword)

	err = config.DB.QueryRow(context.Background(), "INSERT INTO users (email, pass, role) VALUES ($1, $2, $3) RETURNING id", u.Email, u.Password, u.Role).Scan(&u.ID)
	if err != nil {
		fmt.Println("Error creating user", err)
		return echo.NewHTTPError(500, "Error inserting user") //TODO make this less of a specific error
	}

	return e.JSON(201, map[string]string{"id": u.ID, "email": u.Email, "role": u.Role})

}

type UserClaims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	// Companies []string `json:"companies"`
	// Yat       string   `json:"yat"`

	jwt.RegisteredClaims
}

func Login(e echo.Context) error {

	u := User{}
	if err := e.Bind(&u); err != nil {
		fmt.Println("Error binding user", err.Error())
		return err
	}

	if u.Email == "" || u.Password == "" {
		return echo.NewHTTPError(400, "Email and password are required") //TODO make this less of a specific error
	}

	foundPassword := ""
	err := config.DB.QueryRow(context.Background(), "SELECT id, role, pass FROM users WHERE email = $1", u.Email).Scan(&u.ID, &u.Role, &foundPassword)
	if err != nil {
		fmt.Println("Error getting user", err)
		return echo.NewHTTPError(500, "Error getting user") //TODO make this less of a specific error
	}

	// fmt.Println("Found password: ", foundPassword)

	err = bcrypt.CompareHashAndPassword([]byte(foundPassword), []byte(u.Password))
	if err != nil {
		fmt.Println("Error comparing password", err)
		return echo.NewHTTPError(500, "Error comparing password") //TODO make this less of a specific error
	}

	// idByte, err := MakeYat(u.ID, u.Companies)
	// if err != nil {
	// 	return echo.NewHTTPError(500, "Error marshalling idMap") //TODO make this less of a specific error
	// }

	// yat, err := bcrypt.GenerateFromPassword([]byte(idByte), 5)
	// if err != nil {
	// 	fmt.Println("Error hashing yat", err)
	// 	return echo.NewHTTPError(500, "Error hashing yat") //TODO make this less of a specific error
	// }

	claims := UserClaims{
		u.ID,
		u.Role,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		fmt.Println("Error signing token", err)
		return echo.NewHTTPError(500, "Error signing token") //TODO make this less of a specific error
	}

	return e.JSON(200, map[string]any{"access_token": ss, "id": u.ID, "role": u.Role})
}
