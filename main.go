package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-playground/validator/v10"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// ContactRequest represents the incoming request
type ContactRequest struct {
	Email   string `validate:"required,email"`
	Message string `validate:"required"`
}

// Response is an alias for events.APIGatewayProxyResponse
type Response events.APIGatewayProxyResponse

func errorMessageMap(validationTag string) string {
	messages := map[string]string{
		"required": "is required",
		"email":    "is not a valid email",
	}
	return messages[validationTag]
}

func validateEmail(fl validator.FieldLevel) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return re.MatchString(fl.Field().String())
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	validate := validator.New()
	validate.RegisterValidation("email", validateEmail)

	contactRequest := ContactRequest{
		Email:   "",
		Message: "",
	}

	// Unmarshal the json into our type return 404 if error
	err := json.Unmarshal([]byte(request.Body), &contactRequest)
	if err != nil {
		return Response{Body: err.Error(), StatusCode: 404}, nil
	}

	// Validate the incoming request
	validationErr := validate.Struct(contactRequest)
	if validationErr != nil {
		errors := make(map[string]interface{})

		for _, fieldErr := range validationErr.(validator.ValidationErrors) {
			var fieldErrors []string
			msg := fmt.Sprintf("The %s field %s", strings.ToLower(fieldErr.Field()), errorMessageMap(fieldErr.Tag()))
			fieldErrors = append(fieldErrors, msg)
			errors[strings.ToLower(fieldErr.Field())] = fieldErrors
		}

		errorRes := map[string]interface{}{
			"errors": errors,
		}

		return handleResponse(errorRes, 422)
	}

	// Send email here
	from := mail.NewEmail(contactRequest.Email, contactRequest.Email)
	subject := "Contact form submission by website"
	to := mail.NewEmail("Alex Kearns", "alex@alexkearns.co.uk")
	plainTextContent := contactRequest.Message
	htmlContent := contactRequest.Message
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err = client.Send(message)
	if err != nil {
		return Response{Body: err.Error(), StatusCode: 500}, nil
	}

	successResponse := map[string]interface{}{
		"status": "OK",
	}
	return handleResponse(successResponse, 200)
}

func main() {
	lambda.Start(handleRequest)
}

func handleResponse(responseBody map[string]interface{}, statusCode int) (Response, error) {
	var buf bytes.Buffer
	body, err := json.Marshal(responseBody)
	if err != nil {
		return Response{Body: err.Error(), StatusCode: 404}, nil
	}

	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      statusCode,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil
}
