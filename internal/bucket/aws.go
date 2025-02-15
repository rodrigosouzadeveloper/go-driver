package bucket

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type AwsConfig struct {
	Config *aws.Config
	BucketDownload string
	BucketUpload string
}

func newAwsSession(cfg AwsConfig) *awsSession {
	c, err := session.NewSession(cfg.Config)
	if err != nil {
		panic(err)
	}

	return &awsSession{
		sess: c,
		bucketDownload: cfg.BucketDownload,
		bucketUpload: cfg.BucketUpload,
	}
}

type awsSession struct {
	sess *session.Session
	bucketDownload string
	bucketUpload string
}

func (as *awsSession) Upload(file io.Reader, key string) error {
	uploader := s3manager.NewUploader(as.sess)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(as.bucketUpload),
		Key:   aws.String(key),
		Body: file,
	})
	return err
}

func (as *awsSession) Download(src string, destination string) (file *os.File, err error) {
	file, err = os.Create(destination)
	if err != nil {
		return
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(as.sess)
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(as.bucketDownload),
		Key:    aws.String(src),
	})

	return
}

func (as *awsSession) Delete(key string) error {
	svc := s3.New(as.sess)

	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(as.bucketDownload),
		Key:    aws.String(key),
	})

	if err != nil {
		return err
	}

	return svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(as.bucketDownload),
		Key:    aws.String(key),
	})
}
