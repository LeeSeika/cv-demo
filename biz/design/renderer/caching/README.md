# Renderer 缓存设计

![renderer 缓存架构](/assets/renderer-caching/renderer_cache_arch.png)

## 1. 缓存模型

存在三种缓存模型，分别为`基础缓存模型`、`关系缓存模型`、`聚合缓存模型`。

### 1.1. 基础缓存模型

基础缓存模型满足以下要求：
① 存在唯一一个与其对应的数据库模型
② 模型所需的所有数据均来自该数据库模型
③ 使用 Cache Aside 策略实现缓存更新

基础缓存模型的意义：`通过高速缓存加快访问数据库数据的速度`
基础缓存模型的设计理念：`数据库模型有什么，缓存模型就有什么`

以 Product 模型为例：

```go
// pkg/model/cache/product.go
package cache

// Product 缓存模型
type Product struct {
	ID          string                     `json:"id"`
	ShopID      string                     `json:"shop_id"`
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Options     []*jsonmodel.ProductOption `json:"options"`
	UpdatedAt   time.Time                  `json:"updated_at"`
	CreatedAt   time.Time                  `json:"created_at"`
}

// ProductFromObject 构建 Product 缓存模型
func ProductFromObject(productObj *object.Product) *Product {
	return &Product{
		ID:          productObj.ID,
		ShopID:      productObj.ShopID,
		Title:       productObj.Title,
		Description: productObj.Description,
		Options:     productObj.Options,
		UpdatedAt:   productObj.UpdatedAt,
		CreatedAt:   productObj.CreatedAt,
	}
}
```

```go
// pkg/model/object/product.go
package object

// Product 数据库模型
type Product struct {
	ID          string `gorm:"size:128;primarykey"`
	ShopID      string `gorm:"size:128"`
	Title       string `gorm:"size:128;index"`
	Description string `gorm:"size:10240"`
	Options     datatype.JSONSlice[*jsonmodel.ProductOption]
	CreatedAt   time.Time `gorm:"index"`
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	// references
	TemplateID      sql.NullString   `gorm:"size:128;index;constraint:OnUpdate:CASCADE;constraint:OnDelete:SET NULL"`
	Template        Template         `gorm:"foreignkey:TemplateID"`                            // The template associated with this product
	ProductVariants []ProductVariant `gorm:"foreignkey:ProductID;constraint:OnDelete:CASCADE"` // The product variants associated with this product
}
```

### 1.2. 关系缓存模型

关系缓存模型的意义：`描述各种缓存模型之间的联系`

关系缓存模型用于描述基础数据模型之间的关系，它只保存基础模型之间的索引（如 id），并不保存具体的业务数据。
例如，Product 模型与其他模型有以如下关系：
① Product 拥有多个变体（Product Variant）
② Product 可以绑定多个图片（Image）
③ Product 属于一个店铺（Shop）
据此，我们可以定义出如下关系模型：

```go
// pkg/model/cache/product.go
package cache

// ProductReference 缓存模型
type ProductReference struct {
	ID         string   `json:"id"`
	ImageIDs   []string `json:"image_ids"`
	VariantIDs []string `json:"variant_ids"`
}

// BuildProductReference 构建 ProductReference 缓存模型
func BuildProductReference(
	id string,
	imageIDs []string,
	variantIDs []string,
) *ProductReference {
	return &ProductReference{
		ID:         id,
		ImageIDs:   imageIDs,
		VariantIDs: variantIDs,
	}
}
```

除此之外，我们可能还会存在一种特殊的关系模型，例如 Catalog（目录）与 Product 有如下关系：
① 一个 Catalog 可以拥有多个 Product
② Catalog 可以调整 Product 的顺序，例如按照创建时间、Product 名字等方式进行排序
对于这种情况，我们不再使用 Key-Value 形式的缓存索引方式，而是使用 Redis 提供 Sorted Set 实现索引和存储

对于缓存更新策略，如果是 ProductReference 形式的关系模型，我们继续使用 Cache Aside 策略；对于 Catalog 与 Product 之间的关系模型，我们选择让数据常驻在 Redis 内存中，通过 ZSet 相关命令实现缓存更新。

### 1.3. 聚合缓存模型

聚合缓存模型满足以下要求：
① 模型由其他基础缓存模型、聚合缓存模型组成
② 内部包含的模型之间的联系由关系缓存模型描述
③ 仅通过设置过期时间来实现缓存淘汰、更新

聚合缓存模型的设计理念：`API 接口响应需要什么，缓存模型就有什么`

以 GetProductDetails API 返回的 ProductDetail 为例：

```go
// pkg/model/dto/product_detail.go
package dto

// ProductDetail 缓存模型
type ProductDetail struct {
	cache.Product  // 内嵌
	Variants []*ProductVariantDetail `json:"variants"`
	Images   []*cache.Image          `json:"images"`
}

// BuildProductDetail 构建 ProductDetail 缓存模型
func BuildProductDetail(
	product cache.Product,
	variants []*ProductVariantDetail,
	images []*cache.Image,
) *ProductDetail {
	return &ProductDetail{
		Product:  product,
		Variants: variants,
		Images:   images,
	}
}
```

