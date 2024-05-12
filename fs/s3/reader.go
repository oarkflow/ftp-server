package s3

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type reader struct{ object *s3.GetObjectOutput }

func (reader reader) ReadAt(buffer []byte, offset int64) (int, error) {
	if n, err := reader.object.Body.Read(buffer); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}
