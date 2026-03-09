package pubsub

import "context"

type Subscriber interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func StartSubscribers(ctx context.Context, subscribers ...Subscriber) error {
	for _, s := range subscribers {
		if err := s.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

func StopSubscribers(ctx context.Context, subscribers ...Subscriber) error {
	for _, s := range subscribers {
		if err := s.Stop(ctx); err != nil {
			return err
		}
	}
	return nil
}
