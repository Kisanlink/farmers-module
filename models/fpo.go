package models

type FPO struct {
	FpoRegNo string `gorm:"column:fpo_reg_no;primaryKey;size:50" json:"fpo_reg_no"`

	State    string `gorm:"column:state"    json:"state,omitempty"`
	District string `gorm:"column:district" json:"district,omitempty"`
	Block    string `gorm:"column:block"    json:"block,omitempty"`
	IaName   string `gorm:"column:ia_name"  json:"ia_name,omitempty"`
	CbbName  string `gorm:"column:cbb_name" json:"cbb_name,omitempty"`
	FpoName  string `gorm:"column:fpo_name" json:"fpo_name,omitempty"`
}
