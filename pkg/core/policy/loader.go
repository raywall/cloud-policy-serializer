package policy

import (
	"fmt"
	"os"

	"github.com/raywall/aws-policy-engine-go/pkg/core/loader"
)

const withDecryption bool = true

type (
	PolicyLoader interface {
		Load() (*PolicyEngine, error)
	}

	localLoader struct {
		loader *loader.LocalLoader
	}

	s3Loader struct {
		loader *loader.S3Loader
	}

	ssmLoader struct {
		loader *loader.SSMLoader
	}
)

func NewLoader(source string) (PolicyLoader, error) {
	ld, err := loader.NewLoader(source)
	if err != nil {
		return nil, err
	}

	switch v := ld.(type) {
	case *loader.LocalLoader:
		return &localLoader{
			loader: v,
		}, nil
	case *loader.S3Loader:
		return &s3Loader{
			loader: v,
		}, nil
	case *loader.SSMLoader:
		return &ssmLoader{
			loader: v,
		}, nil
	default:
		return nil, fmt.Errorf("tipo de loader não suportado para schema: %T", v)
	}
}

func (l *localLoader) Load() (*PolicyEngine, error) {
	data, err := os.ReadFile(l.loader.Path)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	jsonSchema, err := NewPolicyEngine(data)
	if err != nil {
		return nil, err
	}

	return jsonSchema, nil
}

func (l *ssmLoader) Load() (*PolicyEngine, error) {
	data, err := loader.GetParameter(l.loader.Client, l.loader.Path, withDecryption)
	if err != nil {
		return nil, err
	}

	jsonSchema, err := NewPolicyEngine(data)
	if err != nil {
		return nil, err
	}

	return jsonSchema, nil
}

func (l *s3Loader) Load() (*PolicyEngine, error) {
	bucket, key := loader.ParseS3Path(l.loader.Path)

	data, err := loader.GetObject(l.loader.Client, bucket, key)
	if err != nil {
		return nil, err
	}

	jsonSchema, err := NewPolicyEngine(data)
	if err != nil {
		return nil, err
	}

	return jsonSchema, nil
}
