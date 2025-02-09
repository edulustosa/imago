package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Error struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

type Errors struct {
	Errors []Error `json:"errors"`
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func Encode[T any](w http.ResponseWriter, status int, v T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, fmt.Sprintf("encode json: %v", err), http.StatusInternalServerError)
	}
}

func Decode[T any](r *http.Request) (T, map[string]string, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}

	if err := validate.Struct(v); err != nil {
		errs := err.(validator.ValidationErrors)
		problems := make(map[string]string, len(errs))

		for _, e := range errs {
			problems[strings.ToLower(e.Field())] = e.Error()
		}

		return v, problems, fmt.Errorf("validate: %w", err)
	}

	return v, nil, nil
}

func InvalidRequest(w http.ResponseWriter, problems map[string]string) {
	if problems == nil {
		SendError(w, http.StatusBadRequest, Error{
			Message: "failed to parse request",
		})
		return
	}

	errs := make([]Error, 0, len(problems))
	for field, tag := range problems {
		errs = append(errs, Error{
			Message: fmt.Sprintf("invalid %s", field),
			Details: tag,
		})
	}

	SendError(w, http.StatusBadRequest, errs...)
}

func SendError(w http.ResponseWriter, status int, errs ...Error) {
	Encode(w, status, Errors{Errors: errs})
}

func InternalError(w http.ResponseWriter) {
	SendError(w, http.StatusInternalServerError, Error{
		Message: "something went wrong, please try again later",
	})
}
