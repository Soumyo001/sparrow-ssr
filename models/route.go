package models

// PageLocation identifies the local files associated with a public page URL.
type PageLocation struct {
	Domain           string `json:"domain"`
	URLPath          string `json:"url_path"`
	ContentDirectory string `json:"content_directory"`
	PageJSON         string `json:"page_json"`
	HeaderJSON       string `json:"header_json"`
	FooterJSON       string `json:"footer_json"`
}
