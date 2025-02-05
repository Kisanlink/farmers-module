package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Order model definition to match MongoDB document structure
type Order struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	OrderID            int32              `bson:"orderId" json:"orderId"`
	OrderStatus        string             `bson:"orderStatus" json:"orderStatus"`
	OrderDate          OrderDate          `bson:"orderDate" json:"orderDate"`
	IncentiveAmount    float64            `bson:"incentiveAmount" json:"incentiveAmount"`
	DeliveryFee        float64            `bson:"deliveryFee" json:"deliveryFee"`
	SubTotal           float64            `bson:"subTotal" json:"subTotal"`
	TotalAmount        float64            `bson:"totalAmount" json:"totalAmount"`
	TotalTax           float64            `bson:"totalTax" json:"totalTax"`
	TotalMRP           float64            `bson:"totalMrp" json:"totalMrp"`
	FarmerName         string             `bson:"farmerName" json:"farmerName"`
	FarmerID           int32              `bson:"farmerId" json:"farmerId"`
	Location           string             `bson:"location" json:"location"`
	MobileNumber       int64              `bson:"mobileNumber" json:"mobileNumber"`
	FarmID             int32              `bson:"farmId" json:"farmId"`
	KisanSathiID       int32              `bson:"kisanSathiId" json:"kisanSathiId"`
	AreaManagerID      string             `bson:"areaManagerId" json:"areaManagerId"`
	AreaManagerName    string             `bson:"areaManagerName" json:"areaManagerName"`
	OrderType          string             `bson:"orderType" json:"orderType"`
	FarmerNumber       int64              `bson:"farmerNumber" json:"farmerNumber"`
	LandingPrice       float64            `bson:"landingPrice" json:"landingPrice"`
	GrossMargin        float64            `bson:"grossMargin" json:"grossMargin"`
	PaymentMode        string             `bson:"paymentMode" json:"paymentMode"`
	PaymentDone        bool               `bson:"paymentDone" json:"paymentDone"`
	WithInterestAmount float64            `bson:"withInterestAmount" json:"withInterestAmount"`
	Interest           float64            `bson:"interest" json:"interest"`
	IsCashBackUsed     bool               `bson:"isCashBackUsed" json:"isCashBackUsed"`
	CashBackAmountUsed float64            `bson:"cashBackAmountUsed" json:"cashBackAmountUsed"`
	CustomerType       string             `bson:"customerType" json:"customerType"`
	ShippingAddress    string             `bson:"shippingAddress" json:"shippingAddress"`
	CollaboratorID     int32              `bson:"collaboratorId" json:"collaboratorId"`
	CreatedBy          string             `bson:"createdBy" json:"createdBy"`
	CreatedAt          OrderDate          `bson:"createdAt" json:"createdAt"`
	ModifiedBy         string             `bson:"modifiedBy" json:"modifiedBy"`
	ModifiedAt         OrderDate          `bson:"modifiedAt" json:"modifiedAt"`
	IsDeleted          bool               `bson:"isDeleted" json:"isDeleted"`
	Class              string             `bson:"_class" json:"_class"`
}

// OrderDate struct to handle the custom date format in MongoDB
type OrderDate struct {
	Year             int32  `bson:"year" json:"year"`
	Month            int32  `bson:"month" json:"month"`
	Day              int32  `bson:"day" json:"day"`
	Hour             int32  `bson:"hour" json:"hour"`
	Minute           int32  `bson:"minute" json:"minute"`
	Second           int32  `bson:"second" json:"second"`
	FractionalSecond string `bson:"fractionalSecond" json:"fractionalSecond"`
	Timezone         int32  `bson:"timezone" json:"timezone"`
}