## 2. 缓存管理

### 2.1. 缓存管理服务

我们给缓存模型划分成了三类，同时 service 层也需要三类服务实现对应缓存模型的管理。

#### 2.1.1. 基础缓存模型管理服务

- **依赖**
  基础缓存模型管理服务只允许依赖对应数据库、缓存模型的 DAO 实例，不允许依赖其他模型的 DAO，也不允许 Join、Preload 等跨表访问操作。
  例子：

  ```mermaid
    classDiagram
    class ProductService{
        +cacheDAO ProductCacheDAO
        +dao ProductDAO
    }
    class ImageService{
        +cacheDAO ImageCacheDAO
        +dao ImageDAO
    }
  ```

- **职责**
  仅负责缓存模型的查询、构建、淘汰等操作。

#### 2.1.2. 关系缓存模型管理服务

- **依赖**
  因为关系缓存模型需要查询模型之间的联系，所以允许它访问所有数据库模型的 DAO 实例，以及所有 reference 缓存模型。
  例子：

  ```mermaid
  classDiagram
  class ReferenceService{
      +productCacheDAO ProductCacheDAO
      +productDAO ProductDAO
  	  +imageCacheDAO ImageCacheDAO
      +imageDAO ImageDAO
  }
  ```

- **职责**
  负责构建、维护、淘汰各种缓存模型之间的联系。

#### 2.1.3. 聚合缓存模型管理服务

- **依赖**
  聚合缓存模型管理服务在 DAO 层只需要依赖对应的 Aggregation Cache DAO，它不会直接依赖任何数据库模型的 DAO，相反，它通过依赖各种基础缓存模型管理服务实现基础缓存模型的查询。
  例子：

  ```mermaid
  classDiagram
  class AggregationService{
      +cacheDAO AggregationCacheDAO
      +productSvc ProductService
  	  +imageSvc ImageService
  	  +shopSvc ShopService
  }
  ```

- **职责**
  负责构建、查询 API 需要的所有数据，实现`缓存前置`。

### 2.2. Cache Aside

