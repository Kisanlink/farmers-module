package models

import (
	"database/sql"
	"fmt"

	"github.com/Kisanlink/farmers-module/entities"
	pb "github.com/kisanlink/protobuf/pb-aaa"
	"gorm.io/gorm"
)

// FarmerSignupRequest defines the request structure for farmer registration
type FarmerSignupRequest struct {
	CountryCode        string  `json:"country_code"                 validate:"omitempty,numeric,len=2"`
	MobileNumberString string  `json:"mobile_number"                validate:"required,numeric,len=10"`
	MobileNumber       uint64  `json:"-"`
	AadhaarNumber      *string `json:"aadhar_number,omitempty"   validate:"omitempty,numeric,len=12"`
	UserName           *string `json:"username,omitempty"            validate:"omitempty,min=2,max=100"`

	Gender         string `json:"gender,omitempty"            validate:"omitempty,oneof=male female other"`
	FatherName     string `json:"father_name,omitempty"`
	SocialCategory string `json:"social_category,omitempty"`
	EquityShare    string `json:"equity_share,omitempty"      validate:"omitempty,numeric"`
	TotalShare     string `json:"total_share,omitempty"       validate:"omitempty,numeric"`
	AreaType       string `json:"area_type,omitempty"`

	IsFPO    bool   `json:"isFPO"`
	State    string `json:"state,omitempty"`
	District string `json:"district,omitempty"`
	Block    string `json:"block,omitempty"`
	IaName   string `json:"iaName,omitempty"`
	CbbName  string `json:"cbbName,omitempty"`
	FpoName  string `json:"fpoName,omitempty"`
	FpoRegNo string `json:"fpoRegNo,omitempty"`

	UserId           *string `json:"user_id,omitempty"            validate:"omitempty,uuid"`
	KisansathiUserId *string `json:"kisansathi_user_id,omitempty" validate:"omitempty,uuid"`
	Type             string  `json:"type,omitempty"`
}

type Farmer struct {
	Base

	UserId           string  `gorm:"column:user_id;type:varchar(36);uniqueIndex" json:"user_id"`
	KisansathiUserId *string `gorm:"column:kisansathi_user_id;type:varchar(36)"  json:"kisansathi_user_id,omitempty"`

	// ─── profile ───────────────────────────────────────────────────────────
	Gender         string `gorm:"column:gender;type:varchar(6)"           json:"gender"`
	SocialCategory string `gorm:"column:social_category;type:varchar(10)" json:"social_category"`
	FatherName     string `gorm:"column:father_name;type:varchar(100)"    json:"father_name"`
	EquityShare    string `gorm:"column:equity_share;type:varchar(10)"    json:"equity_share"`
	TotalShare     string `gorm:"column:total_share;type:varchar(10)"     json:"total_share"`
	AreaType       string `gorm:"column:area_type;type:varchar(10)"       json:"area_type"`

	// ─── FPO details ───────────────────────────────────────────────────────
	IsFPO    bool           `gorm:"column:is_fpo;default:false" json:"is_fpo"`
	State    sql.NullString `gorm:"column:state"                json:"state,omitempty"`
	District sql.NullString `gorm:"column:district"             json:"district,omitempty"`
	Block    sql.NullString `gorm:"column:block"                json:"block,omitempty"`
	IaName   sql.NullString `gorm:"column:ia_name"              json:"ia_name,omitempty"`
	CbbName  sql.NullString `gorm:"column:cbb_name"             json:"cbb_name,omitempty"`
	FpoName  sql.NullString `gorm:"column:fpo_name"             json:"fpo_name,omitempty"`
	FpoRegNo sql.NullString `gorm:"column:fpo_reg_no"           json:"fpo_reg_no,omitempty"`

	// ─── misc flags ────────────────────────────────────────────────────────
	IsActive     bool                `gorm:"column:is_active;default:true"   json:"is_active"`
	IsSubscribed bool                `gorm:"column:is_subscribed;default:false" json:"is_subscribed"`
	Type         entities.FarmerType `gorm:"column:type;type:varchar(10);not null;default:'OTHER'" json:"type"`

	UserDetails *pb.User `json:"user_details,omitempty" gorm:"-"`
}

// wipe FPO columns when IsFPO=false
func (f *Farmer) BeforeSave(tx *gorm.DB) (err error) {
	if !f.IsFPO {
		f.State, f.District, f.Block,
			f.IaName, f.CbbName, f.FpoName, f.FpoRegNo = sql.NullString{}, sql.NullString{},
			sql.NullString{}, sql.NullString{}, sql.NullString{}, sql.NullString{}, sql.NullString{}
	}
	return nil
}

func (f *Farmer) BeforeCreate(tx *gorm.DB) (err error) {
	if err = f.Base.BeforeCreate(tx); err != nil {
		return err
	}
	if !entities.FARMER_TYPES.IsValid(string(f.Type)) {
		return fmt.Errorf("invalid farmer type: %s (valid: %v)",
			f.Type, entities.FARMER_TYPES.StringValues())
	}
	return nil
}
