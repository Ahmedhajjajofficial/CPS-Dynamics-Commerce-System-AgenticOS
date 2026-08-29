package featurestore

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cps-enterprise/dcs/ml-platform/feature-store/pkg/db"
	"github.com/cps-enterprise/dcs/ml-platform/feature-store/pkg/redis"
	"go.uber.org/zap"
)

// FeatureStore provides access to ML features
type FeatureStore struct {
	postgres *db.Postgres
	redis    *redis.Client
	logger   *zap.Logger
	mu       sync.RWMutex
}

// Feature represents a computed feature
type Feature struct {
	Name      string
	EntityID  string
	Value     float64
	Timestamp time.Time
	TTL       time.Duration
	Tags      map[string]string
}

// FeatureGroup represents a group of related features
type FeatureGroup struct {
	Name        string
	Description string
	Features    []string
	TTL         time.Duration
}

// NewFeatureStore creates a new feature store
func NewFeatureStore(postgresURL string, redisURL string, logger *zap.Logger) (*FeatureStore, error) {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}

	postgres, err := db.NewPostgres(postgresURL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	redisClient, err := redis.NewClient(redisURL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &FeatureStore{
		postgres: postgres,
		redis:    redisClient,
		logger:   logger,
	}, nil
}

// GetFeature retrieves a feature value
func (fs *FeatureStore) GetFeature(ctx context.Context, entityID string, featureName string) (*Feature, error) {
	cacheKey := fmt.Sprintf("feature:%s:%s", entityID, featureName)

	// Try Redis cache first
	cached, err := fs.redis.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var feature Feature
		if err := json.Unmarshal([]byte(cached), &feature); err == nil {
			return &feature, nil
		}
	}

	// Fallback to PostgreSQL
	feature, err := fs.postgres.GetFeature(ctx, entityID, featureName)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if feature != nil {
		data, _ := json.Marshal(feature)
		fs.redis.Set(ctx, cacheKey, string(data), feature.TTL)
	}

	return feature, nil
}

// SetFeature stores a feature value
func (fs *FeatureStore) SetFeature(ctx context.Context, feature *Feature) error {
	// Write to PostgreSQL
	if err := fs.postgres.SetFeature(ctx, feature); err != nil {
		return err
	}

	// Update cache
	cacheKey := fmt.Sprintf("feature:%s:%s", feature.EntityID, feature.Name)
	data, _ := json.Marshal(feature)
	return fs.redis.Set(ctx, cacheKey, string(data), feature.TTL)
}

// GetFeaturesBatch retrieves multiple features for an entity
func (fs *FeatureStore) GetFeaturesBatch(ctx context.Context, entityID string, featureNames []string) (map[string]*Feature, error) {
	result := make(map[string]*Feature)

	for _, name := range featureNames {
		feature, err := fs.GetFeature(ctx, entityID, name)
		if err != nil {
			return nil, err
		}
		if feature != nil {
			result[name] = feature
		}
	}

	return result, nil
}

// ComputeFeature computes a feature from raw data
func (fs *FeatureStore) ComputeFeature(ctx context.Context, entityID string, featureName string, computeFunc func(ctx context.Context) (float64, error)) (*Feature, error) {
	value, err := computeFunc(ctx)
	if err != nil {
		return nil, err
	}

	feature := &Feature{
		Name:      featureName,
		EntityID:  entityID,
		Value:     value,
		Timestamp: time.Now(),
		TTL:       24 * time.Hour,
		Tags:      make(map[string]string),
	}

	if err := fs.SetFeature(ctx, feature); err != nil {
		return nil, err
	}

	return feature, nil
}

// Close closes the feature store connections
func (fs *FeatureStore) Close() error {
	if fs.redis != nil {
		if err := fs.redis.Close(); err != nil {
			return err
		}
	}
	if fs.postgres != nil {
		fs.postgres.Close()
	}
	return nil
}
