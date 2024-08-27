package models

type SimilarQuery struct {
  Count     int     `json:"count"`
  Threshold float64 `json:"threshold"`
}

type SimilarRequest struct {
  Count     int     `json:"count"`
  Threshold float64 `json:"threshold"`
}

type SimilarResponse struct {
  Count     int     `json:"count"`
  Threshold float64 `json:"threshold"`
}
