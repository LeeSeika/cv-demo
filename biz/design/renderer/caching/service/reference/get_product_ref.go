package reference

import (
	"context"

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
	if !errs.IsKVError(err, kvcache.ErrKeyCacheMissed) {
		log.Err(err).Msg("failed to get product reference cache")
	}

	// query from database
	variantIDs, err := r.productVariantDAO.GetVariantIDsByProductID(ctx, productID)
	if err != nil {
		log.Err(err).Msg("failed to get variant ids by product id")
		return nil, err
	}
	imageIDMap, err := r.productImageDAO.GetImageIDsByProductIDs(ctx, []string{productID})
	if err != nil {
		log.Err(err).Msg("failed to get image ids by product id")
		return nil, err
	}
	productImages := imageIDMap[productID]
	imageIDs := make([]string, 0, len(productImages))
	for _, img := range productImages {
		imageIDs = append(imageIDs, img.ImageID)
	}

	productRefCache := cache.BuildProductReference(
		productID,
		imageIDs,
		variantIDs,
	)

	// set cache
	err = r.referenceCacheDAO.SetProductRef(ctx, productID, productRefCache)
	if err != nil {
		log.Err(err).Msg("failed to set product reference cache")
	}

	return productRefCache, nil
}
