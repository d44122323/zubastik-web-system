package main

import (
	"errors"
	"strings"
)

func ValidateForm(data FormData) error {
	if len(strings.TrimSpace(data.Name)) < 2 {
		return errors.New("введите имя")
	}
	if len(strings.TrimSpace(data.Phone)) < 5 {
		return errors.New("введите телефон")
	}
	return nil
}
