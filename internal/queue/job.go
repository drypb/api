package queue

import "mime/multipart"

type Job struct {
	ID       string
	File     *multipart.FileHeader
	Template int
}
