package error

import "github.com/gin-gonic/gin"

var (
	InvalidJSONStringError  = "Invalid json string"
	AuthenticationError     = "Authentication failed"
	EmailValidationError    = "Email validation failed"
	PasswordValidationError = "Password validation failed"
	CodeValidationError     = "Code validation failed"
	AlreadyRegisteredError  = "Already registered"
	RequestError            = "Invalid request"
)

func NewResponseErrorWithMessage(error string) gin.H {
	return gin.H{
		"error": error,
	}
}
