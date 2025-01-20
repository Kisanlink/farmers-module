package models

import(
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FPO represents the Farmer Producer Organization (FPO) details.
type FPO struct {
	Name          string  ⁠ json:"name" ⁠            // Name of the FPO
	NumberOfShares int     ⁠ json:"number_of_shares" ⁠ // Number of shares allotted to the farmer
	ShareValue    float64 ⁠ json:"share_value" ⁠     // Value of each share (e.g., 100 per share)
}

// Farmer represents a farmer's profile with personal details, Kisansathi, and FPO information.
type Farmer struct {
	ID            primitive.ObjectID ⁠ json:"id,omitempty" bson:"_id,omitempty" ⁠ // MongoDB _id field
	Name          string    ⁠ json:"name" ⁠            // Name of the farmer
	FatherName    string    ⁠ json:"father_name" ⁠     // Father's name of the farmer
	DateOfBirth   time.Time ⁠ json:"dob" ⁠             // Date of birth (DOB)
	Age           int       ⁠ json:"age" ⁠             // Age of the farmer (calculated or entered)
	Gender        string    ⁠ json:"gender" ⁠          // Gender of the farmer
	Address       string    ⁠ json:"address" ⁠         // Address of the farmer
	ContactNumber string    ⁠ json:"contact_number" ⁠  // Contact number
	Acres         float64   ⁠ json:"acres" ⁠           // Acres of land owned by the farmer
	Kisansathi    string    ⁠ json:"kisansathi" ⁠      // Link or reference to the Kisansathi profile
	FPO           FPO       ⁠ json:"fpo" ⁠             // FPO details (name, shares, share value)
}

// CalculateSharesValue calculates the total value of the farmer's shares.
func (f *Farmer) CalculateSharesValue() float64 {
	return float64(f.FPO.NumberOfShares) * f.FPO.ShareValue
}

// CalculateAge calculates the farmer's age based on the date of birth.
func (f *Farmer) CalculateAge() int {
	currentYear := time.Now().Year()
	birthYear := f.DateOfBirth.Year()
	age := currentYear - birthYear
	// If the farmer's birthday hasn't occurred yet this year, subtract one year
	if time.Now().Before(f.DateOfBirth.AddDate(age, 0, 0)) {
		age--
	}
	return age
}