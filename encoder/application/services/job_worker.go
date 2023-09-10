package services

import (
	"encoder/domain"
	"encoder/framework/utils"
	"encoding/json"
	"os"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

//pegar msg do rabbitmq e iniciar o jobservice para dar o start do processo

type JobWorkerResult struct {
	Job domain.Job
	Message *amqp.Delivery
	Error error
}

//variavel para evitar race condition entre os workers
var Mutex = &sync.Mutex{}

func JobWorker(messageChannel chan amqp.Delivery, returnChan chan JobWorkerResult, jobService JobService, job domain.Job, workerID int) {

	//padrao do json recebido na queue
	//{
		//resource_id: id do servico que enviou a msg, pasta na bucket
		//file_path: caminho do video na bucket
	//}
	
	//ler o messageChannel, que é recebida no canal que contem o video que se quer converter
	//messageChannel contem resourceID e video, e vem do rabbitmq
	for message := range messageChannel {
		//pegar msg do body do json
		//validar se o json é valido
		err := utils.IsJson(string(message.Body))
		if err != nil {
			//envia o error para o canal
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		Mutex.Lock() // travar a operacao jobService.VideoService.Video.ID para evitar race condition

		//quando receber a message.body que contem resource_id e file_path,
		//preenche o jobservice os valores em video de resource_id e file_path
		err = json.Unmarshal(message.Body, &jobService.VideoService.Video)
		//criar id do video
		jobService.VideoService.Video.ID = uuid.NewV4().String()
		Mutex.Unlock()//apos executado destravar para os outros workers

		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		
		//validar video
		err = jobService.VideoService.Video.Validate()
		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		
		Mutex.Lock()
		//inserir video no db
		err = jobService.VideoService.InsertVideo()
		Mutex.Unlock()
		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		//preparar o job para rodar
		job.Video = jobService.VideoService.Video
		job.OutputBucketPath = os.Getenv("OUTPUT_BUCKET_NAME")
		job.ID = uuid.NewV4().String()
		job.Status = "STARTING"
		job.CreatedAt = time.Now()

		Mutex.Lock()
		//inserir o job no db
		_, err = jobService.JobRepository.Insert(&job)
		Mutex.Unlock()
		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		
		//dar o start
		jobService.Job = &job
		err = jobService.Start()
		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		returnChan <- returnJobResult(job,message, err)
	}
}

//retorna o JobResult com error
func returnJobResult(job domain.Job, message amqp.Delivery, err error) JobWorkerResult {
	result := JobWorkerResult{
		Job: job,
		Message: &message,
		Error: err,
	}
	return result
}