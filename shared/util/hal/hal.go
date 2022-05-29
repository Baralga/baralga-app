package hal

const relationSelf = "self"

type (
	Links map[string]*Link

	Link struct {
		Href string `json:"href"`
	}
)

func (l *Links) HrefOf(rel string) string {
	for r, link := range *l {
		if r == rel {
			return link.Href
		}
	}
	return ""
}

func (l *Links) Size() int {
	i := 0
	for range *l {
		i++
	}
	return i
}

func (l *Links) Href() string {
	for _, v := range *l {
		return v.Href
	}
	return ""
}

func (l *Links) Relation() string {
	for k := range *l {
		return k
	}
	return ""
}

func NewLink(relation, href string) *Links {
	link := make(Links, 1)
	link[relation] = &Link{Href: href}
	return &link
}

func NewSelfLink(href string) *Links {
	return NewLink(relationSelf, href)
}

func NewLinks(links ...*Links) *Links {
	ls := make(Links, len(links))
	for _, l := range links {
		ls[l.Relation()] = &Link{Href: l.Href()}
	}
	return &ls
}
