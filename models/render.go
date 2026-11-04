package models

// PageContent contains the decoded JSON documents needed to render one website page.
type PageContent struct {
	Header SiteHeader
	Page   PageDocument
	Footer SiteFooter
}
