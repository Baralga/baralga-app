package hx

import (
	"bytes"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestDelete(t *testing.T) {
	// Arrange
	is := is.New(t)
	var b bytes.Buffer

	// Act
	node := Delete("my-url")

	// Assert
	err := node.Render(&b)
	is.NoErr(err)
	html := b.String()
	is.True(strings.Contains(html, "hx-delete=\"my-url\""))
}

func TestConfirm(t *testing.T) {
	// Arrange
	is := is.New(t)
	var b bytes.Buffer

	// Act
	node := Confirm("my-url")

	// Assert
	err := node.Render(&b)
	is.NoErr(err)
	html := b.String()
	is.True(strings.Contains(html, "hx-confirm=\"my-url\""))
}
