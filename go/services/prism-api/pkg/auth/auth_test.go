package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"code.prism.io/go/services/prism-api/pkg/config"
)

func TestAuth(t *testing.T) {
	e := echo.New()
	config := config.Auth0{
		Domain:   "test",
		Audience: "testaudience",
	}
	mw := Auth0TokenMiddleware(config)

	e.GET("/", mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}))

	req := httptest.NewRequest(echo.GET, "/", nil)
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}
