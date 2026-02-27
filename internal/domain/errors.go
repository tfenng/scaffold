package domain

import "net/http"

type Code string

const (
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeNotFound        Code = "NOT_FOUND"
	CodeConflict        Code = "CONFLICT"
	CodeInternal        Code = "INTERNAL"
)

type AppError struct {
	Code       Code   `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Cause      error  `json:"-"`
}

func (e *AppError) Error() string { return string(e.Code) + ": " + e.Message }

func Invalid(msg string) *AppError  { return &AppError{Code: CodeInvalidArgument, Message: msg, HTTPStatus: http.StatusBadRequest} }
func NotFound(msg string) *AppError { return &AppError{Code: CodeNotFound, Message: msg, HTTPStatus: http.StatusNotFound} }
func Conflict(msg string) *AppError { return &AppError{Code: CodeConflict, Message: msg, HTTPStatus: http.StatusConflict} }
func Internal(err error) *AppError  { return &AppError{Code: CodeInternal, Message: "internal error", HTTPStatus: http.StatusInternalServerError, Cause: err} }
