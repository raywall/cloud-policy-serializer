package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type (
	Loader interface {
		Load() (*Schema, error)
	}

	LocalLoader struct {
		Path string
	}

	S3Loader struct {
		Path   string
		Client S3Client
	}

	S3Client interface {
		GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	}

	SSMLoader struct {
		Path   string
		Client SSMClient
	}

	SSMClient interface {
		GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
	}
)

func NewLoader(source string) (Loader, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	switch {
	case strings.HasPrefix(source, "s3://"):
		return &S3Loader{
			Path:   source,
			Client: s3.NewFromConfig(cfg),
		}, nil
	case strings.HasPrefix(source, "ssm://"):
		return &SSMLoader{
			Path:   source,
			Client: ssm.NewFromConfig(cfg),
		}, nil
	default:
		return &LocalLoader{
			Path: source,
		}, nil
	}
}

// Load carrega um JSON Schema draft-07
func (s *Schema) load(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return fmt.Errorf("unable to serialize ssm template file: %v", err)
	}

	return nil
}

func (l *LocalLoader) Load() (*Schema, error) {
	jsonSchema := Schema{}

	data, err := os.ReadFile(l.Path)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if err = jsonSchema.load(data); err != nil {
		return nil, err
	}

	return &jsonSchema, nil
}

func (l *SSMLoader) Load() (*Schema, error) {
	var (
		jsonSchema     = Schema{}
		withDecryption = true
	)

	output, err := l.Client.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           &l.Path,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting parameter %s: %v", l.Path, err)
	}

	if output.Parameter == nil || output.Parameter.Value == nil {
		return nil, fmt.Errorf("invalid parameter: %v", err)
	}

	if err = jsonSchema.load([]byte(*output.Parameter.Value)); err != nil {
		return nil, err
	}

	return &jsonSchema, nil
}

func (l *S3Loader) Load() (*Schema, error) {
	jsonSchema := Schema{}

	bucket, key := parseS3Path(l.Path)
	output, err := l.Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting object %s from bucket %s: %v", key, bucket, err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if err = jsonSchema.load(data); err != nil {
		return nil, err
	}

	return &jsonSchema, nil
}

func parseS3Path(path string) (bucket, key string) {
	path = strings.TrimPrefix(path, "s3://")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
