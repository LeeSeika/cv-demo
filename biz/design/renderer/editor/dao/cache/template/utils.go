package template

import "time"

const defaultTTL = 1 * time.Hour

func templateDraftKey(id string, userID string) string {
	return "template:draft:" + id + ":" + userID
}
