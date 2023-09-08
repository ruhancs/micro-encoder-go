package repositories

import (
	"encoder/domain"
	"fmt"

	"github.com/jinzhu/gorm"
)

type JobRepository interface {
	Insert(job *domain.Job) (*domain.Job,error)
	Find(id string) (*domain.Job)
	Update(job *domain.Job) (*domain.Job, error)
}

type JobRepositoryDb struct {
	Db *gorm.DB
}

func (repo *JobRepositoryDb) Insert(job *domain.Job) (*domain.Job,error) {
	err := repo.Db.Create(job).Error
	if err != nil {
		return nil, err
	}

	return job, nil
}

func(repo *JobRepositoryDb) Find(id string) (*domain.Job, error) {
	var job domain.Job
	repo.Db.Preload("Video").First(&job, "id=?", id)//ao achar o job no db preenche a variavel job com os valores

	if job.ID == ""{
		return nil, fmt.Errorf("job dnot found")
	}

	return &job, nil
}

func (repo *JobRepositoryDb) Update(job *domain.Job) (*domain.Job,error) {
	err := repo.Db.Save(&job).Error
	if err != nil {
		return nil, err
	}

	return job,err
}