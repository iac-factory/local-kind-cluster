package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"authentication/internal/name"
)

func hash(password string) (string, error) {
	bytes, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if e != nil {
		slog.Error("Unable to Hash Password", slog.Group(name.Name, slog.String("error", e.Error())))
		return "", e
	}

	return string(bytes), nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Password: ")
	text, e := reader.ReadString('\n')
	if e != nil {
		panic(e)
	}

	input := strings.TrimSpace(text)

	output, e := hash(input)
	if e != nil {
		panic(e)
	}

	fmt.Println(output)
}
