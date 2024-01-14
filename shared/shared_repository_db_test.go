package shared

import (
	"context"
	"testing"
)

func TestMustTxFromContext(t *testing.T) {

	t.Run("context without principal", func(t *testing.T) {
		// turn off panic
		defer func() { _ = recover() }()

		MustTxFromContext(context.Background())

		// fail if no panic
		t.Errorf("did not panic")
	})

}
