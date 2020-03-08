package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ContactRequest represents the incoming request
type ContactRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Message string `json:"message" binding:"required"`
}

func errorMessageMap(validationTag string) string {
	messages := map[string]string{
		"required": "is required",
		"email":    "is not a valid email",
	}
	return messages[validationTag]
}

func main() {
	validate := validator.New()
	validate.RegisterValidation("email", validateEmail)

	r := gin.Default()
	r.POST("/contact", contactHandler)
	r.Run()
}

func validateEmail(fl validator.FieldLevel) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return re.MatchString(fl.Field().String())
}

func contactHandler(c *gin.Context) {
	var request ContactRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		type fieldErrorsType map[string][]string
		errors := make(fieldErrorsType)

		for _, fieldErr := range err.(validator.ValidationErrors) {
			var fieldErrors []string
			msg := fmt.Sprintf("The %s field %s", strings.ToLower(fieldErr.Field()), errorMessageMap(fieldErr.Tag()))
			fieldErrors = append(fieldErrors, msg)
			errors[strings.ToLower(fieldErr.Field())] = fieldErrors
		}

		errorRes := make(map[string]fieldErrorsType)
		errorRes["errors"] = errors
		c.JSON(http.StatusBadRequest, errorRes)

		return
	}

	// TODO: Send request to my email address

	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}
