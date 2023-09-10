package main

import (
	"encoder/application/services"
	"encoder/framework/database"
	"encoder/framework/queue"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

//criar no rabbitmq exchange dlx para receber msg rejeitadas, exchange deve ter o type fanout
//criar fila com binding em dlx para pegar os resultados das falhas de processamento

//criar fila para enviar os resultados do processamento de videos
//fazer binding da fila criada com amq.direct e inserir a Routing key = jobs

//fila de videos é criada automaticamente, fila para consumir os dados do video para processar

//detectar race condition: go run -race framework/cmd/server/server.go

var db database.Database

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	autoMigrateDB,err := strconv.ParseBool(os.Getenv("AUTO_MIGRATE_DB"))
	if err != nil {
		log.Fatalf("Error parsing bool env file")
	}
	debug,err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		log.Fatalf("Error parsing bool env file")
	}

	db.AutoMigrateDb = autoMigrateDB
	db.Debug = debug
	db.DsnTest = os.Getenv("DSN_TEST")
	db.Dsn = os.Getenv("DSN")
	db.DbtypeTest = os.Getenv("DB_TYPE_TEST")
	db.DbType = os.Getenv("DB_TYPE")
	db.Env = os.Getenv("ENV")
}

func main() {
	//canal do job manager
	messageChannel := make(chan amqp.Delivery)
	//canal para receber resultado dos jobWorks
	jobReturnChannel := make(chan services.JobWorkerResult)
	
	//conexao com db
	dbConnection,err := db.Connect()
	if err != nil {
		log.Fatalf("Error connecting to db")
	}
	defer dbConnection.Close()

	rabbitMQ := queue.NewRabbitMQ()
	ch := rabbitMQ.Connect()
	defer ch.Close()

	//rabbitmq consumir as msg de entrada, messageChannel envia para jobManager
	rabbitMQ.Consume(messageChannel)

	//conecta com db
	//pega msg recebidas de outro servico no rabbitmq
	//retorna o resultado do processamento das msg
	jobManager := services.NewJobManager(dbConnection,rabbitMQ,jobReturnChannel,messageChannel)
	jobManager.Start(ch)//conexao com rabbitmq
}