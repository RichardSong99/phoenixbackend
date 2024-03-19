package datacube

import "go.mongodb.org/mongo-driver/bson/primitive"

type Cell struct {
	Values map[string]*float64
	// Add more fields as needed
}

type DataCube struct {
	UserID primitive.ObjectID `json:"UserID" bson:"user_id"`
	Rows   map[string]Row
}

type Row struct {
	Cells map[string]Cell
}
