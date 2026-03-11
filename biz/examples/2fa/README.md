# 2FA (两步认证)

## 启动步骤

在仓库根目录执行：

```bash
make 2fa-start
```

程序默认监听 `:9100`，并使用内存 SQLite 作为示例数据库，因此本地直接启动即可。

启动后可以访问：

- `http://localhost:9100/`：默认会重定向到登录页
- `http://localhost:9100/page/login`：登录页面
- `http://localhost:9100/page/register`：注册页面
- `http://localhost:9100/healthz`：健康检查接口

## 可选环境变量

- `TWO_FA_PORT`：自定义服务端口，默认值为 `9100`
- `TWO_FA_DB_DSN`：自定义数据库 DSN，默认值为 `sqlite://file::memory:?cache=shared`

例如：

```bash
TWO_FA_PORT=9200 TWO_FA_DB_DSN='sqlite://tmp/2fa.db' make 2fa-start
```

## 运行截图

① 首先我们访问 `http://localhost:9100/` 重定向到登录页面

![login_page](/assets/2fa/login_page.png)

此时我们还没有注册账号，点击 Go Register 前往注册页面

② 注册一个账号

![register](/assets/2fa/register.png)

点击 Register，跳转到二维码页面

![qr_code](/assets/2fa/qr_code_page.png)

③ 打开手机上的 2FA 类 APP 进行扫码
例如打开 Microsoft Authenticator 扫码后，即可获取实时更新的 TOTP 码

![2fa_app](/assets/2fa/2fa_app.png)

④ 回到登录界面，输入刚刚注册的账号和密码

![login_registered](/assets/2fa/login_registered.png)

点击 Login，进入输入 OTP 页面，打开 Microsoft Authenticator 输入当前显示的 TOTP 码

![2fa_otp](/assets/2fa/2fa_otp.png)

输入 TOTP 码

![input_otp](/assets/2fa/input_otp.png)

⑤ 成功完成登录

![login_succeeded](/assets/2fa/login_succeded.png)
