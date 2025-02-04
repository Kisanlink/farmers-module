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

// GetSoilTestReportsByFarmID retrieves soil test reports for a given farm ID
func (repo *SoilTestReportRepository) GetSoilTestReportsByFarmID(ctx context.Context, farmID int) ([]models.SoilTestReport, error) {
	// Debug log
	log.Printf("DEBUG: Querying soil test reports for farmID: %d", farmID)

	// Query soil test reports by farm ID
	cursor, err := repo.Collection.Find(ctx, bson.M{"farmId": farmID})
	if err != nil {
		log.Printf("ERROR: Failed to retrieve soil test reports for farmID %d: %v", farmID, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	// Parse results
	var reports []models.SoilTestReport
	for cursor.Next(ctx) {
		var report models.SoilTestReport
		if err := cursor.Decode(&report); err != nil {
			log.Printf("ERROR: Failed to decode soil test report: %v", err)
			return nil, err
		}
		reports = append(reports, report)
	}

	// Debug log
	log.Printf("DEBUG: Successfully retrieved %d soil test reports for farmID %d", len(reports), farmID)

	return reports, nil
}
