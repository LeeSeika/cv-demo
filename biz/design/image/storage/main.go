package main

import (
	_ "embed"
	"os"

	"github.com/gin-gonic/gin"
	imageapi "github.com/leeseika/cv-demo/biz/design/image/storage/api"
	imagesvc "github.com/leeseika/cv-demo/biz/design/image/storage/service"
	"github.com/leeseika/cv-demo/pkg/config"
	"github.com/leeseika/cv-demo/pkg/driver"
	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/model/object"
)

//go:embed web/index.html
var indexHTML []byte

func main() {
	initDB()
	initStorage()
	imagesvc.Init()

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", indexHTML)
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	apiGroup := r.Group("/api")
	apiGroup.Use(mockAuthInfoMiddleware())
	imageapi.RegisterRoutes(apiGroup)

	port := os.Getenv("IMAGE_STORAGE_PORT")
	if len(port) == 0 {
		port = "2020"
	}

	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}

func initDB() {
	dsn := os.Getenv("IMAGE_STORAGE_DB_DSN")
	if len(dsn) == 0 {
		dsn = "sqlite://file::memory:?cache=shared"
	}

	driver.InitDB(config.DB{DSN: dsn})
	db := driver.GetDB()
	if err := db.AutoMigrate(&object.Image{}); err != nil {
		panic("failed to migrate image tables: " + err.Error())
	}
}

func initStorage() {
	storageType := os.Getenv("IMAGE_STORAGE_TYPE")
	if len(storageType) == 0 {
		storageType = driver.S3Client
	}

	storageConf := config.StorageConfig{
		OSSType: storageType,
		AssetsBuilderConfig: config.AssetsBuilderConfig{
			AssetsDomain:    getEnv("IMAGE_STORAGE_ASSETS_DOMAIN", "http://127.0.0.1:4000"),
			AssetsURLFormat: getEnv("IMAGE_STORAGE_ASSETS_URL_FORMAT", "{{ .Domain }}/{{ .Bucket }}/{{ .Source }}"),
		},
		S3: config.S3Config{
			IsAuth:          false,
			Bucket:          getEnv("IMAGE_STORAGE_S3_BUCKET", "cv-demo"),
			AccessKeyID:     getEnv("IMAGE_STORAGE_S3_ACCESS_KEY_ID", "demo-access-key"),
			SecretAccessKey: getEnv("IMAGE_STORAGE_S3_SECRET_ACCESS_KEY", "demo-secret-key"),
			Region:          getEnv("IMAGE_STORAGE_S3_REGION", "us-east-1"),
			Endpoint:        getEnv("IMAGE_STORAGE_S3_ENDPOINT", "127.0.0.1:4000"),
			URLScheme:       getEnv("IMAGE_STORAGE_S3_URL_SCHEME", "http"),
			UsePathStyle:    true,
		},
		GCS: config.GCSConfig{
			IsAuth:    false,
			Bucket:    getEnv("IMAGE_STORAGE_GCS_BUCKET", "cv-demo"),
			URLScheme: getEnv("IMAGE_STORAGE_GCS_URL_SCHEME", "http"),
			Host:      getEnv("IMAGE_STORAGE_GCS_HOST", "127.0.0.1"),
			Port:      getEnv("IMAGE_STORAGE_GCS_PORT", "4001"),
		},
	}

	driver.InitStorageProvider(storageConf)
	driver.InitStorageURLBuilder(storageConf)
}

func mockAuthInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		shopID := c.GetHeader("X-Shop-ID")
		if len(shopID) == 0 {
			shopID = c.Query("shop_id")
		}
		if len(shopID) == 0 {
			shopID = "demo-shop"
		}

		c.Set("auth_info", &jsonmodel.AuthInfo{
			ShopID:    shopID,
			AccountID: getEnv("IMAGE_STORAGE_DEMO_ACCOUNT_ID", "demo-account"),
			Role:      "admin",
		})
		c.Next()
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if len(v) == 0 {
		return fallback
	}
	return v
}
