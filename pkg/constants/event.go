package constants

type Topic string

const (
	TopicProduct Topic = "product"
)

type Action string

const (
	ActionCreated Action = "created"
	ActionUpdated Action = "updated"
	ActionDeleted Action = "deleted"
)
