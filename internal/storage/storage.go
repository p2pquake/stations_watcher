package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/p2pquake/stations_watcher/internal/retreiver"
)

const (
	defaultRegion = "ap-northeast-1"
	jsonKey       = "stations.json"
	csvKey        = "Stations.csv"
)

type S3Storage struct {
	bucket  string
	jsonKey string
	csvKey  string
	client  *s3.Client
}

func NewS3Storage(ctx context.Context, bucket string) (*S3Storage, error) {
	cfg, err := loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &S3Storage{
		bucket:  bucket,
		jsonKey: jsonKey,
		csvKey:  csvKey,
		client:  s3.NewFromConfig(cfg),
	}, nil
}

func loadConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(defaultRegion))
	if err != nil {
		return aws.Config{}, err
	}
	if cfg.Region == "" {
		cfg.Region = defaultRegion
	}
	return cfg, nil
}

func (s *S3Storage) Load(ctx context.Context) ([]retreiver.SeismicIntensityStation, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.jsonKey),
	})
	if err != nil {
		var nsk *s3types.NoSuchKey
		if errors.As(err, &nsk) {
			return []retreiver.SeismicIntensityStation{}, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var stations []retreiver.SeismicIntensityStation
	if err := json.Unmarshal(body, &stations); err != nil {
		return nil, err
	}
	return stations, nil
}

func (s *S3Storage) Save(ctx context.Context, stations []retreiver.SeismicIntensityStation) error {
	data, err := json.Marshal(stations)
	if err != nil {
		return err
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.jsonKey),
		Body:   bytes.NewReader(data),
	})
	return err
}

func (s *S3Storage) SaveCSV(ctx context.Context, csv string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.csvKey),
		Body:   bytes.NewReader([]byte(csv)),
	})
	return err
}

func GetObjectBytes(ctx context.Context, bucket, key string) ([]byte, error) {
	cfg, err := loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
