package reference

import (
	"context"
	"fmt"
	"time"

	kvcache "github.com/leeseika/cv-demo/pkg/driver/kv-cache"
	"github.com/leeseika/cv-demo/pkg/errs"
	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/rs/zerolog/log"
)

func (r *reference) GetProductRef(ctx context.Context, productID string) (*cache.ProductReference, error) {
	productRef, err := r.referenceCacheDAO.GetProductRef(ctx, productID)
	if err == nil {
		return productRef, nil
	}

	// log error if it's not cache miss error
	if !errs.IsKVCacheError(err, kvcache.ErrKeyCacheMissed) {
		log.Err(err).Msg("failed to get product reference cache")
	}

	resultCh := r.singleFlightGroup.DoChan(fmt.Sprintf("product_ref:%s", productID), func() (interface{}, error) {
		cached, cacheErr := r.referenceCacheDAO.GetProductRef(ctx, productID)
		if cacheErr == nil {
			return cached, nil
		}

		variantIDs, dbErr := r.productVariantDAO.GetVariantIDsByProductID(ctx, productID)
		if dbErr != nil {
			log.Err(dbErr).Msg("failed to get variant ids by product id")
			return nil, dbErr
		}
		imageIDMap, dbErr := r.productImageDAO.GetImageIDsByProductIDs(ctx, []string{productID})
		if dbErr != nil {
			log.Err(dbErr).Msg("failed to get image ids by product id")
			return nil, dbErr
		}
		productImages := imageIDMap[productID]
		imageIDs := make([]string, 0, len(productImages))
		for _, img := range productImages {
			imageIDs = append(imageIDs, img.ImageID)
		}

		rebuilt := cache.BuildProductReference(
			productID,
			imageIDs,
			variantIDs,
		)

		setErr := r.referenceCacheDAO.SetProductRef(ctx, productID, rebuilt)
		if setErr != nil {
			log.Err(setErr).Msg("failed to set product reference cache")
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
		return result.Val.(*cache.ProductReference), nil
	}
}
