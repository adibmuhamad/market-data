package helper

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"

)

type Response struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data"`
}

type Meta struct {
	Code    int `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func APIResponse(message string, code int, status string, data interface{}) Response {
	meta := Meta{
		Message: message,
		Code:    code,
		Status:  status,
	}

	jsonResponse := Response{
		Meta: meta,
		Data: data,
	}

	return jsonResponse
}

func FormatValidationError(err error) []string {
	var errors []string

	for _, e := range err.(validator.ValidationErrors) {
		errors = append(errors, e.Error())
	}

	return errors
}

func Marshaler(a interface{}) (map[string]interface{}, error) {
	marshal, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	var b map[string]interface{}
	err = json.Unmarshal(marshal, &b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
