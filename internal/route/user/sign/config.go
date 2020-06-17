package sign

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	// Oauth2 scopes: https://developers.google.com/identity/protocols/oauth2/scopes
	oauthConf = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		Endpoint:     google.Endpoint,
		RedirectURL:  "",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",   // email address
			"https://www.googleapis.com/auth/userinfo.profile", // public profile
			"openid", // open id connect
		},
	}
	domain = ""
)

type GoogleOauthConfig struct {
	ClientId     string
	ClientSecret string
	RedirectUrl  string
}

func VarConfig(_domain string, googleConf GoogleOauthConfig) {
	oauthConf.RedirectURL = googleConf.RedirectUrl
	domain = _domain
	oauthConf.ClientID = googleConf.ClientId
	oauthConf.ClientSecret = googleConf.ClientSecret
}
