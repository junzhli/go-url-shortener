package sign

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	// Oauth2 scopes: https://developers.google.com/identity/protocols/oauth2/scopes
	oauthConf = &oauth2.Config{
		ClientID:     "959723324236-0e23oe704fp1rtf3k5qc780mijahd1b3.apps.googleusercontent.com",
		ClientSecret: "xG1-yt61nKfvPUAfZumduCNO",
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

func VarConfig(redirectUrl string, _domain string) {
	oauthConf.RedirectURL = redirectUrl
	domain = _domain
	// TODO:
	//oauthConf.ClientID = clientId
	//oauthConf.ClientSecret = clientSecret
}
