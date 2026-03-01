package auth

import (
	"bytes"
	"html/template"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/leeseika/cv-demo/pkg/driver"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

type loginPageData struct {
	Error string
	Next  string
}

var loginPageTmpl = template.Must(template.New("auth_login_page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login</title>
    <link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" />
</head>
<body>
    <div class="container" style="margin-top: 24px; max-width: 520px;">
        <div class="panel panel-default">
            <div class="panel-heading"><h3 class="panel-title">Login</h3></div>
            <div class="panel-body">
                {{ if .Error }}
                <div class="alert alert-danger">{{ .Error }}</div>
                {{ end }}
                <form method="POST" action="/auth/login">
                    <input type="hidden" name="next" value="{{ .Next }}" />
                    <div class="form-group">
                        <label>Email</label>
                        <input class="form-control" type="email" name="email" required />
                    </div>
                    <div class="form-group">
                        <label>Password</label>
                        <input class="form-control" type="password" name="password" required />
                    </div>
                    <button class="btn btn-primary" type="submit">Login</button>
                </form>
            </div>
        </div>
    </div>
</body>
</html>
`))

func LoginPage(c *gin.Context) {
	next := c.Query("next")
	if len(next) == 0 {
		next = "/apps/page"
	}

	var buf bytes.Buffer
	err := loginPageTmpl.Execute(&buf, loginPageData{
		Error: c.Query("error"),
		Next:  next,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render login page"})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

func Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	next := c.PostForm("next")
	if len(next) == 0 {
		next = "/apps/page"
	}

	if len(email) == 0 || len(password) == 0 {
		c.Redirect(http.StatusFound, "/auth/login?error=invalid%20credentials&next="+url.QueryEscape(next))
		return
	}

	var account object.Account
	err := driver.GetDB().Where("email = ? AND password = ?", email, password).First(&account).Error
	if err != nil {
		c.Redirect(http.StatusFound, "/auth/login?error=invalid%20credentials&next="+url.QueryEscape(next))
		return
	}

	token, err := issueLoginToken(jsonmodel.AuthInfo{
		AccountID: account.ID,
		ShopID:    "shop_demo",
		SessionID: "sess_demo",
		Role:      "owner",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue login token"})
		return
	}

	c.SetCookie(authTokenCookieName, token, 3600*24, "/", "", false, true)
	c.Redirect(http.StatusFound, next)
}
