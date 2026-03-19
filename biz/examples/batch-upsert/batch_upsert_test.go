package batchupsert

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	jsonmodel "github.com/leeseika/cv-demo/pkg/model/json"
	"github.com/leeseika/cv-demo/pkg/model/object"
	"github.com/leeseika/cv-demo/pkg/utils/datatype"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	benchmarkSeedRowCount    = 1_000_000
	benchmarkUpdateBatchSize = 20
	benchmarkInsertBatchSize = 1_000
)

const (
	batchUpsertDBDSNEnv        = "BATCH_UPSERT_DB_DSN"
	batchUpsertForceSQLiteEnv  = "BATCH_UPSERT_FORCE_SQLITE"
	localCockroachStartupLimit = 20 * time.Second
)

var autoCockroachCluster localCockroachCluster

type upsertStrategy struct {
	name string
	run  func(context.Context, *ProductVariantService, []object.ProductVariant) error
}

type benchmarkProductVariantRow struct {
	ID              string `gorm:"primaryKey"`
	Price           int    `gorm:"not null;default:0"`
	SKU             string
	Title           string `gorm:"not null"`
	SelectedOptions datatype.JSONSlice[*jsonmodel.SelectedOption]
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ProductID       string `gorm:"not null"`
}

func (benchmarkProductVariantRow) TableName() string {
	return "product_variants"
}

func TestMain(m *testing.M) {
	code := m.Run()
	autoCockroachCluster.Stop()
	os.Exit(code)
}

func TestProductVariantService_UpsertVariants(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	strategies := []upsertStrategy{
		{
			name: "on_conflict",
			run: func(ctx context.Context, svc *ProductVariantService, variants []object.ProductVariant) error {
				return svc.UpsertVariants_OnConflict(ctx, variants)
			},
		},
		{
			name: "delete_then_insert",
			run: func(ctx context.Context, svc *ProductVariantService, variants []object.ProductVariant) error {
				return svc.UpsertVariants_ReplaceInto(ctx, variants)
			},
		},
	}

	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			db := openBenchmarkDB(t)
			service := NewProductVariantService(db)

			base := makeBenchmarkVariants(4, 0)
			require.NoError(t, db.Create(&base).Error)

			updates := makeBenchmarkVariants(4, 1)
			require.NoError(t, strategy.run(ctx, service, updates))

			var variants []object.ProductVariant
			require.NoError(t, db.Order("id ASC").Find(&variants).Error)
			require.Len(t, variants, len(base))

			for i, variant := range variants {
				expected := updates[i]
				require.Equal(t, expected.ID, variant.ID)
				require.Equal(t, expected.Title, variant.Title)
				require.Equal(t, expected.Price, variant.Price)
				require.True(t, expected.UpdatedAt.Equal(variant.UpdatedAt))
				require.Equal(t, base[i].SKU, variant.SKU)
				require.Equal(t, base[i].ProductID, variant.ProductID)
				require.True(t, base[i].CreatedAt.Equal(variant.CreatedAt))
				require.Equal(t, base[i].SelectedOptions.String(), variant.SelectedOptions.String())
			}
		})
	}
}

func BenchmarkProductVariantService_UpsertVariants(b *testing.B) {
	ctx := context.Background()
	strategies := []upsertStrategy{
		{
			name: "on_conflict",
			run: func(ctx context.Context, svc *ProductVariantService, variants []object.ProductVariant) error {
				return svc.UpsertVariants_OnConflict(ctx, variants)
			},
		},
		{
			name: "delete_then_insert",
			run: func(ctx context.Context, svc *ProductVariantService, variants []object.ProductVariant) error {
				return svc.UpsertVariants_ReplaceInto(ctx, variants)
			},
		},
	}

	base := makeBenchmarkVariants(benchmarkSeedRowCount, 0)

	for _, strategy := range strategies {
		b.Run(fmt.Sprintf("%s/seed_%d/update_%d", strategy.name, benchmarkSeedRowCount, benchmarkUpdateBatchSize), func(b *testing.B) {
			db := openBenchmarkDB(b)
			service := NewProductVariantService(db)
			require.NoError(b, db.CreateInBatches(&base, benchmarkInsertBatchSize).Error)

			rng := rand.New(rand.NewSource(42))

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				workload := makeRandomUpdateBatch(base, i+1, rng, benchmarkUpdateBatchSize)
				b.StartTimer()

				require.NoError(b, strategy.run(ctx, service, workload))
			}
		})
	}
}

