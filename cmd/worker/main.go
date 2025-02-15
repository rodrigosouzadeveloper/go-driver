package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/rodrigosouzadeveloper/go-driver/internal/bucket"
	"github.com/rodrigosouzadeveloper/go-driver/internal/queue"
)

func main() {
	qcfg := queue.RabbitMQConfig{
		URL: os.Getenv("RABBIT_URL"),
		TopicName: os.Getenv("RABBIT_TOPIC_NAME"),
		Timeout: time.Second * 30,
	}

	qc, err := queue.New(queue.RabbitMQ, qcfg)
	if err != nil {
		panic(err)
	}

	c := make(chan queue.QueueDto)
	qc.Consume(c)

	bcfg := bucket.AwsConfig{
		Config: &aws.Config{
			Region: aws.String(os.Getenv("AWS_REGION")),
			Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_KEY"), os.Getenv("AWS_SECRET"), ""),
		},
		BucketDownload: "aprenda-golang-drive-raw",
		BucketUpload: "aprenda-golang-drive-gzip",
	}

	b, err := bucket.New(bucket.AwsProvider, bcfg)
	if err != nil {
		panic(err)
	}

	for msg := range c {
		src := fmt.Sprintf("%s/%s", msg.Path, msg.Filename)
		destination := fmt.Sprintf("%d_%s", msg.ID, msg.Filename)

		file, err := b.Download(src, destination)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		body, err := io.ReadAll(file)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, err = zw.Write(body)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		if err := zw.Close(); err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		zr, err := gzip.NewReader(&buf)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		err = b.Upload(zr, src)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		err = os.Remove(destination)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}
	}
}
