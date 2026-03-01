package oauth2

import (
	"bytes"
	"context"
	"html/template"

	"github.com/leeseika/cv-demo/pkg/model/dto"
)

type authorizePageContextKey string

const (
	authorizePagePayloadKey authorizePageContextKey = "oauth2_authorize_page_payload"
)

type authorizeInitialProps struct {
	// Locale string `json:"locale"`
	// ReCaptchaEnabled    bool   `json:"recaptcha_enabled"`
	// ReCaptchaSiteKey    string `json:"recaptcha_site_key"`
	ShopLoginURL        string `json:"shop_login_url"`
	AppAuthorizeURL     string `json:"app_authorize_url"`
	API                 string `json:"api"`
	ClientID            string `json:"client_id"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	RedirectURI         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
}

type authorizePageData struct {
	ClientID            string
	CodeChallenge       string
	CodeChallengeMethod string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
}

var authorizePageTmpl = template.Must(template.New("oauth2_app_authorize").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
		<title>Auth</title>
		<link
			rel="stylesheet"
			href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css"
		/>
		<script src="//code.jquery.com/jquery-2.2.4.min.js"></script>
		<script src="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js"></script>
</head>
<body>
		<div class="container">
			<div class="jumbotron">
				<form id="authorize-form" action="/oauth/app/authorize" method="POST">
					<input type="hidden" name="client_id" value="{{ .ClientID }}" />
					<input type="hidden" name="code_challenge" value="{{ .CodeChallenge }}" />
					<input type="hidden" name="code_challenge_method" value="{{ .CodeChallengeMethod }}" />
					<input type="hidden" name="redirect_uri" value="{{ .RedirectURI }}" />
					<input type="hidden" name="response_type" value="{{ .ResponseType }}" />
					<input type="hidden" name="scope" value="{{ .Scope }}" />
					<input type="hidden" name="state" value="{{ .State }}" />
					<input type="hidden" id="csrf-token" name="csrf_token" value="" />

					<h1>Authorize</h1>
					<p>The client would like to perform actions on your behalf.</p>
					<hr />
					<h3>App Information</h3>
					<p id="app-info-tip">Loading app information...</p>
					<ul id="app-info" style="display:none;"></ul>
					<p id="session-tip">Preparing authorization session...</p>
					<p>
						<button
							id="allow-btn"
							type="submit"
							class="btn btn-primary btn-lg"
							style="width:200px;"
							disabled
						>
							Allow
						</button>
					</p>
				</form>
			</div>
		</div>

		<script>
			(function () {
				const form = document.getElementById("authorize-form");
				const allowBtn = document.getElementById("allow-btn");
				const csrfInput = document.getElementById("csrf-token");
				const tip = document.getElementById("session-tip");
				const appInfoTip = document.getElementById("app-info-tip");
				const appInfo = document.getElementById("app-info");

				const payload = {
					client_id: form.querySelector('input[name="client_id"]').value,
					code_challenge: form.querySelector('input[name="code_challenge"]').value,
					code_challenge_method: form.querySelector('input[name="code_challenge_method"]').value,
					redirect_uri: form.querySelector('input[name="redirect_uri"]').value,
					response_type: form.querySelector('input[name="response_type"]').value,
					scope: form.querySelector('input[name="scope"]').value,
					state: form.querySelector('input[name="state"]').value,
				};

				fetch("/apps/" + encodeURIComponent(payload.client_id), {
					method: "GET",
				})
					.then(function (resp) {
						if (!resp.ok) {
							throw new Error("failed to fetch app");
						}
						return resp.json();
					})
					.then(function (app) {
						var appID = app.id || app.ID || "";
						var appName = app.name || app.Name || "";
						var appHomepage = app.home_page_url || app.HomePageURL || "";

						appInfo.innerHTML = "";
						var item1 = document.createElement("li");
						item1.textContent = "ID: " + appID;
						appInfo.appendChild(item1);
						var item2 = document.createElement("li");
						item2.textContent = "Name: " + appName;
						appInfo.appendChild(item2);
						var item3 = document.createElement("li");
						item3.textContent = "Homepage: " + appHomepage;
						appInfo.appendChild(item3);

						appInfo.style.display = "block";
						appInfoTip.textContent = "";
					})
					.catch(function () {
						appInfoTip.textContent = "Failed to load app information.";
					});

				fetch("/oauth/app/create_auth_session", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(payload),
				})
					.then(function (resp) {
						if (!resp.ok) {
							throw new Error("failed to create auth session");
						}
						return resp.json();
					})
					.then(function (data) {
						if (!data || !data.csrf_token) {
							throw new Error("invalid auth session response");
						}
						csrfInput.value = data.csrf_token;
						allowBtn.disabled = false;
						tip.textContent = "Authorization session is ready.";
					})
					.catch(function () {
						tip.textContent = "Failed to prepare authorization session.";
					});
			})();
		</script>
</body>
</html>
`))

func (o *oauth2) AuthorizePage(ctx context.Context, payload dto.OAuthPayload) (string, error) {
	_ = authorizeInitialProps{
		ShopLoginURL:    "",
		AppAuthorizeURL: "",
		API:             "",
	}

	viewData := authorizePageData{
		ClientID:            payload.ClientID,
		CodeChallenge:       payload.CodeChallenge,
		CodeChallengeMethod: payload.CodeChallengeMethod,
		RedirectURI:         payload.RedirectURI,
		ResponseType:        payload.ResponseType,
		Scope:               payload.Scope,
		State:               payload.State,
	}

	var buf bytes.Buffer
	if err := authorizePageTmpl.Execute(&buf, viewData); err != nil {
		return "", err
	}

	return buf.String(), nil
}
