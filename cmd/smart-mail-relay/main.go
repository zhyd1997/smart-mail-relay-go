package main

import (
	"github.com/sirupsen/logrus"

	"smart-mail-relay-go/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		logrus.Fatalf("application error: %v", err)
	}
}
