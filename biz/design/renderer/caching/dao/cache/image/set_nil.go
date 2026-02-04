package image

import "context"

func (i *image) SetNil(ctx context.Context, imageID string) error {
	key := imageKey(imageID)

	return i.kvCache.SetEmptyValuePlaceholder(ctx, key)
}
