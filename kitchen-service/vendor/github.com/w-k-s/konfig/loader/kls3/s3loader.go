package kls3

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/w-k-s/konfig"
	"github.com/w-k-s/konfig/parser"
	"github.com/w-k-s/konfig/watcher/kwpoll"
)

var (
	_ konfig.Loader = (*Loader)(nil)
	// ErrNoObjects is the error thrown when trying to create a s3 loader with no objects to download in config
	ErrNoObjects = errors.New("no objects provided")
	// ErrNoParser is the error thrown when trying to create a file loader with no parser
	ErrNoParser = errors.New("no parser provided")
	// DefaultRate is the default polling rate to check files
	DefaultRate = 10 * time.Second
)

const (
	defaultName = "s3"
)

// Object is an object to load from s3 object storage
type Object struct {
	// Bucket is the bucket of the file
	Bucket string
	// Name is the name of the object
	Key string 
	// Parser is the parser used to parse file and add it to the config store
	Parser parser.Parser
}

func (o Object) Download(downloader Downloader) (io.Reader, error){
	buff := &aws.WriteAtBuffer{}
	if _, err := downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(o.Bucket),
		Key:    aws.String(o.Key),
	}); err != nil{
		return nil, fmt.Errorf("failed to download file from s3://%s/%s. Reason: %w", o.Bucket,o.Key, err)
	}
	return bytes.NewReader(buff.Bytes()),nil
}

// Downloader is the interface used to download objects from S3.
// It is implemented by s3manager.Downloader.
type Downloader interface {
	Download(io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64,error)
}

var defaultDownloader = s3manager.NewDownloader(session.Must(session.NewSession()))

// Config is the config for the file loader
type Config struct {
	// Name is the name of the loader
	Name string
	// StopOnFailure tells whether a failure to load configs should closed the config and all registered closers
	StopOnFailure bool
	// Objects is the path to the objects to load
	Objects []Object
	// MaxRetry is the maximum number of times load can be retried in config
	MaxRetry int
	// RetryDelay is the delay between each retry
	RetryDelay time.Duration
	// Debug sets the debug mode on the file loader
	Debug bool
	// Watch sets the whether changes should be watched
	Watch bool
	// Rater is the rater to pass to the poll write
	Rater kwpoll.Rater
	// S3 Downloader. IF null, creates a NewDownloader with a default session.
	// In this case, the Access Key, Secret Key and region will be determined from the ~/.aws/credentials
	// or, from the IAM role.
	Downloader Downloader
}

// Loader loads a configuration remotely
type Loader struct {
	*kwpoll.PollWatcher
	cfg *Config
}

// New returns a new Loader with the given Config.
func New(cfg *Config) *Loader {
	if cfg.Downloader == nil {
		cfg.Downloader = defaultDownloader
	}

	if cfg.Objects == nil || len(cfg.Objects) == 0 {
		panic(ErrNoObjects)
	}

	// make sure all files have a parser
	for _, f := range cfg.Objects {
		if f.Parser == nil {
			panic(ErrNoParser)
		}
	}

	if cfg.Name == "" {
		cfg.Name = defaultName
	}

	var l = &Loader{
		cfg: cfg,
	}

	if cfg.Watch {
		var v = konfig.Values{}
		var err = l.Load(v)
		if err != nil {
			panic(err)
		}
		l.PollWatcher = kwpoll.New(&kwpoll.Config{
			Loader:    l,
			Rater:     cfg.Rater,
			InitValue: v,
			Diff:      true,
			Debug:     cfg.Debug,
		})
	}

	return l
}

// Name returns the name of the loader
func (r *Loader) Name() string { return r.cfg.Name }

// Load loads the config from sources and parses the response
func (r *Loader) Load(s konfig.Values) error {
	for _, object := range r.cfg.Objects {
		if b, err := object.Download(r.cfg.Downloader); err == nil {
			if err := object.Parser.Parse(b, s); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// MaxRetry returns the MaxRetry config property, it implements the konfig.Loader interface
func (r *Loader) MaxRetry() int {
	return r.cfg.MaxRetry
}

// RetryDelay returns the RetryDelay config property, it implements the konfig.Loader interface
func (r *Loader) RetryDelay() time.Duration {
	return r.cfg.RetryDelay
}

// StopOnFailure returns whether a load failure should stop the config and the registered closers
func (r *Loader) StopOnFailure() bool {
	return r.cfg.StopOnFailure
}
