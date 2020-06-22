package models

import "github.com/jinzhu/gorm"

type Image struct {
	gorm.Model
	Name  string `json:"name"`
	S3Url string `json:"s3url"`
}

func (image *Image) Create() *Image {
	GetDb().Create(image)

	return image
}

func GetImage(u uint) *Image {
	image := &Image{}
	GetDb().Table("images").Where("id = ?", u).First(image)
	return image
}