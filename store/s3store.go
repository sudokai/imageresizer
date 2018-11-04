package store

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
)

type S3Store struct {
	bucket     *string
	prefix     string
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	S3         *s3.S3
}

type S3Config struct {
	Region string
	Bucket string
	Prefix string
}

func NewS3Store(config *S3Config) (*S3Store, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region)},
	)
	if err != nil {
		return nil, err
	}

	return &S3Store{
		bucket:     aws.String(config.Bucket),
		prefix:     config.Prefix,
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		S3:         s3.New(sess),
	}, nil
}

func (s *S3Store) Get(filename string) ([]byte, error) {
	writeAtBuf := aws.NewWriteAtBuffer([]byte{})
	_, err := s.downloader.Download(writeAtBuf,
		&s3.GetObjectInput{
			Bucket: s.bucket,
			Key:    aws.String(s.prefix + "/" + filename),
		})
	if err != nil {
		s3err, ok := err.(awserr.RequestFailure)
		if ok && s3err.StatusCode() == 404 {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	return writeAtBuf.Bytes(), nil
}

func (s *S3Store) Put(filename string, buf []byte) error {
	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: s.bucket,
		Key:    aws.String(s.prefix + "/" + filename),
		Body:   bytes.NewReader(buf),
	})
	return err
}

func (s *S3Store) Remove(filename string) error {
	key := aws.String(s.prefix + "/" + filename)
	_, err := s.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: s.bucket,
		Key:    key,
	})
	if err != nil {
		return err
	}
	err = s.S3.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: s.bucket,
		Key:    key,
	})
	return err
}
