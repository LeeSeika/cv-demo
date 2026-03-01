package pubsub

type Subscriber interface {
	Start() error
	Stop() error
}

func StartSubscribers(subscribers ...Subscriber) error {
	for _, s := range subscribers {
		if err := s.Start(); err != nil {
			return err
		}
	}
	return nil
}

func StopSubscribers(subscribers ...Subscriber) error {
	for _, s := range subscribers {
		if err := s.Stop(); err != nil {
			return err
		}
	}
	return nil
}
