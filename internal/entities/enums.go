package entities

// Season represents the agricultural season
type Season string

const (
	SeasonUnspecified Season = "UNSPECIFIED"
	SeasonRabi        Season = "RABI"
	SeasonKharif      Season = "KHARIF"
	SeasonZaid        Season = "ZAID"
	SeasonPerennial   Season = "PERENNIAL"
	SeasonOther       Season = "OTHER"
)

// CycleStatus represents the status of a crop cycle
type CycleStatus string

const (
	CycleStatusUnspecified CycleStatus = "UNSPECIFIED"
	CycleStatusPlanned     CycleStatus = "PLANNED"
	CycleStatusActive      CycleStatus = "ACTIVE"
	CycleStatusCompleted   CycleStatus = "COMPLETED"
	CycleStatusCancelled   CycleStatus = "CANCELLED"
)

// ActivityType represents the type of farm activity
type ActivityType string

const (
	ActivityTypeUnspecified ActivityType = "UNSPECIFIED"
	ActivityTypePlanting    ActivityType = "PLANTING"
	ActivityTypeFertilizing ActivityType = "FERTILIZING"
	ActivityTypeIrrigation  ActivityType = "IRRIGATION"
	ActivityTypePestControl ActivityType = "PEST_CONTROL"
	ActivityTypeHarvesting  ActivityType = "HARVESTING"
	ActivityTypeOther       ActivityType = "OTHER"
)

// Resource represents the resource type for authorization
type Resource string

const (
	ResourceFarmer       Resource = "farmer"
	ResourceFarm         Resource = "farm"
	ResourceCropCycle    Resource = "crop_cycle"
	ResourceFarmActivity Resource = "farm_activity"
	ResourceFPORef       Resource = "fpo_ref"
)

// Action represents the action type for authorization
type Action string

const (
	ActionCreate   Action = "create"
	ActionRead     Action = "read"
	ActionUpdate   Action = "update"
	ActionDelete   Action = "delete"
	ActionList     Action = "list"
	ActionAssign   Action = "assign"
	ActionStart    Action = "start"
	ActionEnd      Action = "end"
	ActionComplete Action = "complete"
)

// Role represents the user role for authorization
type Role string

const (
	RoleFarmer         Role = "FARMER"
	RoleKisanSathi     Role = "KISAN_SATHI"
	RoleFPOCEO         Role = "FPO_CEO"
	RoleFPODirector    Role = "FPO_DIRECTOR"
	RoleFPOShareholder Role = "FPO_SHAREHOLDER"
)
