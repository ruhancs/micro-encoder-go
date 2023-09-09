package services_test

import (
	"encoder/application/services"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func TestVideoServiceUpload(t *testing.T) {
	video,repo := prepare()
	videoService := services.NewVideoService()
	videoService.Video = video
	videoService.VideoRepository = repo

	err := videoService.Download("videosgo")
	require.Nil(t,err)
	
	err = videoService.Fragment()
	fmt.Println("ERROR")
	fmt.Println(err)
	require.Nil(t,err)

	err = videoService.Encode()
	require.Nil(t,err)
	
	videoUpload := services.NewVideoUpload()
	videoUpload.OutputBucket = "videosgo"
	videoUpload.VideoPath = os.Getenv("localstoragePath") + "/" + video.ID

	doneUpload := make(chan string)
	go videoUpload.ProcessUpload(50,doneUpload)

	result := doneUpload
	require.Equal(t,result, "uploaded completed")

	err = videoService.Finish()
	require.Nil(t,err)
}