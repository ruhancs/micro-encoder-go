package repositories_test

import (
	"encoder/application/repositories"
	"encoder/domain"
	"encoder/framework/database"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestJobRepositoryInsert(t *testing.T) {
	db := database.NewDbTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repoVideo := repositories.VideoRepositoryDb{Db: db}
	repoVideo.Insert(video)

	job,err := domain.NewJob("Output path", "pending", video)
	require.Nil(t,err)
	
	repoJob := repositories.JobRepositoryDb{Db: db}
	repoJob.Insert(job)

	jobFounded, err := repoJob.Find(job.ID)
	require.NotEmpty(t,jobFounded.ID)
	require.Nil(t,err)
	require.Equal(t,jobFounded.ID,job.ID)
	require.Equal(t,jobFounded.VideoID,job.VideoID)
}

func TestJobRepositoryUpdate(t *testing.T) {
	db := database.NewDbTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repoVideo := repositories.VideoRepositoryDb{Db: db}
	repoVideo.Insert(video)

	job,err := domain.NewJob("Output path", "pending", video)
	require.Nil(t,err)
	
	repoJob := repositories.JobRepositoryDb{Db: db}
	repoJob.Insert(job)

	job.Status = "Complete"
	repoJob.Update(job)

	jobFounded, err := repoJob.Find(job.ID)
	require.NotEmpty(t,jobFounded.ID)
	require.Nil(t,err)
	require.Equal(t,jobFounded.Status,"Complete")
	require.Equal(t,jobFounded.VideoID,job.VideoID)
}