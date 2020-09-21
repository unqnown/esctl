package backup

type Document struct {
	ID    string      `json:"_id"`
	Index string      `json:"index"`
	Body  interface{} `json:"body"`
}
