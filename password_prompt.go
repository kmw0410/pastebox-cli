package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"

	"golang.org/x/term"
)

func readTerminalPassword(output io.Writer, prompt string) (string, error) {
	terminal, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", errors.New("interactive terminal is required")
	}
	defer terminal.Close()

	if _, err := fmt.Fprint(output, prompt); err != nil {
		return "", err
	}
	password, err := term.ReadPassword(int(terminal.Fd()))
	_, _ = fmt.Fprintln(output)
	if err != nil {
		return "", err
	}
	return string(password), nil
}

func (a application) promptExistingPassword(prompt string) (string, error) {
	return a.promptSecret(prompt, "password must not be empty")
}

func (a application) promptSecret(prompt string, emptyMessage string) (string, error) {
	if a.readPassword == nil {
		return "", errors.New("interactive terminal is required")
	}
	password, err := a.readPassword(prompt)
	if err != nil {
		return "", err
	}
	if password == "" {
		return "", errors.New(emptyMessage)
	}
	return password, nil
}

func (a application) promptNewPassword() (string, error) {
	password, err := a.promptExistingPassword("New paste password: ")
	if err != nil {
		return "", err
	}
	if !validPromptedPassword(password) {
		return "", errors.New("password must contain 8-128 characters without control characters")
	}
	confirmation, err := a.promptExistingPassword("Confirm new paste password: ")
	if err != nil {
		return "", err
	}
	if password != confirmation {
		return "", errors.New("passwords do not match")
	}
	return password, nil
}

func (a application) promptPasswordProtection() (string, bool, error) {
	if a.readPassword == nil {
		return "", false, errors.New("interactive terminal is required")
	}
	password, err := a.readPassword("New paste password (leave blank for random): ")
	if err != nil {
		return "", false, err
	}
	if password == "" {
		return "", true, nil
	}
	if !validPromptedPassword(password) {
		return "", false, errors.New("password must contain 8-128 characters without control characters")
	}
	confirmation, err := a.promptExistingPassword("Confirm new paste password: ")
	if err != nil {
		return "", false, err
	}
	if password != confirmation {
		return "", false, errors.New("passwords do not match")
	}
	return password, false, nil
}

func validPromptedPassword(password string) bool {
	if !utf8.ValidString(password) {
		return false
	}
	length := utf8.RuneCountInString(password)
	if length < 8 || length > 128 {
		return false
	}
	for _, r := range password {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
