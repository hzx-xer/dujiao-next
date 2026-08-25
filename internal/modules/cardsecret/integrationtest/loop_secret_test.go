package integrationtest

import (
	"errors"
	"testing"
	"time"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	productgormstore "github.com/dujiao-next/internal/modules/catalog/product/store/gormstore"

	"github.com/dujiao-next/internal/constants"
	cardsecretapp "github.com/dujiao-next/internal/modules/cardsecret/application"
	cardsecretdomain "github.com/dujiao-next/internal/modules/cardsecret/domain"
	cardsecretgormstore "github.com/dujiao-next/internal/modules/cardsecret/infrastructure/gormstore"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/shopspring/decimal"
)

func newLoopTestProductAndSKU(t *testing.T, slug string) (*productdomain.Product, *productdomain.ProductSKU) {
	t.Helper()
	product := &productdomain.Product{
		CategoryID:      1,
		Slug:            slug,
		TitleJSON:       jsonmap.JSON{"zh-CN": "循环卡密商品"},
		PriceAmount:     money.FromDecimal(decimal.NewFromInt(10)),
		PurchaseType:    constants.ProductPurchaseMember,
		FulfillmentType: constants.FulfillmentTypeAuto,
		IsActive:        true,
	}
	return product, &productdomain.ProductSKU{
		SKUCode:     productdomain.DefaultSKUCode,
		PriceAmount: money.FromDecimal(decimal.NewFromInt(10)),
		IsActive:    true,
	}
}

// TestLoopCardSecretRecycledOnMarkUsed 循环卡密标记已用后应立刻自动变回可售。
func TestLoopCardSecretRecycledOnMarkUsed(t *testing.T) {
	db := setupCardSecretServiceTestDB(t)
	product, sku := newLoopTestProductAndSKU(t, "loop-card-recycled-on-mark-used")
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	sku.ProductID = product.ID
	if err := db.Create(sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}

	svc := NewCardSecretService(
		cardsecretgormstore.New(db),
		cardsecretgormstore.NewBatch(db),
		productgormstore.NewProductStore(db),
		productgormstore.NewSKUStore(db),
	)

	_, created, err := svc.CreateCardSecretBatch(CreateCardSecretBatchInput{
		ProductID: product.ID,
		Secrets:   []string{"LOOP-SECRET-001"},
		BatchNo:   "LOOP-BATCH-001",
		Source:    constants.CardSecretSourceManual,
		IsLoop:    true,
	})
	if err != nil {
		t.Fatalf("create loop batch failed: %v", err)
	}
	if created != 1 {
		t.Fatalf("want created=1 got %d", created)
	}

	items, _, err := svc.ListCardSecrets(ListCardSecretInput{ProductID: product.ID, Page: 1, PageSize: 10})
	if err != nil || len(items) != 1 {
		t.Fatalf("list secrets failed: err=%v len=%d", err, len(items))
	}
	secretID := items[0].ID

	store := cardsecretgormstore.New(db)
	affected, err := store.MarkUsed([]uint{secretID}, 999, time.Now())
	if err != nil {
		t.Fatalf("mark used failed: %v", err)
	}
	if affected != 1 {
		t.Fatalf("want 1 row affected by MarkUsed, got %d", affected)
	}

	var refreshed cardsecretdomain.Secret
	if err := db.First(&refreshed, secretID).Error; err != nil {
		t.Fatalf("reload secret failed: %v", err)
	}
	if refreshed.Status != cardsecretdomain.StatusAvailable {
		t.Fatalf("want status=available after recycle, got %s", refreshed.Status)
	}
	if refreshed.OrderID != nil || refreshed.UsedAt != nil || refreshed.ReservedAt != nil {
		t.Fatalf("want order_id/used_at/reserved_at cleared after recycle, got order_id=%v used_at=%v reserved_at=%v",
			refreshed.OrderID, refreshed.UsedAt, refreshed.ReservedAt)
	}
}

