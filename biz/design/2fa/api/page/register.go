package page

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

var registerPageTmpl = template.Must(template.New("register_page").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Register</title><link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" /></head>
<body><div class="container" style="margin-top:24px;max-width:560px;"><div class="panel panel-default"><div class="panel-heading"><h3 class="panel-title">Register</h3></div><div class="panel-body"><form id="register-form"><div class="form-group"><label>Email</label><input class="form-control" id="email" name="email" type="email" required /></div><div class="form-group"><label>Password</label><input class="form-control" id="password" name="password" type="password" required /></div><p id="error" class="text-danger" style="display:none;"></p><button class="btn btn-primary" type="submit">Register</button> <a class="btn btn-link" href="/page/login">Go Login</a></form></div></div></div><script>(function(){const form=document.getElementById('register-form');const errorEl=document.getElementById('error');form.addEventListener('submit',function(e){e.preventDefault();errorEl.style.display='none';const email=document.getElementById('email').value;const password=document.getElementById('password').value;fetch('/api/auth/register',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({email:email,password:password})}).then(function(resp){return resp.json().then(function(data){return{status:resp.status,data:data};});}).then(function(result){if(result.status<200||result.status>=300){const message=result.data&&result.data.error?result.data.error:'register failed';throw new Error('Request failed (HTTP '+result.status+'): '+message);}if(!result.data||!result.data.qr_code_base64){throw new Error('Request failed: register response is missing qr_code_base64');}sessionStorage.setItem('register_qr',result.data.qr_code_base64);window.location.href='/page/register/qrcode';}).catch(function(err){errorEl.textContent=err&&err.message?err.message:'Request failed: network or server error';errorEl.style.display='block';});});})();</script></body>
</html>`))

var registerQRCodePageTmpl = template.Must(template.New("register_qrcode_page").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Scan QRCode</title><link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" /></head>
<body><div class="container" style="margin-top:24px;max-width:720px;"><div class="jumbotron"><h3>Registration Successful</h3><p>Use Google Authenticator or another TOTP app to scan this QR code.</p><img id="qrcode" style="max-width:320px;display:none;" /><p id="error" class="text-danger" style="display:none;margin-top:12px;"></p><p style="margin-top:16px;"><a class="btn btn-primary" href="/page/login">Go to Login</a></p></div></div><script>(function(){const img=document.getElementById('qrcode');const errorEl=document.getElementById('error');let qr=(sessionStorage.getItem('register_qr')||'').trim();if(!qr){window.location.href='/page/register';return;}if(!/^data:image\//i.test(qr)){qr='data:image/png;base64,'+qr;}if(!/^data:image\//i.test(qr)){errorEl.textContent='invalid qrcode data';errorEl.style.display='block';return;}img.src=qr;img.style.display='block';})();</script></body>
</html>`))

func RegisterPage(c *gin.Context) {
	var buf bytes.Buffer
	_ = registerPageTmpl.Execute(&buf, nil)
	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

func RegisterQRCodePage(c *gin.Context) {
	var buf bytes.Buffer
	err := registerQRCodePageTmpl.Execute(&buf, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to render template")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
