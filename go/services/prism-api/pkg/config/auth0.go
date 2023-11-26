package config

import "github.com/kelseyhightower/envconfig"

type Auth0 struct {
	Domain      string  `envconfig:"AUTH0_DOMAIN"`
	Audience    string  `envconfig:"AUTH0_AUDIENCE"`
	AccessToken *string `envconfig:"AUTH0_ACCESS_TOKEN"`
}

func GetAuth0Config() Auth0 {
	var auth0 Auth0
	envconfig.MustProcess("", &auth0)
	return auth0
}
