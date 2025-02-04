package repositories

import (
	"context"
	"log"

	"github.com/Kisanlink/farmers-module/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SoilTestReportRepository struct {
	Collection *mongo.Collection
}

func NewSoilTestReportRepository(db *mongo.Database) *SoilTestReportRepository {
	return &SoilTestReportRepository{
		Collection: db.Collection("SoilTestReports"),
	}
}

// GetSoilTestReports retrieves soil test reports by farmID
func (repo *SoilTestReportRepository) GetSoilTestReports(ctx context.Context, farmID int) ([]models.SoilTestReport, error) {
	log.Printf("DEBUG: Fetching soil test reports for farmID: %d", farmID)

	cursor, err := repo.Collection.Find(ctx, bson.M{"farmId": farmID})
	if err != nil {
		log.Printf("ERROR: Failed to retrieve soil test reports for farmID %d: %v", farmID, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var reports []models.SoilTestReport
	for cursor.Next(ctx) {
		var report models.SoilTestReport
		if err := cursor.Decode(&report); err != nil {
			log.Printf("ERROR: Failed to decode soil test report: %v", err)
			return nil, err
		}
		reports = append(reports, report)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("ERROR: Cursor error while fetching soil test reports: %v", err)
		return nil, err
	}

	log.Printf("DEBUG: Successfully retrieved %d soil test reports for farmID %d", len(reports), farmID)
	return reports, nil
}
