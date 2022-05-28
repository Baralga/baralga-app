package shared

import (
	"testing"

	"github.com/matryer/is"
)

func TestIsProduction(t *testing.T) {
	is := is.New(t)

	a := &App{
		Config: &Config{
			Env: "dev",
		},
	}
	is.True(!a.IsProduction())
}
