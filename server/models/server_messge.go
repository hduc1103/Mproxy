package models

type ServerMessage struct {
	DeviceId string `json:"device_id"`
	Message  string `json:"message"`
}