package auth

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"

	"code.prism.io/go/pkg/must"
	"code.prism.io/go/services/prism-api/pkg/config"
)

type CustomClaims struct {
	Scope string `json:"scope"`
}

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func Auth0TokenMiddleware(auth0 config.Auth0) echo.MiddlewareFunc {
	issuerURL := must.Must(url.Parse("https://" + auth0.Domain + "/"))
	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
	jwtValidator := must.Must(validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{auth0.Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(1*time.Minute),
	))

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)
	return echo.WrapMiddleware(func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	})
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
}
