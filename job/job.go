package job

import "time"

type Job struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	File      *File  `json:"file"`
	CreatedAt int64  `json:"created_at"`
}

func NewJob(id, key string, file File) Job {
	createdAt := time.Now().Unix()
	return Job{
		ID:        id,
		Key:       key,
		File:      &file,
		CreatedAt: createdAt,
	}
}
