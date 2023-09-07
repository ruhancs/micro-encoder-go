package domain

import (
	"time"
	"github.com/asaskevich/govalidator"
)

type Video struct {
	ID string `valid:"uuid"` //indica para o go validator qual é o tipo do campo
	//identificacao do servico que enviou o video ResourceID
	ResourceID string `valid:"notnull"`
	FilePath string `valid:"notnull"`
	CreatedAt time.Time `valid:"-"`
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