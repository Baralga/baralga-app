package shared

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestMustPrincipalFromContext(t *testing.T) {
	is := is.New(t)

	t.Run("context without principal", func(t *testing.T) {
		// turn off panic
		defer func() { _ = recover() }()

		MustPrincipalFromContext(context.Background())

		// fail if no panic
		t.Errorf("did not panic")
	})

	t.Run("context with principal", func(t *testing.T) {
		ctx := ToContextWithPrincipal(context.Background(), &Principal{Name: "john"})

		p := MustPrincipalFromContext(ctx)

		is.True(p != nil)
		is.Equal(p.Name, "john")
	})
}

func TestToContextWithPrincipal(t *testing.T) {
	is := is.New(t)

	t.Run("new context with principal", func(t *testing.T) {
		ctx := ToContextWithPrincipal(context.Background(), &Principal{Name: "john"})

		_, ok := ctx.Value(contextKeyPrincipal).(*Principal)

		is.True(ok)
	})
}

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
