package part

type Part struct {
	ID        string `json:"id"`
	SourceID  string `json:"source_id"`
	Data      Data   `json:"data"`
	Page      int    `json:"page"`
	Offset    int    `json:"offset"`
	CreatedAt int64  `json:"created_at"`
}
