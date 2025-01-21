package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// AreaManager represents the area manager model in MongoDB
type AreaManager struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`  // MongoDB _id
	Name      string             `bson:"areaManagerName"`
	CreatedAt time.Time          `bson:"createdAt"`
	ModifiedAt time.Time         `bson:"modifiedAt"`
	IsDeleted bool               `bson:"isDeleted"`
	Class     string             `bson:"_class"`
}

