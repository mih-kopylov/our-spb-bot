package spb

import "gopkg.in/yaml.v3"

type ErrorResponse map[string]any

func (r *ErrorResponse) String() string {
	out, _ := yaml.Marshal(r)
	return string(out)
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type NearestBuildingResponse struct {
	Buildings []BuildingResponse `json:"buildings"`
}

type BuildingResponse struct {
	Id        int64   `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"full_address"`
}

type CityResponse struct {
	Id         int64              `json:"id"`
	Name       string             `json:"name"`
	Categories []CategoryResponse `json:"categories"`
}

type CategoryResponse struct {
	Id      int64            `json:"id"`
	Name    string           `json:"name"`
	Reasons []ReasonResponse `json:"reasons"`
}

type ReasonResponse struct {
	Id           int64        `json:"id"`
	Name         string       `json:"name"`
	PositionType PositionType `json:"wizard_widget"`
}

type PositionType int

const (
	PositionTypeBuilding     PositionType = 1
	PositionTypeStreet       PositionType = 2
	PositionTypeNearBuilding PositionType = 4
	//PositionTypeNearBuilding2 I have no idea why there are two types for the same thing
	PositionTypeNearBuilding2 PositionType = 5
)

type CreateProblemRequest interface {
}

type CreateStreetProblemRequest struct {
	Body                   string  `json:"body"`
	ReasonId               int64   `json:"reason"`
	Latitude               float64 `json:"latitude"`
	Longitude              float64 `json:"longitude"`
	ManuallySelectedReason bool    `json:"manually_selected_reason"`
}

type CreateBuildingProblemRequest struct {
	CreateStreetProblemRequest
	Building int64 `json:"building"`
}

type CreateNearBuildingProblemRequest struct {
	CreateStreetProblemRequest
	NearestBuilding int64 `json:"nearest_building"`
}

type Message struct {
	CategoryId  int
	FileUrls    []string
	Description string
}

type SentMessageResponse struct {
	Id string `json:"id"`
}
