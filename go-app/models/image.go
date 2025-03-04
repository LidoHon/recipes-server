package models

type ImageInput struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Base64String string `json:"base64String"`
}
