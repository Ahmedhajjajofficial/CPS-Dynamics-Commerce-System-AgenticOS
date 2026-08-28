package anomaly

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/cps-enterprise/dcs/ml-platform/anomaly-detection/pkg/stats"
	"github.com/cps-enterprise/dcs/proto"
	pb "github.com/cps-enterprise/dcs/master-agent/internal/proto"
	"go.uber.org/zap"
)

// AnomalyDetector detects anomalies in transaction patterns
type AnomalyDetector struct {
	logger    *zap.Logger
	models    map[string]*AnomalyModel
	modelsMu  sync.RWMutex
	threshold float64
}

// AnomalyModel represents an anomaly detection model
type AnomalyModel struct {
	EntityID      string
	Mean          float64
	StdDev        float64
	SampleCount   int64
	LastUpdated   time.Time
	WindowSize    int
	Samples       []float64
}

// AnomalyAlert represents an anomaly alert
type AnomalyAlert struct {
	EntityID     string
	Score        float64
	IsAnomaly    bool
	Confidence   float64
	DetectedAt   time.Time
	MetricValues map[string]float64
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(logger *zap.Logger) *AnomalyDetector {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}

	return &AnomalyDetector{
		logger:    logger,
		models:    make(map[string]*AnomalyModel),
		threshold: 3.0,
	}
}

// Detect detects anomalies in a transaction
func (d *AnomalyDetector) Detect(ctx context.Context, event *pb.SovereignFinancialEvent) (*AnomalyAlert, error) {
	if event == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}

	entityID := event.BranchId
	if entityID == "" {
		entityID = event.AgentId
	}

	d.modelsMu.RLock()
	model, exists := d.models[entityID]
	d.modelsMu.RUnlock()

	if !exists {
		model = &AnomalyModel{
			EntityID:    entityID,
			WindowSize:  100,
			Samples:     make([]float64, 0, 100),
			LastUpdated: time.Now(),
		}
		d.modelsMu.Lock()
		d.models[entityID] = model
		d.modelsMu.Unlock()
	}

	amount := event.Amount
	model.Samples = append(model.Samples, amount)
	if len(model.Samples) > model.WindowSize {
		model.Samples = model.Samples[1:]
	}
	model.SampleCount++
	model.LastUpdated = time.Now()

	if model.SampleCount < 10 {
		return &AnomalyAlert{
			EntityID:     entityID,
			Score:        0,
			IsAnomaly:    false,
			Confidence:   0,
			DetectedAt:   time.Now(),
			MetricValues: map[string]float64{"amount": amount},
		}, nil
	}

	mean, stdDev := stats.MeanStdDev(model.Samples)
	model.Mean = mean
	model.StdDev = stdDev

	zScore := 0.0
	if stdDev > 0 {
		zScore = math.Abs((amount - mean) / stdDev)
	}

	isAnomaly := zScore > d.threshold
	confidence := math.Min(zScore/5.0, 1.0)

	alert := &AnomalyAlert{
		EntityID:     entityID,
		Score:        zScore,
		IsAnomaly:    isAnomaly,
		Confidence:   confidence,
		DetectedAt:   time.Now(),
		MetricValues: map[string]float64{"amount": amount, "mean": mean, "stddev": stdDev},
	}

	if isAnomaly {
		d.logger.Warn("Anomaly detected",
			zap.String("entity_id", entityID),
			zap.Float64("amount", amount),
			zap.Float64("z_score", zScore),
			zap.Float64("mean", mean),
			zap.Float64("stddev", stdDev),
		)
	}

	return alert, nil
}

// UpdateModel updates an anomaly detection model
func (d *AnomalyDetector) UpdateModel(ctx context.Context, entityID string, samples []float64) error {
	d.modelsMu.Lock()
	defer d.modelsMu.Unlock()

	model, exists := d.models[entityID]
	if !exists {
		model = &AnomalyModel{
			EntityID:   entityID,
			WindowSize: 100,
			Samples:    make([]float64, 0, 100),
		}
		d.models[entityID] = model
	}

	model.Samples = append(model.Samples, samples...)
	if len(model.Samples) > model.WindowSize {
		model.Samples = model.Samples[len(model.Samples)-model.WindowSize:]
	}
	model.LastUpdated = time.Now()

	if len(model.Samples) > 0 {
		mean, stdDev := stats.MeanStdDev(model.Samples)
		model.Mean = mean
		model.StdDev = stdDev
	}

	return nil
}

// SetThreshold sets the anomaly detection threshold
func (d *AnomalyDetector) SetThreshold(threshold float64) {
	if threshold <= 0 {
		threshold = 3.0
	}
	d.threshold = threshold
}

// GetModelStats returns model statistics
func (d *AnomalyDetector) GetModelStats(entityID string) (mean, stdDev float64, sampleCount int64, err error) {
	d.modelsMu.RLock()
	defer d.modelsMu.RUnlock()

	model, exists := d.models[entityID]
	if !exists {
		return 0, 0, 0, fmt.Errorf("model not found for entity %s", entityID)
	}

	return model.Mean, model.StdDev, model.SampleCount, nil
}