[Cache Aside](https://www.geeksforgeeks.org/system-design/cache-aside-pattern/)（旁路缓存）是一种常见的缓存更新策略，非常适合在`读多写少`的场景下使用，接下来将描述项目对 Cache Aside 策略的具体实现。

#### 2.2.1. 发布数据库更新通知

当数据库数据发生变动时（通常来自 Admin 端），我们需要异步地通知 Renderer 端的缓存管理服务删除对应的缓存模型。
通常我们使用消息队列作为中间件实现异步通知，消息模型是`pub/sub`模型，下面以`Google Pub/Sub`为中间件的代码例子如下：

```go
package product

func(p *product) UpdateProduct(c context.Context, req *dto.UpdateProductReq) error {
	err := p.dao.Update(xxx)
	if err != nil {
		return err
	}

	event := dto.BuildUpdateProductEvent(xxx)
	p.mqClient.SendEvent(c, event)

	return nil
}
```

#### 2.2.2. 接收通知并删除相关缓存

在负责管理缓存模型的 Renderer 端，我们通过 Subscriber 订阅`Google Pub/Sub`的 Topic 消费感兴趣的事件。通过配置 DeadLetter Policy 和选择性响应 ACK 实现`退避重试`和`死信队列`。

```go
package pubsub

func(ps *productSubscriber) Start(ctx context.Context) error {
	host := "localhost"
	port := 8085
	gpsTopic := "product"
	projectID := "cv-demo"
	subscriberID := "product-subscriber-node-1"

	client, err := gcpPubSub.NewClient(ctx, projectID,
		gcpOption.WithEndpoint(fmt.Sprintf("%s:%d", host, port)),
		gcpOption.WithoutAuthentication(),
		gcpOption.WithGRPCDialOption(grpc.WithInsecure()),
	)
	if err != nil {
		return err
	}
	topic, err := client.CreateTopic(ctx, gpsTopic)
	if err != nil {
		return err
	}
	productSvc := productsvc.Get()

	sub, err := client.CreateSubscription(ctx, subscriberID, gcpPubSub.SubscriptionConfig{
		Topic: topic,
		// TODO: dead letter policy
		// TODO: backoff retry policy
	})
	if err != nil {
		return err
	}

	fn := func() {
		err = sub.Receive(ctx, func(ctx context.Context, msg *gcpPubSub.Message) {
			// 只在成功或不需要重试的情况下响应 ACK，需要重试的情况响应 Nack
			shouldAck := true
			defer func() {
				if shouldAck {
					msg.Ack()
				} else {
					msg.Nack()
				}
			}

			var event dto.Event
			err := json.Unmarshal(msg.Data, &event)
			if err != nil {
				return
			}

			if event.Action != "update" || event.Action != "delete" {
				return
			}

			p := event.Payload
			var payload dto.ProductPayload
			err = json.Unmarshal(p, &payload)
			if err != nil {
				return
			}

			id := payload.ID
			err = productSvc.DeleteCacheByIDs(ctx, []string{id})
			if err != nil {
				// 响应 Nack，需要重试
				shouldAck = false
				return err
			}

			return nil
		})
		if err != nil {
			log.Err(err).Msg("failed to subscribe pub/sub message")
		}
	}

	threading.GoSafe(fn, "panic happened in product subscriber", nil)

	return nil
}

```

#### 2.2.3. 避免缓存穿透

在 Cache Aside 策略下，当请求的数据在缓存层和数据库都不存在时，我们必然需要查询一次数据库来得知这个结果，当有大量请求去查询不存在的数据时，同样也会给数据库带来巨大的压力。
为了让请求在不需要查询数据库的情况下就能判断数据是否存在，通常我们有`布隆过滤器`和`缓存层存储空值`两种实现方案。
在本项目中，我们通过缓存空值的方案避免缓存穿透，以查询 Product 为例子，实现方案如下：

首先我们定义了一个缓存层错误`ErrKeyNotFound`和一个字符串常量`__EMPTY VALUE PLACEHOLDER__`：

```go
// pkg/driver/kv-cache/errors.go
package kvcache

var (
	// ErrKeyNotFound indicates that the specified key does not exist in the KV store and any other storage layer.
	ErrKeyNotFound = errors.New("key not found")
)
```

```go
// pkg/driver/kv-cache/consts.go
package kvcache

var emptyValuePlaceholder = []byte("__EMPTY VALUE PLACEHOLDER__")
```

在`KV Cache Driver`层（类似于访问数据库的 gorm，用于屏蔽不同缓存中间件客户端 SDK 的差异）对缓存查询结果进行判断，当发现查询返回的 Value 是空值占位符常量时，返回`ErrKeyNotFound`。我们以 Redis Cache Provider 的实现代码为例子：

```go
// pkg/driver/kv-cache/provider_redis.go
package kvcache

func (p *RedisProvider) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := p.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare(val, emptyValuePlaceholder) == 1 {
		return nil, ErrKeyNotFound
	}
	return val, nil
}
```

在缓存 DAO 层中，我们直接抛出 Driver 层返回的错误，不做封装：

```go
// biz/design/renderer/caching/dao/cache/product/get_by_id.go
package product

func (p *product) GetByID(ctx context.Context, productID string) (*cache.Product, error) {
	key := productKey(productID)

	b, err := p.kvCache.Get(ctx, key)
	if err != nil {
		// 直接抛出 Driver 层返回的错误
		return nil, err
	}

	var productCache cache.Product
	err = json.Unmarshal(b, &productCache)
	if err != nil {
		return nil, err
	}

	return &productCache, nil
}
```

在 Service 层中，就可以通过判断 DAO 返回的 error 是否为`ErrKeyNotFound`，来判断是否命中了空值占位符：

```go
// biz/design/renderer/caching/service/product/get_product_by_id.go
package product

func (p *product) GetProductByID(ctx context.Context, productID string) (*cache.Product, error) {
	productCache, err := p.productCacheDAO.GetByID(ctx, productID)
	if err == nil {
		return productCache, nil
	}
	if errs.IsKVCacheError(err, kvcache.ErrKeyNotFound) {
		return nil, errors.New("product not found")
	}
}
```

#### 2.2.4. SingleFlight

在 Cache Aside 策略下，当请求的数据在缓存层不存在时，我们需要前往数据库进行查询，如果有大量请求在查询缓存层时发生了 miss，我们通过 SingleFlight 实现在同一个服务实例下，只有一个请求真正去到数据库执行查询，其他请求只需等待数据库查询结果即可。
以查询 Product 为例子，代码如下：

```go
// biz/design/renderer/caching/service/product/get_product_by_id.go
package product

func (p *product) GetProductByID(ctx context.Context, productID string) (*cache.Product, error) {
	productCache, err := p.productCacheDAO.GetByID(ctx, productID)
	if err == nil {
		return productCache, nil
	}
	if errs.IsKVCacheError(err, kvcache.ErrKeyNotFound) {
		return nil, err
	}

	resultCh := p.singleFlightGroup.DoChan(fmt.Sprintf("product:%s", productID), func() (interface{}, error) {
		cached, cacheErr := p.productCacheDAO.GetByID(ctx, productID)
		if cacheErr == nil {
			return cached, nil
		}
		if errs.IsKVCacheError(cacheErr, kvcache.ErrKeyNotFound) {
			return nil, cacheErr
		}

		productObj, dbErr := p.productDAO.GetByID(ctx, productID)
		if dbErr != nil {
			if errs.IsDBError(dbErr, gorm.ErrRecordNotFound) {
				// 发生了缓存穿透，在缓存层设置空值占位符
				p.productCacheDAO.SetNil(ctx, productID)
			}
			return nil, dbErr
		}

		rebuilt := cache.ProductFromObject(productObj)
		setErr := p.productCacheDAO.Set(ctx, productID, rebuilt)
		if setErr != nil {
			log.Err(setErr).Msg("failed to set product cache")
		}

		return rebuilt, nil
	})

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timeout.C:
		// 超时控制，避免大量请求阻塞等待
		return nil, context.DeadlineExceeded
	case result := <-resultCh:
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Val.(*cache.Product), nil
	}
}

```
