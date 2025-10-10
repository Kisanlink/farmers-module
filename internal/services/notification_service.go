package services

import (
	"context"
	"fmt"
	"time"

	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
)

// NotificationChannel represents the type of notification channel
type NotificationChannel string

const (
	ChannelEmail   NotificationChannel = "email"
	ChannelWebhook NotificationChannel = "webhook"
	ChannelInApp   NotificationChannel = "in_app"
	ChannelSMS     NotificationChannel = "sms"
)

// NotificationPriority represents the priority level of a notification
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "LOW"
	PriorityMedium   NotificationPriority = "MEDIUM"
	PriorityHigh     NotificationPriority = "HIGH"
	PriorityCritical NotificationPriority = "CRITICAL"
)

// NotificationRequest represents a notification to be sent
type NotificationRequest struct {
	RecipientID   string // AAA user ID or org ID
	RecipientType string // "user" or "org"
	Channel       NotificationChannel
	Priority      NotificationPriority
	Subject       string
	Message       string
	Data          map[string]interface{} // Additional structured data
	TemplateID    string                 // Optional template ID
}

// NotificationService handles sending notifications to users and admins
type NotificationService interface {
	// SendOrphanedLinkAlert sends an alert when farmer_links become ORPHANED
	SendOrphanedLinkAlert(ctx context.Context, fpoOrgID string, orphanedLinks []*farmerentity.FarmerLink) error

	// SendDataQualityAlert sends alerts for data quality issues
	SendDataQualityAlert(ctx context.Context, alert DataQualityAlert) error

	// SendNotification sends a generic notification
	SendNotification(ctx context.Context, req *NotificationRequest) error

	// QueueNotification queues a notification for async delivery
	QueueNotification(ctx context.Context, req *NotificationRequest) error
}

// DataQualityAlert represents a data quality alert
type DataQualityAlert struct {
	OrgID       string
	AlertType   string
	Severity    NotificationPriority
	Title       string
	Description string
	AffectedIDs []string
	Timestamp   time.Time
}

// NotificationServiceImpl implements NotificationService
type NotificationServiceImpl struct {
	aaaService AAAService
	queue      chan *NotificationRequest
}

// NewNotificationService creates a new notification service
func NewNotificationService(aaaService AAAService) NotificationService {
	service := &NotificationServiceImpl{
		aaaService: aaaService,
		queue:      make(chan *NotificationRequest, 1000),
	}

	// Start background worker for async notifications
	go service.processQueue()

	return service
}

// SendOrphanedLinkAlert sends an alert when farmer_links become ORPHANED
// Business Rule 6.2: Notify FPO admin of data inconsistency
func (s *NotificationServiceImpl) SendOrphanedLinkAlert(ctx context.Context, fpoOrgID string, orphanedLinks []*farmerentity.FarmerLink) error {
	if len(orphanedLinks) == 0 {
		return nil
	}

	// Build notification message
	subject := fmt.Sprintf("Data Inconsistency Alert: %d Farmer Links Orphaned", len(orphanedLinks))
	message := fmt.Sprintf(`
Data Inconsistency Detected

Organization: %s
Orphaned Farmer Links: %d

%d farmer link(s) have been marked as ORPHANED due to missing references in the AAA (Authentication, Authorization & Accounting) system. This typically occurs when:
- A user account was deleted in the AAA admin panel
- An organization was removed from the AAA system
- External data modifications were made outside the normal workflow

Affected Links:
`, fpoOrgID, len(orphanedLinks), len(orphanedLinks))

	affectedIDs := make([]string, 0, len(orphanedLinks))
	for i, link := range orphanedLinks {
		if i < 10 { // Show first 10 in message
			message += fmt.Sprintf("- Farmer User ID: %s (Link ID: %s)\n", link.AAAUserID, link.ID)
		}
		affectedIDs = append(affectedIDs, link.ID)
	}

	if len(orphanedLinks) > 10 {
		message += fmt.Sprintf("\n... and %d more\n", len(orphanedLinks)-10)
	}

	message += `
Action Required:
1. Review the affected farmer links in the system
2. Either restore the missing AAA references or delete the local records
3. Use the manual resolution workflow to clean up orphaned data

For assistance, please contact your system administrator.
	`

	// Create notification request
	notificationReq := &NotificationRequest{
		RecipientID:   fpoOrgID,
		RecipientType: "org",
		Channel:       ChannelEmail, // Default to email, can be configured
		Priority:      PriorityHigh,
		Subject:       subject,
		Message:       message,
		Data: map[string]interface{}{
			"alert_type":        "ORPHANED_LINKS",
			"org_id":            fpoOrgID,
			"orphaned_count":    len(orphanedLinks),
			"affected_link_ids": affectedIDs,
			"timestamp":         time.Now().Format(time.RFC3339),
		},
	}

	// Queue for async delivery to avoid blocking reconciliation
	return s.QueueNotification(ctx, notificationReq)
}

// SendDataQualityAlert sends alerts for data quality issues
func (s *NotificationServiceImpl) SendDataQualityAlert(ctx context.Context, alert DataQualityAlert) error {
	notificationReq := &NotificationRequest{
		RecipientID:   alert.OrgID,
		RecipientType: "org",
		Channel:       ChannelEmail,
		Priority:      alert.Severity,
		Subject:       alert.Title,
		Message:       alert.Description,
		Data: map[string]interface{}{
			"alert_type":   alert.AlertType,
			"affected_ids": alert.AffectedIDs,
			"timestamp":    alert.Timestamp.Format(time.RFC3339),
		},
	}

	return s.QueueNotification(ctx, notificationReq)
}

// SendNotification sends a notification synchronously
func (s *NotificationServiceImpl) SendNotification(ctx context.Context, req *NotificationRequest) error {
	// TODO: Implement actual notification delivery
	// This should integrate with:
	// - Email service (SendGrid, AWS SES, etc.)
	// - Webhook delivery
	// - In-app notification system
	// - SMS gateway

	// For now, just log the notification
	fmt.Printf("[NOTIFICATION] Channel=%s Priority=%s Recipient=%s Subject=%s\n",
		req.Channel, req.Priority, req.RecipientID, req.Subject)

	// In production, this would call external services:
	// switch req.Channel {
	// case ChannelEmail:
	//     return s.sendEmail(ctx, req)
	// case ChannelWebhook:
	//     return s.sendWebhook(ctx, req)
	// case ChannelInApp:
	//     return s.createInAppNotification(ctx, req)
	// case ChannelSMS:
	//     return s.sendSMS(ctx, req)
	// }

	return nil
}

// QueueNotification queues a notification for async delivery
func (s *NotificationServiceImpl) QueueNotification(ctx context.Context, req *NotificationRequest) error {
	select {
	case s.queue <- req:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue is full, send synchronously as fallback
		return s.SendNotification(ctx, req)
	}
}

// processQueue processes notifications from the queue asynchronously
func (s *NotificationServiceImpl) processQueue() {
	for req := range s.queue {
		ctx := context.Background()
		if err := s.SendNotification(ctx, req); err != nil {
			// Log error but don't block the queue
			fmt.Printf("[NOTIFICATION ERROR] Failed to send notification: %v\n", err)
			// TODO: Implement retry logic or dead letter queue
		}
	}
}
