package respond

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

// Success represents a standard structure for successful responses.
type Success struct {
	Result interface{} `json:"result"`
}

// Error represents a standard structure for error responses.
type Error struct {
	Message string `json:"message"`
}

// JPEG sends a JPEG image as the HTTP response with the specified status code.
// It sets the Content-Type header to "image/jpeg" and writes the image bytes
// directly using the Gin context.
func JPEG(c *ginext.Context, status int, data []byte) {
	c.Data(status, "image/jpeg", data)
}

// JSON sends a JSON response with the specified HTTP status code and data.
// It uses the Gin context to encode the data into JSON format.
func JSON(c *ginext.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// OK sends a 200 OK JSON response, wrapping the given result in a Success struct.
func OK(c *ginext.Context, result interface{}) {
	JSON(c, http.StatusOK, Success{Result: result})
}

// Created sends a 201 Created JSON response, wrapping the given result in a Success struct.
func Created(c *ginext.Context, result interface{}) {
	JSON(c, http.StatusCreated, Success{Result: result})
}

// Fail sends an error JSON response with the specified HTTP status code.
// The error message is wrapped in an Error struct.
func Fail(c *ginext.Context, status int, err error) {
	JSON(c, status, Error{Message: err.Error()})
}
