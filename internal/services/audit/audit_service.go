package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuditService handles audit logging
type AuditService struct {
	logger *zap.Logger
	queue  chan *AuditEvent
	client AuditClient
}

// AuditEvent represents an audit log event
type AuditEvent struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	UserID        string                 `json:"user_id"`
	OrgID         string                 `json:"org_id"`
	Action        string                 `json:"action"`
	ResourceType  string                 `json:"resource_type"`
	ResourceID    string                 `json:"resource_id"`
	OldValue      interface{}            `json:"old_value,omitempty"`
	NewValue      interface{}            `json:"new_value,omitempty"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	CorrelationID string                 `json:"correlation_id"`
	Status        string                 `json:"status"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AuditFilters represents filters for querying audit events
type AuditFilters struct {
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	UserID       string     `json:"user_id,omitempty"`
	Action       string     `json:"action,omitempty"`
	ResourceType string     `json:"resource_type,omitempty"`
	ResourceID   string     `json:"resource_id,omitempty"`
	Status       string     `json:"status,omitempty"`
	Page         int        `json:"page,omitempty"`
	PageSize     int        `json:"page_size,omitempty"`
}

// AuditClient interface for remote audit service
type AuditClient interface {
	SendAuditEvent(ctx context.Context, event *AuditEvent) error
	SendAuditEventBatch(ctx context.Context, events []*AuditEvent) error
	QueryAuditEvents(ctx context.Context, filters *AuditFilters) ([]*AuditEvent, error)
}

// NewAuditService creates a new audit service
func NewAuditService(logger *zap.Logger, client AuditClient) *AuditService {
	svc := &AuditService{
		logger: logger,
		queue:  make(chan *AuditEvent, 1000),
		client: client,
	}

	// Start background worker
	go svc.processQueue()

	return svc
}

// LogEvent logs an audit event asynchronously
func (s *AuditService) LogEvent(ctx context.Context, event *AuditEvent) error {
	// Add context metadata
	if event.ID == "" {
		event.ID = generateUUID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Send to queue for async processing
	select {
	case s.queue <- event:
		s.logger.Debug("Audit event queued", zap.String("event_id", event.ID))
		return nil
	default:
		// Queue is full, log synchronously
		return s.logEventSync(ctx, event)
	}
}

// LogEventSync logs an audit event synchronously
func (s *AuditService) logEventSync(ctx context.Context, event *AuditEvent) error {
	// Log to external service if client is available
	if s.client != nil {
		if err := s.client.SendAuditEvent(ctx, event); err != nil {
			s.logger.Error("Failed to send audit event",
				zap.Error(err),
				zap.String("event_id", event.ID))
			// Fall back to local logging
			return s.logToFile(event)
		}
	} else {
		// No client available, log locally
		return s.logToFile(event)
	}

	return nil
}

// processQueue processes audit events from the queue in batches
func (s *AuditService) processQueue() {
	batch := make([]*AuditEvent, 0, 100)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event := <-s.queue:
			batch = append(batch, event)

			if len(batch) >= 100 {
				s.flushBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch sends a batch of audit events
func (s *AuditService) flushBatch(events []*AuditEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.client != nil {
		if err := s.client.SendAuditEventBatch(ctx, events); err != nil {
			s.logger.Error("Failed to send audit batch",
				zap.Error(err),
				zap.Int("batch_size", len(events)))

			// Fall back to individual logging
			for _, event := range events {
				_ = s.logToFile(event)
			}
			return
		}
	} else {
		// No client available, log locally
		for _, event := range events {
			_ = s.logToFile(event)
		}
	}

	s.logger.Debug("Audit batch processed successfully",
		zap.Int("batch_size", len(events)))
}

// logToFile logs audit event to local logger
func (s *AuditService) logToFile(event *AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("Failed to marshal audit event", zap.Error(err))
		return err
	}

	s.logger.Info("AUDIT",
		zap.String("event", string(data)),
		zap.String("event_id", event.ID),
		zap.String("action", event.Action),
		zap.String("user_id", event.UserID),
		zap.String("resource_type", event.ResourceType),
		zap.String("resource_id", event.ResourceID))

	return nil
}

// QueryAuditTrail queries audit events with filters
func (s *AuditService) QueryAuditTrail(ctx context.Context, filters *AuditFilters) ([]*AuditEvent, error) {
	if s.client != nil {
		return s.client.QueryAuditEvents(ctx, filters)
	}

	// Return empty result if no client (in production, this should query local storage)
	s.logger.Warn("No audit client available for querying audit trail")
	return []*AuditEvent{}, nil
}

// CreateEvent creates a new audit event with basic information
func (s *AuditService) CreateEvent(userID, orgID, action, resourceType, resourceID string) *AuditEvent {
	return &AuditEvent{
		ID:           generateUUID(),
		Timestamp:    time.Now(),
		UserID:       userID,
		OrgID:        orgID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Status:       "SUCCESS",
		Metadata:     make(map[string]interface{}),
	}
}

// generateUUID generates a new UUID
func generateUUID() string {
	return uuid.New().String()
}
