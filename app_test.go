package main

import (
	"testing"

	"github.com/matryer/is"
)

func TestIsProduction(t *testing.T) {
	is := is.New(t)

	a := &app{
		Config: &config{
			Env: "dev",
		},
	}
	is.True(!a.isProduction())
}
