## 要结合 proj structure 的设计来说明，因为设计规定了 dao 层一定要返回原错误，所以我们在 service 层才能判断 40001

## 同时描述 transaction 通过 ctx 传递的设计，让 db 不暴露在 service 层
