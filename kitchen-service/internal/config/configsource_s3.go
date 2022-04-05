package config

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/w-k-s/konfig"
	"github.com/w-k-s/konfig/loader/kls3"
	"github.com/w-k-s/konfig/parser/kptoml"
)

type s3ConfigSource struct {
	awsAccessKey string
	awsSecretKey string
	awsRegion    string
}

func newS3ConfigSource(accessKey, secretKey, region string) (*s3ConfigSource, error) {
	return &s3ConfigSource{
		awsAccessKey: accessKey,
		awsSecretKey: secretKey,
		awsRegion:    region,
	}, nil
}

type S3Uri struct {
	BucketName string
	Key        string
}

func ParseS3Uri(s3Uri string) (S3Uri, error) {
	uriWithoutProtcol := strings.Replace(s3Uri, "s3://", "", 1)
	parts := strings.Split(uriWithoutProtcol, "/")
	if len(parts) < 2 || len(parts[1]) == 0 {
		return S3Uri{}, fmt.Errorf("failed to parse S3 URI. Not enough parts to extract bucket name and object key")
	}
	return S3Uri{
		BucketName: parts[0],
		Key:        strings.Join(parts[1:], "/"),
	}, nil
}

func (s S3Uri) String() string {
	return fmt.Sprintf("s3://%s/%s", s.BucketName, s.Key)
}

func (s3Config s3ConfigSource) Load(s3UriString string) (*Config, error) {
	var (
		sess *session.Session
		err  error
	)
	if len(s3Config.awsAccessKey) == 0 || len(s3Config.awsSecretKey) == 0 {
		sess, err = session.NewSession()
	} else {
		sess, err = session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region:      aws.String(s3Config.awsRegion),
				Credentials: credentials.NewStaticCredentials(s3Config.awsAccessKey, s3Config.awsSecretKey, ""),
			},
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session. Reason: %w", err)
	}

	s3Uri, err := ParseS3Uri(s3UriString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse s3 uri '%s'. Reason: %w", s3Uri, err)
	}

	configStore := konfig.New(konfig.DefaultConfig())
	configStore.RegisterLoaderWatcher(
		kls3.New(
			&kls3.Config{
				Objects: []kls3.Object{
					{
						Parser: kptoml.Parser,
						Bucket: s3Uri.BucketName,
						Key:    s3Uri.Key,
					},
				},
				Downloader: s3manager.NewDownloader(sess),
			},
		),
	)

	if err := configStore.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config. Reason: '%w'", err)
	}

	return readValues(configStore)
}
