package models

type FPO struct {
	FpoRegNo string `gorm:"column:fpo_reg_no;primaryKey;size:50" json:"fpo_reg_no"`

	// Address fields
	AddressLine1 *string `gorm:"column:address_line1;type:varchar(255)" json:"address_line1,omitempty"`
	Village      *string `gorm:"column:village;type:varchar(100)" json:"village,omitempty"`
	Mandal       *string `gorm:"column:mandal;type:varchar(100)" json:"mandal,omitempty"`
	District     *string `gorm:"column:district;type:varchar(100)" json:"district,omitempty"`
	State        *string `gorm:"column:state;type:varchar(100)" json:"state,omitempty"`
	Pincode      *string `gorm:"column:pincode;type:varchar(10)" json:"pincode,omitempty"`

	// CBBO Name (if any)
	CbboName *string `gorm:"column:cbbo_name;type:varchar(100)" json:"cbbo_name,omitempty"`

	// Chairman details
	ChairmanName    *string `gorm:"column:chairman_name;type:varchar(100)" json:"chairman_name,omitempty"`
	ChairmanContact *string `gorm:"column:chairman_contact;type:varchar(50)" json:"chairman_contact,omitempty"`

	// Board of Directors details
	BoardOfDirectors []BoardMember `gorm:"foreignKey:FPORegNo;references:FpoRegNo" json:"board_of_directors,omitempty"`

	// Bank details
	BankName          *string `gorm:"column:bank_name;type:varchar(100)" json:"bank_name,omitempty"`
	AccountHolderName *string `gorm:"column:account_holder_name;type:varchar(100)" json:"account_holder_name,omitempty"`
	AccountNumber     *string `gorm:"column:account_number;type:varchar(50)" json:"account_number,omitempty"`
	IFSCCode          *string `gorm:"column:ifsc_code;type:varchar(20)" json:"ifsc_code,omitempty"`

	// FPO Logo (Could store URL or file path)
	FpoLogo *string `gorm:"column:fpo_logo;type:varchar(255)" json:"fpo_logo,omitempty"`

	// FPO Signature (Could store URL or file path)
	FpoSignature *string `gorm:"column:fpo_signature;type:varchar(255)" json:"fpo_signature,omitempty"`

	// Licenses
	GstNo             *string `gorm:"column:gst_no;type:varchar(50)" json:"gst_no,omitempty"`
	FertilizerLicense *string `gorm:"column:fertilizer_license;type:varchar(50)" json:"fertilizer_license,omitempty"`
	SeedLicense       *string `gorm:"column:seed_license;type:varchar(50)" json:"seed_license,omitempty"`
	PesticideLicense  *string `gorm:"column:pesticide_license;type:varchar(50)" json:"pesticide_license,omitempty"`

	// The CEO ID (already defined)
	CEOID *string `gorm:"column:ceo_id;type:varchar(36)" json:"ceo_id,omitempty"`
}

type BoardMember struct {
	Name     string `gorm:"column:name;type:varchar(100)" json:"name"`
	Contact  string `gorm:"column:contact;type:varchar(50)" json:"contact"`
	FPORegNo string `gorm:"column:fpo_reg_no;type:varchar(50)" json:"-"`
}
