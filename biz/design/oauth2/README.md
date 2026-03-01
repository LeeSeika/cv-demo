流程：

1. 安装 APP
   ① 查找平台支持的 APP （list APP 表）
   ② 选择 APP，点击 APP
   访问 installationURL，installationURL 相当于 9094
   app 的 client_id 和 client_secret 会被用上
   ③ 重定向到 auth-api 的授权页面（GET oauth/app/authorize），点击授权
   点击后会再重定向到 9094/oauth2，oauth2 代理向 auth-api 请求得到 access_token（POST oauth/app/authorize）
   client 将 access_token 保存到自己的后端
   同时 auth-api 也将 access_token、refresh_token 存储在 etcd
2. 使用 APP
   ① 浏览器访问 9094/api，app 代理向 resource server 请求数据
3. 刷新 token
   ① 当 APP 代理请求 resource server 时，发现 access_token 过期了，app 代理向 auth-api 请求刷新 token（POST oauth/app/refresh）

本地运行授权页面渲染示例（auth-server）：

```bash
go run ./biz/design/oauth2/auth-server/cmd/authorize-page
```

默认监听 `:9095`，可访问：
`http://localhost:9095/oauth/app/authorize?client_id=app_demo&code_challenge=abc&code_challenge_method=S256&redirect_uri=http://localhost/callback&response_type=code&scope=read_products&state=xyz`