func openBenchmarkDB(tb testing.TB) *gorm.DB {
	tb.Helper()

	dsn := strings.TrimSpace(os.Getenv(batchUpsertDBDSNEnv))
	if dsn != "" {
		db, err := openBenchmarkPostgresDB(tb, dsn)
		require.NoError(tb, err)
		return db
	}

	if os.Getenv(batchUpsertForceSQLiteEnv) != "1" {
		if dsn, ok := autoCockroachCluster.Ensure(tb); ok {
			db, err := openBenchmarkPostgresDB(tb, dsn)
			if err == nil {
				tb.Logf("using auto-started CockroachDB cluster: %s", dsn)
				return db
			}
			tb.Logf("auto-started CockroachDB cluster unavailable, falling back to sqlite: %v", err)
		} else {
			tb.Log("CockroachDB unavailable, falling back to sqlite")
		}
	}

	return openBenchmarkSQLiteDB(tb)
}

func openBenchmarkSQLiteDB(tb testing.TB) *gorm.DB {
	tb.Helper()

	dsn := filepath.Join(tb.TempDir(), "batch_upsert.sqlite")

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(tb, err)

	sqlDB, err := db.DB()
	require.NoError(tb, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	tb.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(tb, db.AutoMigrate(&benchmarkProductVariantRow{}))

	return db
}

func openBenchmarkPostgresDB(tb testing.TB, adminDSN string) (*gorm.DB, error) {
	tb.Helper()

	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	dbName := uniqueBenchmarkDatabaseName(tb.Name())
	if err := adminDB.Exec(`CREATE DATABASE "` + dbName + `"`).Error; err != nil {
		return nil, err
	}

	dsn := replaceDatabaseInPostgresDSN(tb, adminDSN, dbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	adminSQLDB, err := adminDB.DB()
	if err != nil {
		return nil, err
	}

	tb.Cleanup(func() {
		_ = sqlDB.Close()
		_ = adminDB.Exec(`DROP DATABASE IF EXISTS "` + dbName + `" CASCADE`).Error
		_ = adminSQLDB.Close()
	})

	if err := db.AutoMigrate(&benchmarkProductVariantRow{}); err != nil {
		return nil, err
	}

	return db, nil
}

func makeBenchmarkVariants(batchSize int, version int) []object.ProductVariant {
	variants := make([]object.ProductVariant, 0, batchSize)

	for i := 0; i < batchSize; i++ {
		createdAt := time.Unix(1_700_000_000+int64(i), 0).UTC()
		updatedAt := createdAt.Add(time.Duration(version) * time.Minute)

		variants = append(variants, object.ProductVariant{
			ID:    fmt.Sprintf("variant-%06d", i),
			Price: 1000 + version + i,
			SKU:   fmt.Sprintf("SKU-%06d", i),
			Title: fmt.Sprintf("Variant %06d v%d", i, version),
			SelectedOptions: datatype.NewJSONSlice([]*jsonmodel.SelectedOption{
				{Name: "color", Value: fmt.Sprintf("color-%02d", i%8)},
				{Name: "size", Value: fmt.Sprintf("size-%02d", i%5)},
			}),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			ProductID: fmt.Sprintf("product-%04d", i%32),
		})
	}

	return variants
}

func makeRandomUpdateBatch(base []object.ProductVariant, version int, rng *rand.Rand, batchSize int) []object.ProductVariant {
	if batchSize > len(base) {
		panic("batch size exceeds base row count")
	}

	selected := make(map[int]struct{}, batchSize)
	updates := make([]object.ProductVariant, 0, batchSize)

	for len(updates) < batchSize {
		idx := rng.Intn(len(base))
		if _, exists := selected[idx]; exists {
			continue
		}
		selected[idx] = struct{}{}

		source := base[idx]
		source.Title = fmt.Sprintf("%s update-%06d", source.Title, version)
		source.Price = source.Price + version
		source.UpdatedAt = source.CreatedAt.Add(time.Duration(version) * time.Second)

		updates = append(updates, source)
	}

	return updates
}

func uniqueBenchmarkDatabaseName(name string) string {
	var b strings.Builder
	b.WriteString("batch_upsert_")
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	b.WriteString(fmt.Sprintf("_%d", time.Now().UnixNano()))
	return b.String()
}

func replaceDatabaseInPostgresDSN(tb testing.TB, dsn string, dbName string) string {
	tb.Helper()

	parsed, err := url.Parse(dsn)
	require.NoError(tb, err)

	parsed.Path = "/" + dbName

	return parsed.String()
}

type localCockroachCluster struct {
	mu        sync.Mutex
	attempted bool
	available bool
	dsn       string
	baseDir   string
	procs     []*exec.Cmd
	logFiles  []*os.File
}

func (c *localCockroachCluster) Ensure(tb testing.TB) (string, bool) {
	tb.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.attempted {
		return c.dsn, c.available
	}
	c.attempted = true

	cockroachPath, err := exec.LookPath("cockroach")
	if err != nil {
		return "", false
	}

	baseDir, err := os.MkdirTemp("", "batch-upsert-crdb-*")
	if err != nil {
		return "", false
	}
	c.baseDir = baseDir

	rpcPort, err := getFreeTCPPort()
	if err != nil {
		c.stopLocked()
		return "", false
	}
	httpPort, err := getFreeTCPPort()
	if err != nil {
		c.stopLocked()
		return "", false
	}

	if err := c.startSingleNode(cockroachPath, baseDir, rpcPort, httpPort); err != nil {
		c.stopLocked()
		return "", false
	}

	dsn := fmt.Sprintf("postgresql://root@127.0.0.1:%d/defaultdb?sslmode=disable", rpcPort)
	if err := c.waitForSQLReady(dsn); err != nil {
		c.stopLocked()
		return "", false
	}

	c.available = true
	c.dsn = dsn
	return c.dsn, true
}

func (c *localCockroachCluster) startSingleNode(
	cockroachPath string,
	baseDir string,
	rpcPort int,
	httpPort int,
) error {
	storeDir := filepath.Join(baseDir, "node1")
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return err
	}

	logFile, err := os.Create(filepath.Join(baseDir, "node1.log"))
	if err != nil {
		return err
	}

	cmd := exec.Command(
		cockroachPath,
		"start-single-node",
		"--insecure",
		fmt.Sprintf("--store=%s", storeDir),
		fmt.Sprintf("--listen-addr=127.0.0.1:%d", rpcPort),
		fmt.Sprintf("--http-addr=127.0.0.1:%d", httpPort),
	)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return err
	}

	c.procs = append(c.procs, cmd)
	c.logFiles = append(c.logFiles, logFile)
	return nil
}

func (c *localCockroachCluster) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stopLocked()
}

func (c *localCockroachCluster) stopLocked() {
	for _, cmd := range c.procs {
		if cmd.Process == nil {
			continue
		}
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}
	for _, logFile := range c.logFiles {
		_ = logFile.Close()
	}
	if c.baseDir != "" {
		_ = os.RemoveAll(c.baseDir)
	}

	c.procs = nil
	c.logFiles = nil
	c.baseDir = ""
	c.available = false
	c.dsn = ""
}

func (c *localCockroachCluster) waitForSQLReady(dsn string) error {
	deadline := time.Now().Add(localCockroachStartupLimit)
	for time.Now().Before(deadline) {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err == nil {
			var ready int
			queryErr := db.Raw("SELECT 1").Scan(&ready).Error
			if sqlDB, dbErr := db.DB(); dbErr == nil {
				_ = sqlDB.Close()
			}
			if queryErr == nil && ready == 1 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for local CockroachDB SQL readiness")
}

func getFreeTCPPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = listener.Close()
	}()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("unexpected listener addr type %T", listener.Addr())
	}

	return addr.Port, nil
}
