package crypto

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

func ReadPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr)

	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	return string(passwordBytes), nil
}

func ReadPasswordTwice(prompt string) (string, error) {
	password1, err := ReadPassword(prompt)
	if err != nil {
		return "", err
	}

	password2, err := ReadPassword("Confirm password: ")
	if err != nil {
		return "", err
	}

	if password1 != password2 {
		return "", fmt.Errorf("passwords do not match")
	}

	return password1, nil
}
