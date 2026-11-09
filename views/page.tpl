
<!DOCTYPE html>
<html lang="{{.Language}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, initial-scale=1.0">

    <title>{{.Title}}</title>

    {{if .MetaDescription}}
    <meta name="description"
          content="{{.MetaDescription}}">
    {{end}}

    {{if .MetaKeywords}}
    <meta name="keywords"
          content="{{.MetaKeywords}}">
    {{end}}

    {{if .OGTitle}}
    <meta property="og:title"
          content="{{.OGTitle}}">
    {{end}}

    {{if .OGDescription}}
    <meta property="og:description"
          content="{{.OGDescription}}">
    {{end}}

    <link rel="canonical" href="{{.CanonicalURL}}">

    <!-- Shared site stylesheet -->
    {{if .GlobalCSS}}
    <link rel="stylesheet" href="{{.GlobalCSS}}">
    {{end}}

    <!-- Shared template stylesheet -->
    {{if .TemplateCSS}}
    <link rel="stylesheet" href="{{.TemplateCSS}}">
    {{end}}

    <!-- Page-specific stylesheet -->
    {{if .PageCSS}}
    <link rel="stylesheet" href="{{.PageCSS}}">
    {{end}}

    <!-- Local enhancements for the existing search popup -->
    <link rel="stylesheet" href="/static/css/property_search_popup.css">

    <!-- Site widget provided by the publisher -->
    {{.WidgetHTML}}
</head>

<body>
    <!-- Shared site header -->
    {{.HeaderHTML}}

    <!-- Requested page body -->
    {{.BodyHTML}}

    <!-- Shared site footer -->
    {{.FooterHTML}}

    <!-- Shared site JavaScript -->
    {{if .GlobalJS}}
    <script src="{{.GlobalJS}}" defer></script>
    {{end}}

    <!-- Page-specific JavaScript -->
    {{if .PageJS}}
    <script src="{{.PageJS}}" defer></script>
    {{end}}

    <!-- Search popup destination, dates and guests -->
    <script src="/static/js/property_search_popup.js" defer></script>
</body>
</html>
