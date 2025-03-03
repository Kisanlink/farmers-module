package repositories

import (
	"context"
	"github.com/Kisanlink/farmers-module/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
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
	log.Printf("DEBUG: Querying soil test reports for farmID: %d", farmID)

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

	log.Printf("DEBUG: Successfully retrieved %d soil test reports for farmID %d", len(reports), farmID)
	return reports, nil
}

// GetSoilTestReportsByFarmerID retrieves soil test reports for a given farmer ID
func (repo *SoilTestReportRepository) GetSoilTestReportsByFarmerID(ctx context.Context, farmerID int) ([]models.SoilTestReport, error) {
	log.Printf("DEBUG: Querying soil test reports for farmerID: %d", farmerID)

	cursor, err := repo.Collection.Find(ctx, bson.M{"farmerId": farmerID})
	if err != nil {
		log.Printf("ERROR: Failed to retrieve soil test reports for farmerID %d: %v", farmerID, err)
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

	log.Printf("DEBUG: Successfully retrieved %d soil test reports for farmerID %d", len(reports), farmerID)
	return reports, nil
}

// GetAllSoilTestReports retrieves all soil test reports
func (repo *SoilTestReportRepository) GetAllSoilTestReports(ctx context.Context) ([]models.SoilTestReport, error) {
	log.Printf("DEBUG: Querying all soil test reports")

	cursor, err := repo.Collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("ERROR: Failed to retrieve soil test reports: %v", err)
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

	log.Printf("DEBUG: Successfully retrieved %d soil test reports", len(reports))
	return reports, nil
}
