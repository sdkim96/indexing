package job

import "time"

type Job struct {
	id        string
	key       string
	file      *File
	createdAt int64
}

func (j Job) ID() string {
	return j.id
}

func (j Job) Key() string {
	return j.key
}

func (j Job) File() *File {
	return j.file
}

func (j Job) CreatedAt() int64 {
	return j.createdAt
}

func NewJob(id, key string, file File) Job {
	createdAt := time.Now().Unix()
	return Job{
		id:        id,
		key:       key,
		file:      &file,
		createdAt: createdAt,
	}
}
