package services

import (
	"encoder/application/repositories"
	"encoder/domain"
	"encoder/framework/queue"
	"encoding/json"

	//"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

//criar job_worker para ler rabbitmq e processar msgs

type JobManager struct {
	DB *gorm.DB
	Domain domain.Job
	MessageChan chan amqp.Delivery
	JobReturnChannel chan JobWorkerResult
	RabbitMQ *queue.RabbitMQ
}

type JobNotificationError struct {
	Message string `json:"message"`
	Error string `json:"error"`
}

func NewJobManager(db *gorm.DB, rabbitMQ *queue.RabbitMQ, jobReturnChannel chan JobWorkerResult, messageChannel chan amqp.Delivery) *JobManager {
	return &JobManager{
		DB: db,
		RabbitMQ: rabbitMQ,
		JobReturnChannel: jobReturnChannel,
		MessageChan: messageChannel,
		Domain: domain.Job{},
	}
}

//ch canal do rabbitmq para ser lido
func(j *JobManager) Start(ch *amqp.Channel) {
	videoService := NewVideoService()
	videoService.VideoRepository = repositories.VideoRepositoryDb{Db: j.DB}

	jobService := JobService{
		JobRepository: repositories.JobRepositoryDb{Db: j.DB},
		VideoService: videoService,
	}

	concurrency,err := strconv.Atoi(os.Getenv("CONCURRENCY_WORKERS"))
	if err != nil {
		log.Fatalf("error loading var: CONCURRENCY_WORKERS")
	}
	for qtdProccess := 0; qtdProccess < concurrency; qtdProccess++ {
		//qtdProccess gera o id do worker cada worker tera um id
		//numero de videos que serao tratados simutaneament = CONCURRENCY_WORKERS
		go JobWorker(j.MessageChan,j.JobReturnChannel, jobService,j.Domain, qtdProccess)
	}

	//tratar o retorno do canal JobReturnChannel
	for jobResult := range j.JobReturnChannel {
		if jobResult.Error != nil {
			//notificacao de erro na fila
			err = j.checkParserErrors(jobResult)
		} else {
			//notificacao de sucesso na fila
			err = j.notifySuccess(jobResult, ch)
		}

		if err != nil {
			//para nao recolacar a menssagem na fila
			jobResult.Message.Reject(false)
		}
	}
}

func (j *JobManager) notifySuccess(jobResult JobWorkerResult, ch *amqp.Channel) error {
	jobJson, err := json.Marshal(jobResult.Job)//resultado do jobResult em json
	if err != nil {
		return err
	}

	//notificacao do resutado na exchange
	err = j.notify(jobJson)
	if err != nil {
		return err
	}

	//apagar a msg que ja foi processada
	err =jobResult.Message.Ack(false)
	if err != nil {
		return err
	}

	return nil
}

func (j *JobManager) checkParserErrors(jobResult JobWorkerResult) error {
	if jobResult.Job.ID != "" {
		//id da msg do rabbitmq jobResult.Message.DeliveryTag
		log.Printf("MessageID #{jobResult.Message.DeliveryTag}. Error parsing job: #{jobResult.Job.ID}")
		} else {
			log.Printf("MessageID #{jobResult.Message.DeliveryTag}. Error parsing message job: #{jobResult.Error}")
	}

	errMSG := JobNotificationError{
		Message: string(jobResult.Message.Body),
		Error: jobResult.Error.Error(),
	}

	//transforma em json
	jobJson,err := json.Marshal(errMSG)
	if err != nil {
		return err
	}

	//notificacao de error para a fila
	err = j.notify(jobJson)
	if err != nil {
		return err
	}

	//rejeitar a msg e nao processar novamente, remove da fila
	err = jobResult.Message.Reject(false)
	if err != nil {
		return err
	}

	return nil
}

func (j *JobManager) notify(jobJson []byte) error {
	err := j.RabbitMQ.Notify(
		string(jobJson),
		"application/json", //content-type
		os.Getenv("RABBITMQ_NOTIFICATION_EX"),//exchange
		os.Getenv("RABBITMQ_NOTIFICATION_ROUTING_KEY"),
	)

	if err != nil {
		return err
	}

	return nil
}