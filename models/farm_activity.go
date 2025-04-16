package models

import (
	"time"

	"gorm.io/gorm"
)

type ActivityType string

const (
	ActivityTILLING     ActivityType = "TILLING"
	ActivitySOWING      ActivityType = "SOWING"
	ActivityFERTILISING ActivityType = "FERTILISING"
	ActivityWEEDING     ActivityType = "WEEDING"
	ActivityHARVESTING  ActivityType = "HARVESTING"
	ActivityCLEANING    ActivityType = "CLEANING"
	ActivityRESTARTING  ActivityType = "RESTARTING"
	ActivityIRRIGATION  ActivityType = "IRRIGATION"
	ActivitySOILTESTING ActivityType = "SOIL-TESTING"
)

type FarmActivity struct {
	Base
	FarmID         string       `json:"farm_id" gorm:"type:varchar(10);not null"`
	CropCycleID    string       `json:"crop_cycle_id" gorm:"type:varchar(10);not null"`
	Activity       ActivityType `json:"activity" gorm:"type:varchar(150);default:'SOWING'"`
	StartDate      *time.Time   `json:"start_date"`
	EndDate        *time.Time   `json:"end_date"`
	ActivityReport string       `json:"activity_report" gorm:"type:text"`

	CropCycle CropCycle `json:"crop_cycle" gorm:"foreignKey:CropCycleID"`
}

func (f *FarmActivity) BeforeCreate(tx *gorm.DB) (err error) {
	if f.Activity == "" {
		f.Activity = ActivitySOWING
	}
	return
}
