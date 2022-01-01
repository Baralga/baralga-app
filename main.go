package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	app, err := newApp()
	if err != nil {
		return errors.Wrap(err, "newApp")
	}

	err = app.run()
	if err != nil {
		return errors.Wrap(err, "run")
	}

	return nil
}
