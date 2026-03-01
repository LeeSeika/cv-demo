package template

func templateDraftKey(id string, userID string) string {
	return "template:draft:" + id + ":" + userID
}
