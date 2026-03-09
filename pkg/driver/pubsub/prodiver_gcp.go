package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gcpPubSub "cloud.google.com/go/pubsub"
	"github.com/leeseika/cv-demo/pkg/config"
	"github.com/leeseika/cv-demo/pkg/constants"
	"github.com/leeseika/cv-demo/pkg/model/dto"
	gcpOption "google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GooglePubsubPublisher struct {
	topics    map[string]*gcpPubSub.Topic
	client    *gcpPubSub.Client
	projectID string
	conf      config.GooglePubsubConf
}

type GooglePubsubSubscriber struct {
	client    *gcpPubSub.Client
	projectID string
	conf      config.GooglePubsubConf
}

func NewGooglePubsubPublisher(conf config.GooglePubsubConf) (*GooglePubsubPublisher, error) {
	ctx := context.Background()
	if conf.UseMock {
		gpsTopics := conf.PublisherConfig.Topics
		client, err := gcpPubSub.NewClient(ctx, conf.MockPubsub.ProjectID,
			gcpOption.WithEndpoint(conf.MockPubsub.Endpoint),
			gcpOption.WithoutAuthentication(),
			gcpOption.WithGRPCDialOption(grpc.WithInsecure()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Event Streams Google Cloud Pub/Sub client: %+v", err)
		}
		topics, err := createTopics(ctx, client, conf.MockPubsub.ProjectID, gpsTopics)
		if err != nil {
			return nil, fmt.Errorf("failed to create Event Streams Google Cloud Pub/Sub topic: %+v", err)
		}

		return &GooglePubsubPublisher{
			topics:    topics,
			client:    client,
			projectID: conf.MockPubsub.ProjectID,
			conf:      conf,
		}, nil
	} else {
		gpsServiceAccount := conf.GPSServiceAccount
		gpsTopics := conf.PublisherConfig.Topics

		gpsServiceAccountContent, err := json.Marshal(gpsServiceAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal Event Streams Google Cloud Pub/Sub Service Account: %+v", err)
		}

		opt := gcpOption.WithCredentialsJSON(gpsServiceAccountContent)
		client, err := gcpPubSub.NewClient(ctx, gpsServiceAccount.ProjectID, opt)
		if err != nil {
			return nil, fmt.Errorf("unable to connect to Event Streams Google Cloud Pub/Sub: %+v", err)
		}

		topics, err := createTopics(ctx, client, gpsServiceAccount.ProjectID, gpsTopics)
		if err != nil {
			return nil, fmt.Errorf("failed to create Event Streams Google Cloud Pub/Sub topic: %+v", err)
		}

		return &GooglePubsubPublisher{
			topics:    topics,
			client:    client,
			projectID: gpsServiceAccount.ProjectID,
			conf:      conf,
		}, nil
	}
}

func createTopics(ctx context.Context, client *gcpPubSub.Client, projectID string, topicNames []string) (map[string]*gcpPubSub.Topic, error) {
	topics := make(map[string]*gcpPubSub.Topic, len(topicNames))
	for _, topicName := range topicNames {
		topic, err := client.CreateTopic(ctx, topicName)
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.AlreadyExists {
				topic = client.TopicInProject(topicName, projectID)
			} else {
				return nil, err
			}
		}

		topics[topicName] = topic
	}
	return topics, nil
}

func (p *GooglePubsubPublisher) Publish(ctx context.Context, event dto.Event) error {
	result, err := p.PublishAsync(ctx, event)
	if err != nil {
		return err
	}

	_, err = result.Get(ctx)
	return err
}

func (p *GooglePubsubPublisher) PublishAsync(ctx context.Context, event dto.Event) (*gcpPubSub.PublishResult, error) {
	topic, ok := p.topics[string(event.Topic)]
	if !ok {
		return nil, fmt.Errorf("topic %s not found in Google Pub/Sub provider", event.Topic)
	}

	b, err := event.ToBytes()
	if err != nil {
		return nil, err
	}

	result := topic.Publish(ctx, &gcpPubSub.Message{
		Data: b,
	})
	return result, nil
}

func NewGooglePubsubSubscriber(conf config.GooglePubsubConf) (*GooglePubsubSubscriber, error) {
	ctx := context.Background()
	if conf.UseMock {
		client, err := gcpPubSub.NewClient(ctx, conf.MockPubsub.ProjectID,
			gcpOption.WithEndpoint(conf.MockPubsub.Endpoint),
			gcpOption.WithoutAuthentication(),
			gcpOption.WithGRPCDialOption(grpc.WithInsecure()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Event Streams Google Cloud Pub/Sub client: %+v", err)
		}

		return &GooglePubsubSubscriber{
			client:    client,
			projectID: conf.MockPubsub.ProjectID,
			conf:      conf,
		}, nil
	} else {
		gpsServiceAccount := conf.GPSServiceAccount

		gpsServiceAccountContent, err := json.Marshal(gpsServiceAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal Event Streams Google Cloud Pub/Sub Service Account: %+v", err)
		}

		opt := gcpOption.WithCredentialsJSON(gpsServiceAccountContent)
		client, err := gcpPubSub.NewClient(ctx, gpsServiceAccount.ProjectID, opt)
		if err != nil {
			return nil, fmt.Errorf("unable to connect to Event Streams Google Cloud Pub/Sub: %+v", err)
		}

		return &GooglePubsubSubscriber{
			client:    client,
			projectID: gpsServiceAccount.ProjectID,
			conf:      conf,
		}, nil
	}
}

func (p *GooglePubsubSubscriber) CreateSubscription(ctx context.Context, topic constants.Topic) (*gcpPubSub.Subscription, error) {
	topicName := string(topic)
	deadLetterTopicName := topicName + "-dead-letter"
	topics := []string{topicName, deadLetterTopicName}
	topicsMap, err := createTopics(ctx, p.client, p.projectID, topics)
	if err != nil {
		return nil, fmt.Errorf("failed to create topics for subscription: %+v", err)
	}

	return p.client.CreateSubscription(ctx, p.conf.SubscriberConfig.SubscriberID, gcpPubSub.SubscriptionConfig{
		Topic: topicsMap[topicName],
		DeadLetterPolicy: &gcpPubSub.DeadLetterPolicy{
			DeadLetterTopic:     topicsMap[deadLetterTopicName].String(),
			MaxDeliveryAttempts: p.conf.SubscriberConfig.MaxDeliveryAttempts,
		},
		RetryPolicy: &gcpPubSub.RetryPolicy{
			MinimumBackoff: time.Duration(p.conf.SubscriberConfig.MinimumBackoff) * time.Second,
			MaximumBackoff: time.Duration(p.conf.SubscriberConfig.MaximumBackoff) * time.Second,
		},
	})
}

func (p *GooglePubsubSubscriber) Close() error {
	return p.client.Close()
}
