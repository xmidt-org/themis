package xhttp

import (
	"bytes"
	"net/http"
)

// MissingValueError indicates a missing header or parameter in a request (or both)
type MissingValueError struct {
	Header    string
	Parameter string
}

func (mve MissingValueError) Error() string {
	var output bytes.Buffer
	output.WriteString("Missing value from")
	if len(mve.Header) > 0 {
		output.WriteString(" header '")
		output.WriteString(mve.Header)
		output.WriteString("'")

		if len(mve.Parameter) > 0 {
			output.WriteString(" or parameter '")
			output.WriteString(mve.Parameter)
			output.WriteString("'")
		}
	} else if len(mve.Parameter) > 0 {
		output.WriteString(" parameter '")
		output.WriteString(mve.Parameter)
		output.WriteString("'")
	}

	return output.String()
}

func (mve MissingValueError) StatusCode() int {
	return http.StatusBadRequest
}

// MissingVariableError indicates a missing URI variable, which is a misconfiguration
type MissingVariableError struct {
	Variable string
}

func (mve MissingVariableError) Error() string {
	var output bytes.Buffer
	output.WriteString("Missing URI variable `")
	output.WriteString(mve.Variable)
	output.WriteString("'")

	return output.String()
}

func (mve MissingVariableError) StatusCode() int {
	return http.StatusInternalServerError
}
