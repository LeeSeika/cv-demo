## CockroachDB Benchmark

这个示例在运行测试时会自动选择数据库后端：

- 如果机器上能启动本地 `cockroach`，默认优先使用本地单节点 `CockroachDB`
- 如果本地 `cockroach` 不可用或启动失败，则自动回退到 `sqlite`
- 如果显式设置 `BATCH_UPSERT_DB_DSN`，则优先使用该 DSN 指向的数据库
- 如果显式设置 `BATCH_UPSERT_FORCE_SQLITE=1`，则强制走 `sqlite`

- 测试文件：`biz/examples/batch-upsert/batch_upsert_test.go`
- 数据库：默认自动探测；也可以通过环境变量显式指定
- 表结构：测试里使用 `AutoMigrate` 创建最小 `product_variants` benchmark 表，保证 `sqlite` / `CockroachDB` 方言都能复用同一套逻辑

包含两类验证：

- 正确性测试：确认 `OnConflict` 和 “delete + insert” 两种实现都能在当前选中的数据库后端下完成批量 upsert
- 基准测试：先构建 `1000000` 行存量数据，再比较两种实现对随机 `20` 行已有记录做 upsert 的性能

默认测试运行方式：

```bash
go test ./biz/examples/batch-upsert -run TestProductVariantService_UpsertVariants
```

默认行为：

- 优先尝试自动启动本地单节点 `CockroachDB`
- 若自动启动失败，则回退到 `sqlite`
- 可以通过测试日志看见实际选中的后端

强制走 `sqlite`：

```bash
BATCH_UPSERT_FORCE_SQLITE=1 \
go test ./biz/examples/batch-upsert -run TestProductVariantService_UpsertVariants
```

显式使用 `CockroachDB`：

- 如果你希望跑固定的本地或远端 `CockroachDB` 实例，设置 `BATCH_UPSERT_DB_DSN`
- 下面这组结果来自显式指定的本地 `3` 节点 `CockroachDB` 集群

本地 `3` 节点基准集群：

- 节点数：`3`
- 版本：`CockroachDB v24.1.2`
- SQL 地址：`127.0.0.1:27257`, `127.0.0.1:27258`, `127.0.0.1:27259`
- 基准命令使用的 DSN：`postgresql://root@127.0.0.1:27257/defaultdb?sslmode=disable`

运行方式：

```bash
BATCH_UPSERT_DB_DSN='postgresql://root@127.0.0.1:27257/defaultdb?sslmode=disable' \
GOCACHE=/tmp/cv-demo-go-build \
go test ./biz/examples/batch-upsert -run TestProductVariantService_UpsertVariants

BATCH_UPSERT_DB_DSN='postgresql://root@127.0.0.1:27257/defaultdb?sslmode=disable' \
GOCACHE=/tmp/cv-demo-go-build \
go test ./biz/examples/batch-upsert -run '^$' -bench BenchmarkProductVariantService_UpsertVariants -benchmem
```

一次本地实测结果（`darwin/arm64`, `Apple M2`, `3-node CockroachDB`）：

| 策略              | 初始表行数 | 单次更新行数 |   ns/op |  B/op | allocs/op |
| ----------------- | ---------: | -----------: | ------: | ----: | --------: |
| `OnConflict`      |  1,000,000 |           20 | 1297857125 | 105144 |       704 |
| `Delete + Insert` |  1,000,000 |           20 | 44339995792 | 111888 |       795 |

从这组结果看：

- 这个 benchmark 更接近“存量大表上的小批量增量更新”
- 两种策略处理的是同一类 workload：每次随机挑选 20 条已存在记录并更新
- 在这组数据下，`OnConflict` 比 `Delete + Insert` 快很多，内存分配也更低
- benchmark 只统计 upsert 语句本身，预先灌入 `1000000` 行基础数据的时间不计入 `ns/op`

benchmark 输出字段说明：

- `BenchmarkProductVariantService_UpsertVariants/...`：benchmark 函数名和子用例名
- `seed_1000000`：开始计时前，表中预先灌入 `1000000` 行基础数据
- `update_20`：每次被计时的 SQL 操作更新 `20` 条已存在记录
- `-8`：运行 benchmark 时的 `GOMAXPROCS=8`
- `1`：testing 框架实际执行了 `1` 轮被计时操作；在 CockroachDB 这个 workload 下单次操作非常重，框架没有继续放大 `b.N`
- `ns/op`：每轮操作的平均耗时；这里的一轮操作是“一次 20 行 upsert”
- `B/op`：每轮操作平均发生的 Go 堆内存分配字节数
- `allocs/op`：每轮操作平均发生的 Go 堆分配次数
- `ok ... 203.373s`：整个 benchmark 命令的总墙钟时间，包含建库、建表、预灌数据、预热和正式测量，不等于 `ns/op`

按这组数据换算：

- 耗时：`OnConflict` 相比 `Delete + Insert` 缩短约 `97.1%`
- 相对倍数：`OnConflict` 约快 `34.2x`
- 内存分配字节数：`OnConflict` 降低约 `6.0%`
- 内存分配次数：`OnConflict` 降低约 `11.4%`

可以把这个结果理解为：

- 当表里已经有 `1000000` 行存量数据时，更新 `20` 条已有记录，`OnConflict` 仍然明显优于“先删后插”
- 在 `CockroachDB` 上，这组差异被进一步放大，多语句事务形式的“先删后插”成本明显更高
- 这组差异主要来自 SQL 策略本身，而不是初始化建表或灌数成本

测试数据设计：

- benchmark 启动时先插入 `1000000` 条固定 `id` 的 `product_variants`
- 基准阶段每次随机挑选 `20` 条已存在记录做 upsert
- 只改变 `title`、`price`、`updated_at`
- 保持 `sku`、`selected_options`、`product_id`、`created_at` 不变

这样可以让两种实现的业务结果保持一致，比较重点落在 SQL 策略本身：

- `UpsertVariants_OnConflict`：单条批量 `INSERT ... ON CONFLICT DO UPDATE`
- `UpsertVariants_ReplaceInto`：事务内先 `DELETE WHERE id IN ?`，再 `INSERT`
