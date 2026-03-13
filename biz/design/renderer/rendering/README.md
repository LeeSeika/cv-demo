## 业务数据初步注入的逻辑，set_legacy_product，参考 shopify 代码

## json_template.go 的 ToProps，把 JSON Template 转换为 map[string]any （value 是 liquid 类型），可用于 liquid 的渲染

## render json template：按照 template 的 order 顺序，逐一找到每个 comp settings 的 wafer.template（由数据库保存的 component wafer 得到，wafer 也就是 .liquid 文件中的类 html 语法的 liquid 代码）

## TODO: template test 代码 preprocess 后，得到一些依然有 translate 前缀 t: 的内容，render 阶段应该实现 i18N（如果不好设计就阉割掉这部分）
