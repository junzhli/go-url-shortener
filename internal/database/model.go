package database

var (
	UserTypeGoogle = "google" // Google Login
	UserTypeLocal  = "local"  // Local account
)

type User struct {
	UserID   string
	Email    string
	Type     string // local: Local Account without Oauth service, google: Google Account
	Password string
}

type GoogleUser struct {
	UserID     string
	GoogleUUID string
}

type URL struct {
	OriginURL  string
	Owner      string
	ShortenURL string
}
