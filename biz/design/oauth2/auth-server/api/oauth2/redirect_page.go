package oauth2

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

type redirectPageData struct {
	Code  string
	State string
	Error string
}

var redirectPageTmpl = template.Must(template.New("oauth2_redirect_page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>OAuth2 Redirect Result</title>
    <link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" />
</head>
<body>
    <div class="container" style="margin-top: 24px;">
        <div class="jumbotron">
            <h1>OAuth2 Redirect</h1>
            {{ if .Error }}
            <p class="text-danger">error: {{ .Error }}</p>
            {{ else }}
            <p class="text-success">authorization completed.</p>
            {{ end }}
            <p><strong>code:</strong> {{ .Code }}</p>
            <p><strong>state:</strong> {{ .State }}</p>
        </div>
    </div>
</body>
</html>
`))

func RedirectPage(c *gin.Context) {
	data := redirectPageData{
		Code:  c.Query("code"),
		State: c.Query("state"),
		Error: c.Query("error"),
	}

	var buf bytes.Buffer
	if err := redirectPageTmpl.Execute(&buf, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render redirect page"})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
