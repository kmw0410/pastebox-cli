package main

import (
	"errors"
	"strings"
	"testing"
)

func testPasswordReader(values ...string) func(string) (string, error) {
	index := 0
	return func(string) (string, error) {
		if index >= len(values) {
			return "", errors.New("unexpected password prompt")
		}
		value := values[index]
		index++
		return value, nil
	}
}

func TestPromptNewPasswordRequiresMatchingConfirmation(t *testing.T) {
	app := application{readPassword: testPasswordReader("custom-secret", "different-secret")}
	if _, err := app.promptNewPassword(); err == nil || !strings.Contains(err.Error(), "do not match") {
		t.Fatalf("error = %v", err)
	}
}

func TestPromptNewPasswordRejectsShortPassword(t *testing.T) {
	app := application{readPassword: testPasswordReader("short")}
	if _, err := app.promptNewPassword(); err == nil || !strings.Contains(err.Error(), "8-128") {
		t.Fatalf("error = %v", err)
	}
}

func TestPromptPasswordProtectionUsesRandomPasswordForEmptyInput(t *testing.T) {
	app := application{readPassword: testPasswordReader("")}
	password, random, err := app.promptPasswordProtection()
	if err != nil {
		t.Fatalf("promptPasswordProtection failed: %v", err)
	}
	if password != "" || !random {
		t.Fatalf("password = %q, random = %v", password, random)
	}
}

func TestPromptPasswordProtectionConfirmsCustomPassword(t *testing.T) {
	app := application{readPassword: testPasswordReader("custom-secret", "custom-secret")}
	password, random, err := app.promptPasswordProtection()
	if err != nil {
		t.Fatalf("promptPasswordProtection failed: %v", err)
	}
	if password != "custom-secret" || random {
		t.Fatalf("password = %q, random = %v", password, random)
	}
}
