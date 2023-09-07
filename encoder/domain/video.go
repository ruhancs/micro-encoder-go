package domain

import (
	"time"
	"github.com/asaskevich/govalidator"
)

type Video struct {
	ID string `json:"encoded_video_folder" valid:"uuid" gorm:"type:uuid;primary_key"` //indica para o go validator qual é o tipo do campo
	//identificacao do servico que enviou o video ResourceID
	ResourceID string `json:"resource_id" valid:"notnull" gorm:"type:varchar(255)"`
	FilePath string `json:"file_path" valid:"notnull" gorm:"type:varchar(255)"`
	CreatedAt time.Time `json:"-" valid:"-"`
	Jobs []*Job `json:"-" valid:"-" gorm:"ForeignKey:VideoID"`// VideoID da entidade job 
}

//roda antes do que qualquer coisa no go
func init() {
	//habilitar a validacao da struct antes de tudo
	govalidator.SetFieldsRequiredByDefault(true)
}

func NewVideo() *Video {
	return &Video{}
}

func (video *Video) Validate() error {
	_,err := govalidator.ValidateStruct(video)
	if err != nil {
		return err
	}

	return nil
}