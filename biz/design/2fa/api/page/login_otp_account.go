package page

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

var loginPageTmpl = template.Must(template.New("login_page").Parse(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Login</title><link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" /></head><body><div class="container" style="margin-top:24px;max-width:560px;"><div class="panel panel-default"><div class="panel-heading"><h3 class="panel-title">Login</h3></div><div class="panel-body"><form id="login-form"><div class="form-group"><label>Email</label><input class="form-control" id="email" name="email" type="text" required /></div><div class="form-group"><label>Password</label><input class="form-control" id="password" name="password" type="password" required /></div><p id="error" class="text-danger" style="display:none;"></p><button class="btn btn-primary" type="submit">Login</button> <a class="btn btn-link" href="/page/register">Go Register</a></form></div></div></div><script>(function(){const form=document.getElementById('login-form');const errorEl=document.getElementById('error');form.addEventListener('submit',function(e){e.preventDefault();errorEl.style.display='none';const email=document.getElementById('email').value;const password=document.getElementById('password').value;fetch('/api/auth/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({email:email,password:password})}).then(function(resp){return resp.json().then(function(data){return{status:resp.status,data:data};});}).then(function(result){if(result.status<200||result.status>=300){const message=result.data&&result.data.error?result.data.error:'login failed';throw new Error('Request failed (HTTP '+result.status+'): '+message);}if(!result.data||!result.data.verification_token){throw new Error('Request failed: login response is missing verification_token');}sessionStorage.setItem('verification_token',result.data.verification_token);window.location.href='/page/otp';}).catch(function(err){errorEl.textContent=err&&err.message?err.message:'Request failed: network or server error';errorEl.style.display='block';});});})();</script></body></html>`))

var otpPageTmpl = template.Must(template.New("otp_page").Parse(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>OTP Verification</title><link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" /></head><body><div class="container" style="margin-top:24px;max-width:560px;"><div class="panel panel-default"><div class="panel-heading"><h3 class="panel-title">Input OTP</h3></div><div class="panel-body"><form id="otp-form"><div class="form-group"><label>OTP</label><input class="form-control" id="otp" name="otp" type="text" required /></div><p id="error" class="text-danger" style="display:none;"></p><button class="btn btn-primary" type="submit">Verify</button></form></div></div></div><script>(function(){const verificationToken=sessionStorage.getItem('verification_token');if(!verificationToken){window.location.href='/page/login';return;}const form=document.getElementById('otp-form');const errorEl=document.getElementById('error');form.addEventListener('submit',function(e){e.preventDefault();errorEl.style.display='none';const otp=document.getElementById('otp').value;fetch('/api/auth/2fa/verify',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({otp:otp,verification_token:verificationToken})}).then(function(resp){return resp.json().then(function(data){return{status:resp.status,data:data};});}).then(function(result){if(result.status<200||result.status>=300){const message=result.data&&result.data.error?result.data.error:'verification failed';throw new Error('Request failed (HTTP '+result.status+'): '+message);}if(!result.data||!result.data.access_token){throw new Error('Request failed: verify response is missing access_token');}sessionStorage.removeItem('verification_token');window.location.href='/page/account#access_token='+encodeURIComponent(result.data.access_token);}).catch(function(err){errorEl.textContent=err&&err.message?err.message:'Request failed: network or server error';errorEl.style.display='block';});});})();</script></body></html>`))

var accountPageTmpl = template.Must(template.New("account_page").Parse(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Account</title><link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" /></head><body><div class="container" style="margin-top:24px;"><div class="jumbotron"><h3>Account Info</h3><p id="tip">Loading account info...</p><pre id="content" style="display:none;"></pre></div></div><script>(function(){const tip=document.getElementById('tip');const content=document.getElementById('content');const hash=new URLSearchParams((window.location.hash||'').replace(/^#/,''));const hashToken=hash.get('access_token');if(hashToken){sessionStorage.setItem('access_token',hashToken);history.replaceState(null,'',window.location.pathname+window.location.search);}const token=sessionStorage.getItem('access_token');if(!token){window.location.href='/page/login';return;}fetch('/api/account/account-info',{method:'GET',headers:{'Authorization':'Bearer '+token}}).then(function(resp){if(resp.status===401){sessionStorage.removeItem('access_token');window.location.href='/page/login';return null;}if(!resp.ok){throw new Error('Request failed (HTTP '+resp.status+'): failed to load account info');}return resp.json();}).then(function(data){if(!data){return;}content.style.display='block';content.textContent=JSON.stringify(data,null,2);tip.textContent='';}).catch(function(err){tip.textContent=err&&err.message?err.message:'Request failed: network or server error';});})();</script></body></html>`))

func LoginPage(c *gin.Context) {
	var buf bytes.Buffer
	_ = loginPageTmpl.Execute(&buf, nil)
	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

func OTPPage(c *gin.Context) {
	var buf bytes.Buffer
	_ = otpPageTmpl.Execute(&buf, nil)
	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

func AccountPage(c *gin.Context) {
	var buf bytes.Buffer
	_ = accountPageTmpl.Execute(&buf, nil)
	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}
