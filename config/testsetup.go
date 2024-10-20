package config

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v4"
)

// MakeEchoTest returns the context and res used for the echo route. It expects a ( json ) string as a body
func MakeEchoTest(body string) (rec *httptest.ResponseRecorder, c echo.Context) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	return rec, c
}
