package pubsub

import (
	"context"
	"strings"
	"time"

	templateCacheDAO "github.com/leeseika/cv-demo/biz/design/renderer/editor/dao/cache/template"
	templateDAO "github.com/leeseika/cv-demo/biz/design/renderer/editor/dao/db/template"
	"github.com/leeseika/cv-demo/pkg/model/cache"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type templateDraftSubscriber struct {
	rdb                   *redis.Client
	templateDraftCacheDAO templateCacheDAO.TemplateDraft
	templateDraftDAO      templateDAO.TemplateDraft
	client                *redis.PubSub
}

func NewTemplateDraftSubscriber(rdb *redis.Client, templateDraftCacheDAO templateCacheDAO.TemplateDraft, templateDraftDAO templateDAO.TemplateDraft) Subscriber {
	return &templateDraftSubscriber{
		rdb:                   rdb,
		templateDraftCacheDAO: templateDraftCacheDAO,
		templateDraftDAO:      templateDraftDAO,
	}
}

func (s *templateDraftSubscriber) Start() error {
	// 订阅所有以 template:draft: 开头的键的变更事件（Keyspace Notifications）
	// 注意：Redis 需要配置 notify-keyspace-events 为 "KA" 或相关组合 (K=Keyspace events, A=All or specific like 'g' for generic commands, 'string' etc)
	// 这里假设配置已开启。
	// 使用 PSUBSCRIBE 模式订阅
	ctx := context.Background()
	pubsub := s.rdb.PSubscribe(ctx, "__keyspace@0__:template:draft:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	s.client = pubsub

	const (
		batchSize     = 10
		flushInterval = 5 * time.Second
	)

	var (
		buffer []string
		timer  = time.NewTimer(flushInterval)
	)

	flush := func() {
		if len(buffer) == 0 {
			return
		}

		// 去重，避免同一个key在一次batch中多次处理
		uniqueKeys := make(map[string]struct{})
		for _, key := range buffer {
			uniqueKeys[key] = struct{}{}
		}

		keys := make([]string, 0, len(uniqueKeys))
		for key := range uniqueKeys {
			keys = append(keys, key)
		}

		draftMap, err := s.templateDraftCacheDAO.GetMultiDraftsByKeys(ctx, keys)
		if err != nil {
			log.Err(err).Msg("failed to get multi template drafts by keys during flush")
			// 如果批量获取失败，可能是redis出现问题，这里如果不清空buffer，可以下次重试
			// 但考虑到buffer可能有旧数据，最好还是清或者逐个获取
			// 简单处理：记录错误并继续，如果需要高可靠性可以引入重试机制
		}

		if len(draftMap) > 0 {
			drafts := make([]*cache.TemplateDraft, 0, len(draftMap))
			for _, draft := range draftMap {
				drafts = append(drafts, draft)
			}
			if err := s.templateDraftDAO.BatchSaveDraft(ctx, drafts); err != nil {
				log.Err(err).Msg("failed to batch save template drafts during flush")
			}
		}

		buffer = buffer[:0]
		timer.Reset(flushInterval)
	}

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				// Channel closed
				flush()
				return nil
			}

			if msg.Payload == "set" {
				key := strings.TrimPrefix(msg.Channel, "__keyspace@0__:")
				buffer = append(buffer, key)
				if len(buffer) >= batchSize {
					flush() // 这里 flush 会重置 timer
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(flushInterval)
				}
			}

		case <-timer.C:
			flush()
		}
	}
}

func (s *templateDraftSubscriber) Stop() error {
	if s.client != nil {
		err := s.client.Close()
		if err == redis.ErrClosed {
			return nil
		}
		return err
	}
	return nil
}
