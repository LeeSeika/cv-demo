package aggregation

import (
	"context"
	"fmt"
	"time"

	imageSvc "github.com/leeseika/cv-demo/biz/design/renderer/caching/service/image"
	productSvc "github.com/leeseika/cv-demo/biz/design/renderer/caching/service/product"
	referenceSvc "github.com/leeseika/cv-demo/biz/design/renderer/caching/service/reference"
	shopSvc "github.com/leeseika/cv-demo/biz/design/renderer/caching/service/shop"
	kvcache "github.com/leeseika/cv-demo/pkg/driver/kv-cache"
	"github.com/leeseika/cv-demo/pkg/errs"
	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	"github.com/rs/zerolog/log"
)

func (a *aggregation) GetProductDetail(ctx context.Context, productID string) (*dto.ProductDetail, error) {
	// try to get product detail from cache
	productDetail, err := a.aggregationCacheDAO.GetProductDetail(ctx, productID)
	if err == nil {
		return productDetail, nil
	}

	if errs.IsKVCacheError(err, kvcache.ErrKeyNotFound) {
		log.Warn().Err(err).Msg("empty value placeholder hit for product detail")
		return nil, err
	}
	// log the error if it's not a cache miss
	if !errs.IsKVCacheError(err, kvcache.ErrKeyCacheMissed) {
		log.Err(err).Msg("failed to get product detail from id")
	}

	resultCh := a.singleFlightGroup.DoChan(fmt.Sprintf("product_detail:%s", productID), func() (interface{}, error) {
		cached, cacheErr := a.aggregationCacheDAO.GetProductDetail(ctx, productID)
		if cacheErr == nil {
			return cached, nil
		}
		if errs.IsKVCacheError(cacheErr, kvcache.ErrKeyNotFound) {
			return nil, cacheErr
		}

		product, queryErr := productSvc.Get().GetProductByID(ctx, productID)
		if queryErr != nil {
			return nil, queryErr
		}
		shop, queryErr := shopSvc.Get().GetByID(ctx, product.ShopID)
		if queryErr != nil {
			return nil, queryErr
		}
		productRef, queryErr := referenceSvc.Get().GetProductRef(ctx, productID)
		if queryErr != nil {
			return nil, queryErr
		}

		variantIDs := productRef.VariantIDs
		productVariantMap, queryErr := productSvc.Get().GetProductVariantsByIDs(ctx, variantIDs)
		if queryErr != nil {
			return nil, queryErr
		}
		variantRefs, queryErr := referenceSvc.Get().GetMultiProductVariantRefs(ctx, variantIDs)
		if queryErr != nil {
			return nil, queryErr
		}
		imageIDs := append([]string{}, productRef.ImageIDs...)
		for _, variantRef := range variantRefs {
			imageIDs = append(imageIDs, variantRef.ImageIDs...)
		}
		imageMap, queryErr := imageSvc.Get().GetImagesByIDs(ctx, imageIDs)
		if queryErr != nil {
			return nil, queryErr
		}

		variantDetails := make([]*dto.ProductVariantDetail, 0, len(variantIDs))
		for _, variantID := range variantIDs {
			variantRef := variantRefs[variantID]
			variantImages := make([]*cache.Image, 0, len(variantRef.ImageIDs))
			for _, imageID := range variantRef.ImageIDs {
				variantImages = append(variantImages, imageMap[imageID])
			}
			variant, ok := productVariantMap[variantID]
			if !ok || variant == nil {
				continue
			}
			variantDetails = append(variantDetails, dto.BuildProductVariantDetail(
				*variant,
				product.Title,
				shop.CurrencyCode,
				variantImages,
			))
		}

		productImages := make([]*cache.Image, 0, len(productRef.ImageIDs))
		for _, imageID := range productRef.ImageIDs {
			productImages = append(productImages, imageMap[imageID])
		}

		rebuilt := dto.BuildProductDetail(
			*product,
			variantDetails,
			productImages,
		)

		setErr := a.aggregationCacheDAO.SetProductDetail(ctx, productID, rebuilt)
		if setErr != nil {
			log.Err(setErr).Msg("failed to set product detail cache")
		}

		return rebuilt, nil
	})

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timeout.C:
		return nil, context.DeadlineExceeded
	case result := <-resultCh:
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Val.(*dto.ProductDetail), nil
	}
}
