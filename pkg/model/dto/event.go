package dto

import (
	"encoding/json"

	"github.com/leeseika/cv-demo/pkg/constants"
)

type Event struct {
	Topic   constants.Topic  `json:"topic"`
	Action  constants.Action `json:"action"`
	Payload any              `json:"payload"`
}

func (e *Event) ToBytes() ([]byte, error) {
	return json.Marshal(e)
}

type EventPayloadProduct struct {
	ProductID string `json:"product_id"`
}