// TestEnablingLoopModeDisablesAvailableNonLoopSecrets 开启循环模式应立刻清空该
// 商品+SKU 下还能卖的普通卡密，两者不能同时存在。
func TestEnablingLoopModeDisablesAvailableNonLoopSecrets(t *testing.T) {
	db := setupCardSecretServiceTestDB(t)
	product, sku := newLoopTestProductAndSKU(t, "enable-loop-disables-non-loop")
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	sku.ProductID = product.ID
	if err := db.Create(sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}

	svc := NewCardSecretService(
		cardsecretgormstore.New(db),
		cardsecretgormstore.NewBatch(db),
		productgormstore.NewProductStore(db),
		productgormstore.NewSKUStore(db),
	)

	if _, _, err := svc.CreateCardSecretBatch(CreateCardSecretBatchInput{
		ProductID: product.ID,
		Secrets:   []string{"NORMAL-001", "NORMAL-002"},
		BatchNo:   "NORMAL-BATCH",
		Source:    constants.CardSecretSourceManual,
	}); err != nil {
		t.Fatalf("create normal batch failed: %v", err)
	}

	if _, _, err := svc.CreateCardSecretBatch(CreateCardSecretBatchInput{
		ProductID: product.ID,
		Secrets:   []string{"LOOP-SECRET-001"},
		BatchNo:   "LOOP-BATCH",
		Source:    constants.CardSecretSourceManual,
		IsLoop:    true,
	}); err != nil {
		t.Fatalf("create loop batch failed: %v", err)
	}

	items, total, err := svc.ListCardSecrets(ListCardSecretInput{ProductID: product.ID, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list secrets failed: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].Secret != "LOOP-SECRET-001" {
		t.Fatalf("want only the loop secret to remain visible, got total=%d items=%+v", total, items)
	}

	var deletedCount int64
	if err := db.Unscoped().Model(&cardsecretdomain.Secret{}).
		Where("product_id = ? AND is_loop = ? AND deleted_at IS NOT NULL", product.ID, false).
		Count(&deletedCount).Error; err != nil {
		t.Fatalf("count soft-deleted secrets failed: %v", err)
	}
	if deletedCount != 2 {
		t.Fatalf("want the 2 previous normal secrets soft-deleted, got %d", deletedCount)
	}
}

// TestCreateNonLoopBatchRejectedWhileInLoopMode 循环模式下不能再加普通卡密，
// 必须先删除循环卡密才能切回普通模式。
func TestCreateNonLoopBatchRejectedWhileInLoopMode(t *testing.T) {
	db := setupCardSecretServiceTestDB(t)
	product, sku := newLoopTestProductAndSKU(t, "reject-normal-batch-in-loop-mode")
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	sku.ProductID = product.ID
	if err := db.Create(sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}

	svc := NewCardSecretService(
		cardsecretgormstore.New(db),
		cardsecretgormstore.NewBatch(db),
		productgormstore.NewProductStore(db),
		productgormstore.NewSKUStore(db),
	)

	if _, _, err := svc.CreateCardSecretBatch(CreateCardSecretBatchInput{
		ProductID: product.ID,
		Secrets:   []string{"LOOP-SECRET-001"},
		BatchNo:   "LOOP-BATCH",
		Source:    constants.CardSecretSourceManual,
		IsLoop:    true,
	}); err != nil {
		t.Fatalf("create loop batch failed: %v", err)
	}

	_, _, err := svc.CreateCardSecretBatch(CreateCardSecretBatchInput{
		ProductID: product.ID,
		Secrets:   []string{"NORMAL-001"},
		BatchNo:   "NORMAL-BATCH",
		Source:    constants.CardSecretSourceManual,
	})
	if !errors.Is(err, cardsecretapp.ErrLoopModeConflict) {
		t.Fatalf("want ErrLoopModeConflict, got %v", err)
	}
}
