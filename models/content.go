package models

import "encoding/json"

// SiteHeader represents the shared header.json.
type SiteHeader struct {
	ID              int    `json:"id"`
	SiteName        string `json:"site_name"`
	Language        string `json:"language"`
	HeaderHTML      string `json:"header"`
	CSSFile         string `json:"css_file"`
	TemplateCSSFile string `json:"template_css_file"`
	JSFile          string `json:"js_file"`
	WidgetCode      string `json:"widget_code"`
}

// SiteFooter represents footer.json.
type SiteFooter struct {
	FooterHTML string `json:"footer"`
}

// PageDocument represents the structure shared by home, privacy-policy, site-terms and blog pages.
type PageDocument struct {
	Head                PageHead `json:"head"`
	Body                PageBody `json:"body"`
	CKEnabled           bool     `json:"ck_enabled"`
	FAQ                 []FAQ    `json:"faq"`
	IsDynamicBlogPage   bool     `json:"isDynamicBlogPage"`
	HasWeatherComponent bool     `json:"has_weather_component"`
}

type PageHead struct {
	Title         string            `json:"title"`
	MetaInfo      PageMeta          `json:"meta_info"`
	CSSFile       *string           `json:"css_file"`
	JSFile        *string           `json:"js_file"`
	CanonicalURL  string            `json:"canonical_url"`
	PreloadImages []json.RawMessage `json:"preload_images"`
}

type PageMeta struct {
	PageType        string  `json:"page_type"`
	MetaDescription *string `json:"meta_description"`
	MetaKeywords    *string `json:"meta_keywords"`
	OGTitle         string  `json:"og_title"`
	OGDescription   string  `json:"og_description"`
	OGURL           string  `json:"og_url"`
}

type PageBody struct {
	Content string `json:"content"`
}

type FAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type SiteLinks struct {
	Global []json.RawMessage          `json:"global"`
	Pages  map[string]json.RawMessage `json:"pages"`
}
