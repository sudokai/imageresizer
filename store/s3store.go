package store

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Store struct {
	bucket     *string
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	S3         *s3.S3
}

func NewS3Store(region string, bucket string) (*S3Store, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return nil, err
	}

	return &S3Store{
		bucket:     aws.String(bucket),
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		S3:         s3.New(sess),
	}, nil
}

func (s *S3Store) Get(filename string) ([]byte, error) {
	writeAtBuf := aws.NewWriteAtBuffer([]byte{})
	_, err := s.downloader.Download(writeAtBuf,
		&s3.GetObjectInput{Bucket: s.bucket, Key: aws.String(filename)})
	if err != nil {
		return nil, err
	}
	return writeAtBuf.Bytes(), nil
}

func (s *S3Store) Put(filename string, buf []byte) error {
	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: s.bucket,
		Key:    aws.String(filename),
		Body:   bytes.NewReader(buf),
	})
	return err
}

func (s *S3Store) Remove(filename string) error {
	_, err := s.S3.DeleteObject(&s3.DeleteObjectInput{Bucket: s.bucket, Key: aws.String(filename)})
	if err != nil {
		return err
	}
	err = s.S3.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: s.bucket,
		Key:    aws.String(filename),
	})
	return err
}
