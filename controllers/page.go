package controllers

import (
	"html/template"
	"log"
	"net/http"

	"sparrow-ssr/services"

	beego "github.com/beego/beego/v2/server/web"
)

type pageDependencies struct {
	catalog *services.Catalog
	loader  *services.ContentLoader
}

// Initialized once before Beego starts accepting requests.
var dependencies pageDependencies

func SetDependencies(
	catalog *services.Catalog,
	loader *services.ContentLoader,
) {
	dependencies = pageDependencies{
		catalog: catalog,
		loader:  loader,
	}
}

type PageController struct {
	beego.Controller
}

func (c *PageController) Get() {
	if dependencies.catalog == nil || dependencies.loader == nil {
		c.writeError(
			http.StatusInternalServerError,
			"Application is not initialized",
		)
		return
	}

	// Resolve the requested URL.
	location, found := dependencies.catalog.Resolve(
		c.Ctx.Request.Host,
		c.Ctx.Request.URL.Path,
	)

	if !found {
		c.writeError(
			http.StatusNotFound,
			"Page not found",
		)
		return
	}

	// Load the corresponding JSON documents.
	content, err := dependencies.loader.Load(location)
	if err != nil {
		log.Printf(
			"Content loading failed: %v", err,
		)

		c.writeError(
			http.StatusInternalServerError,
			"Failed to load page content",
		)
		return
	}

	// Prepare template data.
	title := content.Page.Head.Title
	if title == "" {
		title = content.Header.SiteName
	}

	language := content.Header.Language
	if language == "" {
		language = "en"
	}

	c.Data["Title"] = title
	c.Data["Language"] = language

	c.Data["CanonicalURL"] =
		content.Page.Head.CanonicalURL

	c.Data["MetaDescription"] = nullableString(
		content.Page.Head.MetaInfo.MetaDescription,
	)

	c.Data["MetaKeywords"] = nullableString(
		content.Page.Head.MetaInfo.MetaKeywords,
	)

	c.Data["OGTitle"] =
		content.Page.Head.MetaInfo.OGTitle

	c.Data["OGDescription"] =
		content.Page.Head.MetaInfo.OGDescription

	// Shared and page-specific asset URLs.
	c.Data["GlobalCSS"] = content.Header.CSSFile
	c.Data["TemplateCSS"] =
		content.Header.TemplateCSSFile

	c.Data["PageCSS"] = nullableString(
		content.Page.Head.CSSFile,
	)

	c.Data["GlobalJS"] = content.Header.JSFile

	c.Data["PageJS"] = nullableString(
		content.Page.Head.JSFile,
	)

	// These fragments originate from the trusted site-publishing pipeline, not visitor input.
	c.Data["HeaderHTML"] = template.HTML(
		content.Header.HeaderHTML,
	)

	c.Data["BodyHTML"] = template.HTML(
		content.Page.Body.Content,
	)

	c.Data["FooterHTML"] = template.HTML(
		content.Footer.FooterHTML,
	)

	c.Data["WidgetHTML"] = template.HTML(
		content.Header.WidgetCode,
	)

	// Render the Beego HTML template.
	c.TplName = "page.tpl"

	if err := c.Render(); err != nil {
		log.Printf(
			"HTML rendering failed: %v", err,
		)

		c.writeError(
			http.StatusInternalServerError,
			"Failed to render page",
		)
		return
	}

	// Render was already called explicitly.
	// Prevent Beego from rendering twice.
	c.EnableRender = false
}

func (c *PageController) writeError(
	status int,
	message string,
) {
	c.EnableRender = false
	c.Ctx.Output.SetStatus(status)

	_ = c.Ctx.Output.Body([]byte(message))
}

func nullableString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
