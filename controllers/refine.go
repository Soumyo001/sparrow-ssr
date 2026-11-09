package controllers

import (
	"html/template"
	"log"
	"net/http"

	beego "github.com/beego/beego/v2/server/web"
)

type RefineController struct {
	beego.Controller
}

func (c *RefineController) Get() {
	if dependencies.catalog == nil ||
		dependencies.loader == nil {
		c.writeRefineError(
			http.StatusInternalServerError,
			"Application is not initialized",
		)
		return
	}

	// Verify the actual request host against the
	// deployment catalog. The homepage provides
	// the site's storage-relative file paths.
	location, found := dependencies.catalog.Resolve(
		"test-travel-templates.beta.123presto.com",
		"/",
	)
	if !found {
		c.writeRefineError(
			http.StatusNotFound,
			"Site not found",
		)
		return
	}

	// Reuse the existing loader for the shared
	// header, footer and site configuration.
	content, err := dependencies.loader.Load(location)
	if err != nil {
		log.Printf("Refine page content loading failed: %v", err)
		c.writeRefineError(
			http.StatusInternalServerError,
			"Failed to load site content",
		)
		return
	}

	language := content.Header.Language
	if language == "" {
		language = "en"
	}

	c.Data["Language"] = language
	c.Data["SiteName"] = content.Header.SiteName

	// This virtual application route is available
	// only for a registered hostname.
	c.Data["CanonicalURL"] =
		"https://" + location.Domain + "/refine"

	// Preserve the publisher's asset URLs.
	c.Data["GlobalCSS"] = content.Header.CSSFile
	c.Data["TemplateCSS"] = content.Header.TemplateCSSFile
	c.Data["GlobalJS"] = content.Header.JSFile

	// Trusted publisher-managed HTML fragments.
	c.Data["HeaderHTML"] = template.HTML(
		content.Header.HeaderHTML,
	)
	c.Data["FooterHTML"] = template.HTML(
		content.Footer.FooterHTML,
	)
	c.Data["WidgetHTML"] = template.HTML(
		content.Header.WidgetCode,
	)

	c.TplName = "refine.tpl"

	if err := c.Render(); err != nil {
		log.Printf("Refine template rendering failed: %v", err)
		c.writeRefineError(
			http.StatusInternalServerError,
			"Failed to render search page",
		)
		return
	}

	c.EnableRender = false
}

func (c *RefineController) writeRefineError(
	status int,
	message string,
) {
	c.EnableRender = false
	c.Ctx.Output.SetStatus(status)
	_ = c.Ctx.Output.Body([]byte(message))
}
