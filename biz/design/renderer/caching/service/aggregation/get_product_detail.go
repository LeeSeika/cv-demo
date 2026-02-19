package aggregation

import (
	"context"

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

	if errs.IsKVError(err, kvcache.ErrKeyNotFound) {
		log.Warn().Err(err).Msg("empty value placeholder hit for product detail")
		return nil, err
	}
	// log the error if it's not a cache miss
	if !errs.IsKVError(err, kvcache.ErrKeyCacheMissed) {
		log.Err(err).Msg("failed to get product detail from id")
	}

	// query from model services & rebuild
	// query product
	product, err := productSvc.Get().GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	// query shop
	shop, err := shopSvc.Get().GetByID(ctx, product.ShopID)
	if err != nil {
		return nil, err
	}
	// query product reference
	productRef, err := referenceSvc.Get().GetProductRef(ctx, productID)
	if err != nil {
		return nil, err
	}
	// query product variants
	variantIDs := productRef.VariantIDs
	productVariantMap, err := productSvc.Get().GetProductVariantsByIDs(ctx, variantIDs)
	if err != nil {
		return nil, err
	}
	// query product variant references
	variantRefs, err := referenceSvc.Get().GetMultiProductVariantRefs(ctx, variantIDs)
	if err != nil {
		return nil, err
	}
	// query images
	imageIDs := append([]string{}, productRef.ImageIDs...)
	for _, variantRef := range variantRefs {
		imageIDs = append(imageIDs, variantRef.ImageIDs...)
	}
	imageMap, err := imageSvc.Get().GetImagesByIDs(ctx, imageIDs)
	if err != nil {
		return nil, err
	}

	// assemble product variant details
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
	// assemble product detail
	productImages := make([]*cache.Image, 0, len(productRef.ImageIDs))
	for _, imageID := range productRef.ImageIDs {
		productImages = append(productImages, imageMap[imageID])
	}
	productDetail = dto.BuildProductDetail(
		*product,
		variantDetails,
		productImages,
	)

	// set cache
	err = a.aggregationCacheDAO.SetProductDetail(ctx, productID, productDetail)
	if err != nil {
		log.Err(err).Msg("failed to set product detail cache")
	}

	return productDetail, nil
}
