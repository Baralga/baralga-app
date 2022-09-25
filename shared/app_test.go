package shared

import (
	"testing"

	"github.com/matryer/is"
)

func TestIsProduction(t *testing.T) {
	is := is.New(t)

	config := &Config{
		Env: "dev",
	}
	is.True(!config.IsProduction())
}
