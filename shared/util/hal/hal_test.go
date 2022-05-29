package hal

import (
	"testing"

	"github.com/matryer/is"
)

func TestHrefOfSelf(t *testing.T) {
	is := is.New(t)

	links := NewLinks(NewSelfLink("/api/target-self"))

	selfHref := links.HrefOf("self")

	is.Equal(selfHref, "/api/target-self")
}

func TestHrefOfNotExisting(t *testing.T) {
	is := is.New(t)

	links := NewLinks(NewSelfLink("/api/target-self"))

	selfHref := links.HrefOf("not-here")

	is.Equal(selfHref, "")
}

func TestSize(t *testing.T) {
	is := is.New(t)

	links := NewLinks(NewSelfLink("/api/target-self"))

	size := links.Size()

	is.Equal(size, 1)
}

func TestRelationAndHref(t *testing.T) {
	is := is.New(t)

	link := NewSelfLink("/api/target-self")

	selfRelation := link.Relation()
	selfHref := link.Href()

	is.Equal(selfRelation, "self")
	is.Equal(selfHref, "/api/target-self")
}
