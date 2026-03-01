package app

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

var appListPageTmpl = template.Must(template.New("app_list_page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>App List</title>
    <link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" />
</head>
<body>
    <div class="container" style="margin-top: 24px;">
        <div class="jumbotron">
            <h2>Apps</h2>
            <p>Click one app to open its installation URL.</p>
            <p id="tip">Loading app list...</p>
            <ul id="app-list" class="list-group" style="display:none;"></ul>
        </div>
    </div>

    <script>
        (function () {
            const tip = document.getElementById("tip");
            const appList = document.getElementById("app-list");

            fetch("/apps", { method: "GET" })
                .then(function (resp) {
                    if (resp.status === 401) {
                        window.location.href = "/auth/login?next=" + encodeURIComponent("/apps/page");
                        throw new Error("unauthorized");
                    }
                    if (!resp.ok) {
                        throw new Error("failed to fetch apps");
                    }
                    return resp.json();
                })
                .then(function (apps) {
                    if (!Array.isArray(apps) || apps.length === 0) {
                        tip.textContent = "No apps found.";
                        return;
                    }

                    appList.innerHTML = "";
                    apps.forEach(function (app) {
                        var id = app.id || app.ID || "";
                        var name = app.name || app.Name || id;
                        var installationURL = app.installation_url || app.InstallationURL || "";

                        var item = document.createElement("li");
                        item.className = "list-group-item";

                        var btn = document.createElement("button");
                        btn.type = "button";
                        btn.className = "btn btn-link";
                        btn.style.padding = "0";
                        btn.style.fontSize = "16px";
                        btn.textContent = name + " (" + id + ")";

                        if (!installationURL) {
                            btn.disabled = true;
                            btn.title = "Installation URL is empty";
                        } else {
                            btn.addEventListener("click", function () {
                                window.location.href = installationURL;
                            });
                        }

                        item.appendChild(btn);
                        appList.appendChild(item);
                    });

                    appList.style.display = "block";
                    tip.textContent = "";
                })
                .catch(function () {
                    tip.textContent = "Failed to load app list.";
                });
        })();
    </script>
</body>
</html>
`))

func ListPage(c *gin.Context) {
	var buf bytes.Buffer
	if err := appListPageTmpl.Execute(&buf, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render app list page"})
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
