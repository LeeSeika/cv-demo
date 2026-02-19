package reference

import (
	"context"

	"github.com/leeseika/cv-demo/pkg/model/cache"
)

func (r *reference) GetMultiProductVariantRefs(ctx context.Context, variantIDs []string) (map[string]*cache.ProductVariantReference, error) {
	variantReferenceMap, err := r.referenceCacheDAO.GetMultiProductVariantRefs(ctx, variantIDs)
	if err == nil {
		return variantReferenceMap, nil
	}

	variantImageMap, err := r.productVariantImageDAO.GetImageIDsByProductVariantIDs(ctx, variantIDs)
	if err != nil {
		return nil, err
	}

	variantReferenceMap = make(map[string]*cache.ProductVariantReference, len(variantIDs))
	for _, variantID := range variantIDs {
		variantImages := variantImageMap[variantID]
		imageIDs := make([]string, 0, len(variantImages))
		for _, img := range variantImages {
			imageIDs = append(imageIDs, img.ImageID)
		}
		variantReferenceMap[variantID] = cache.BuildProductVariantReference(
			variantID,
			imageIDs,
		)
	}

	err = r.referenceCacheDAO.SetMultiProductVariantRefs(ctx, variantReferenceMap)
	if err != nil {
		return nil, err
	}

	return variantReferenceMap, nil
}
