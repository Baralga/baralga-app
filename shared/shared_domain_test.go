package shared

import (
	"testing"

	"github.com/matryer/is"
)

func TestHasRole(t *testing.T) {
	is := is.New(t)

	p := &Principal{
		Roles: []string{"ROLE_ADMIN"},
	}

	t.Run("has role", func(t *testing.T) {
		hasClaim := p.HasRole("ROLE_ADMIN")
		is.True(hasClaim)
	})

	t.Run("has no role", func(t *testing.T) {
		hasClaim := p.HasRole("ROLE_NOT_HERE")
		is.True(!hasClaim)
	})
}
