package hx

import (
	"bytes"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestBoost(t *testing.T) {
	// Arrange
	is := is.New(t)
	var b bytes.Buffer

	// Act
	node := Boost()

	// Assert
	err := node.Render(&b)
	is.NoErr(err)
	html := b.String()
	is.True(strings.Contains(html, "hx-boost=\"true\""))
}

func TestPushURLTrue(t *testing.T) {
	// Arrange
	is := is.New(t)
	var b bytes.Buffer

	// Act
	node := PushURLTrue()

	// Assert
	err := node.Render(&b)
	is.NoErr(err)
	html := b.String()
	is.True(strings.Contains(html, "hx-push-url=\"true\""))
}

func TestPushURL(t *testing.T) {
	// Arrange
	is := is.New(t)
	var b bytes.Buffer

	// Act
	node := PushURL("my-url")

	// Assert
	err := node.Render(&b)
	is.NoErr(err)
	html := b.String()
	is.True(strings.Contains(html, "hx-push-url=\"my-url\""))
}

func TestPost(t *testing.T) {
	// Arrange
	is := is.New(t)
	var b bytes.Buffer

	// Act
	node := Post("my-url")

	// Assert
	err := node.Render(&b)
	is.NoErr(err)
	html := b.String()
	is.True(strings.Contains(html, "hx-post=\"my-url\""))
}

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

func TestGet(t *testing.T) {
	// Arrange
	is := is.New(t)
	var b bytes.Buffer

	// Act
	node := Get("my-url")

	// Assert
	err := node.Render(&b)
	is.NoErr(err)
	html := b.String()
	is.True(strings.Contains(html, "hx-get=\"my-url\""))
}

func TestTarget(t *testing.T) {
	// Arrange
	is := is.New(t)
	var b bytes.Buffer

	// Act
	node := Target("my-url")

	// Assert
	err := node.Render(&b)
	is.NoErr(err)
	html := b.String()
	is.True(strings.Contains(html, "hx-target=\"my-url\""))
}

func TestSwap(t *testing.T) {
	// Arrange
	is := is.New(t)
	var b bytes.Buffer

	// Act
	node := Swap("my-url")

	// Assert
	err := node.Render(&b)
	is.NoErr(err)
	html := b.String()
	is.True(strings.Contains(html, "hx-swap=\"my-url\""))
}

func TestTrigger(t *testing.T) {
	// Arrange
	is := is.New(t)
	var b bytes.Buffer

	// Act
	node := Trigger("my-url")

	// Assert
	err := node.Render(&b)
	is.NoErr(err)
	html := b.String()
	is.True(strings.Contains(html, "hx-trigger=\"my-url\""))
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