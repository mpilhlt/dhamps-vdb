package models

type LLMProcessRequest struct {
  ServiceID  string   `json:"serviceId"`
  ProjectID  string   `json:"projectId"`
  ContextID  string   `json:"contextId"`
  TextFields []string `json:"textFields"`
}

type LLMProcessResponse struct {
  TextFields []string `json:"textFields"`
}
