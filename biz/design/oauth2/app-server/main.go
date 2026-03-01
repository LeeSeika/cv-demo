package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	authServerURL = flag.String("auth", "http://localhost:3360", "Auth Server URL")
	resourceServerURL = flag.String("resource", "http://localhost:3370", "Resource Server URL")
	redirectURL   = flag.String("redirectURL", "http://localhost:9094/oauth2", "Redirect URL")
	clientID      = flag.String("client", "app_l0g6nfq800002c86amnkbcaa", "Client ID")
	secret        = flag.String("secret", "mock_secret", "Secret")
)

var (
	globalToken *oauth2.Token // Non-concurrent security
)

func main() {
	flag.Parse()

	log.Printf("Testing OAuth2 APP Authorize Server: %s", *authServerURL)
	log.Printf("Client ID: %s", *clientID)
	log.Printf("Secret : %s", *secret)

	config := oauth2.Config{
		ClientID:     *clientID,
		ClientSecret: *secret,
		Scopes:       []string{"all"},
		RedirectURL:  *redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  *authServerURL + "/oauth/app/authorize",
			TokenURL: *authServerURL + "/oauth/app/token",
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u := config.AuthCodeURL("xyz",
			oauth2.SetAuthURLParam("code_challenge", genCodeChallengeS256("s256example")),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"))
		http.Redirect(w, r, u, http.StatusFound)
	})

	http.HandleFunc("/oauth2", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		state := r.Form.Get("state")
		if state != "xyz" {
			http.Error(w, "State invalid", http.StatusBadRequest)
			return
		}
		code := r.Form.Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}
		token, err := config.Exchange(context.Background(), code, oauth2.SetAuthURLParam("code_verifier", "s256example"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		globalToken = token

		successPage := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<title>Authorization Success</title>
	<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" />
</head>
<body>
	<div class="container" style="margin-top:24px; max-width:720px;">
		<div class="alert alert-success" role="alert">
			<h3 style="margin-top:0;">OAuth Authorization Succeeded</h3>
			<p>Authorization completed successfully. Redirecting to order page in 5 seconds...</p>
			<p><a href="/orders" class="btn btn-success">Go to Orders Now</a></p>
		</div>
	</div>
	<script>
		setTimeout(function () {
			window.location.href = "/orders";
		}, 5000);
	</script>
</body>
</html>`

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(successPage))
	})

	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		if globalToken == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		globalToken.Expiry = time.Now()
		token, err := config.TokenSource(context.Background(), globalToken).Token()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		globalToken = token
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(token)
	})

	http.HandleFunc("/try", func(w http.ResponseWriter, r *http.Request) {
		if globalToken == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		resp, err := http.Get(fmt.Sprintf("%s/test?access_token=%s", *authServerURL, globalToken.AccessToken))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()

		io.Copy(w, resp.Body)
	})

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if globalToken == nil || len(globalToken.AccessToken) == 0 {
			loginURL := *authServerURL + "/auth/login?next=" + url.QueryEscape("/apps/page")
			unauthorizedPage := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<title>401 Unauthorized</title>
	<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" />
</head>
<body>
	<div class="container" style="margin-top:24px; max-width:720px;">
		<div class="alert alert-warning" role="alert">
			<h3 style="margin-top:0;">401 Unauthorized</h3>
			<p>OAuth access token is missing. Please login on auth-server and complete OAuth authorization.</p>
			<p><a href="` + loginURL + `" class="btn btn-primary">Go to Auth-Server Login</a></p>
		</div>
	</div>
</body>
</html>`
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(unauthorizedPage))
			return
		}

		page := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<title>Orders</title>
	<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" />
</head>
<body>
	<div class="container" style="margin-top:24px;">
		<div class="jumbotron">
			<h2>Order List</h2>
			<p id="tip">Loading orders...</p>
			<p id="shop"></p>
			<table id="orders-table" class="table table-striped" style="display:none;">
				<thead><tr><th>ID</th><th>Order No</th><th>Status</th><th>Total</th></tr></thead>
				<tbody></tbody>
			</table>
		</div>
	</div>
	<script>
		(function () {
			const tip = document.getElementById("tip");
			const shop = document.getElementById("shop");
			const table = document.getElementById("orders-table");
			const tbody = table.querySelector("tbody");

			fetch("/orders/data", { method: "GET" })
				.then(function (resp) {
					if (resp.status === 401) {
						tip.textContent = "Unauthorized. Please login on auth-server first.";
						return null;
					}
					if (!resp.ok) {
						throw new Error("failed to fetch orders");
					}
					return resp.json();
				})
				.then(function (data) {
					if (!data) {
						return;
					}

					const orders = Array.isArray(data.orders) ? data.orders : [];
					shop.textContent = "Shop: " + (data.shop_id || "");

					if (orders.length === 0) {
						tip.textContent = "No orders found.";
						return;
					}

					tbody.innerHTML = "";
					orders.forEach(function (o) {
						const tr = document.createElement("tr");
						const id = o.id || o.ID || "";
						const orderNo = o.order_no || o.OrderNo || "";
						const status = o.status || o.Status || "";
						const total = o.total_amount || o.TotalAmount || 0;
						tr.innerHTML = "<td>" + id + "</td><td>" + orderNo + "</td><td>" + status + "</td><td>" + total + "</td>";
						tbody.appendChild(tr);
					});

					table.style.display = "table";
					tip.textContent = "";
				})
				.catch(function () {
					tip.textContent = "Failed to load orders.";
				});
		})();
	</script>
</body>
</html>`

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	})

	http.HandleFunc("/orders/data", func(w http.ResponseWriter, r *http.Request) {
		if globalToken == nil || len(globalToken.AccessToken) == 0 {
			http.Error(w, "oauth access token missing", http.StatusUnauthorized)
			return
		}

		authToken, err := r.Cookie("auth_token")
		if err != nil || len(authToken.Value) == 0 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, *resourceServerURL+"/api/orders", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", "Bearer "+authToken.Value)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})

	http.HandleFunc("/pwd", func(w http.ResponseWriter, r *http.Request) {
		token, err := config.PasswordCredentialsToken(context.Background(), "test", "test")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		globalToken = token
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(token)
	})

	http.HandleFunc("/client", func(w http.ResponseWriter, r *http.Request) {
		cfg := clientcredentials.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			TokenURL:     config.Endpoint.TokenURL,
		}

		token, err := cfg.Token(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(token)
	})

	log.Println("Client is running at 9094 port.Please open http://localhost:9094")
	log.Fatal(http.ListenAndServe(":9094", nil))
}

func genCodeChallengeS256(s string) string {
	s256 := sha256.Sum256([]byte(s))
	return base64.URLEncoding.EncodeToString(s256[:])
}
