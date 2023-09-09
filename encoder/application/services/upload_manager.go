package services

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"cloud.google.com/go/storage"
)

type VideoUpload struct {
	Paths []string //varios caminhos pois sao varios videos
	VideoPath string
	OutputBucket string
	Errors []string
}

func NewVideoUpload() *VideoUpload {
	return &VideoUpload{}
}

//enviar o arquivo armazenado fragmentado para a bucket, sera chamada muitas vezes 
func(vu *VideoUpload) UploadObject(objectPath string, client *storage.Client, ctx context.Context) error {
	//pegar o caminho completo  e separa para pegar somente o nome do arquivo de video
	path := strings.Split(objectPath, os.Getenv("localstoragePath") + "/")
	
	//abrir o arquivo
	f, err := os.Open(objectPath)
	if err != nil {
		return err
	}
	defer f.Close()

	//client para inserir o arquivo na gcp, inseri o arquivo de video
	wc := client.Bucket(vu.OutputBucket).Object(path[1]).NewWriter(ctx)
	//permissao para todos os usuarios que tem permissao para leitura
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}

	if _, err = io.Copy(wc,f); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

//carregar todos caminhos dos videos fragmentados
func (vu *VideoUpload) loadPaths() error {
	//percorrer a pasta dos arquivos fragmentados, roda a funcao para cada arquivo
	err := filepath.Walk(vu.VideoPath, func(path string, info os.FileInfo, err error) error {
		//se nao for diretorio e ser um arquivo armazena o caminho
		if !info.IsDir() {
			vu.Paths = append(vu.Paths, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

//gerenciador de upload de video
 func (vu *VideoUpload) ProcessUpload(concurrency int, doneUpload chan string) error {
	//ponto de entrada para processamento, posicao da path que contem o arquivo
	in := make(chan int, runtime.NumCPU())
	returnChannel := make(chan string)// canal de retorno avisa o final das operacoes

	//carregar os caminhos com os arquivos frag para upload
	err := vu.loadPaths()
	if err != nil {
		return err
	}

	//conexao com ctx e cliente para realizar o upload dos arquivos
	uploadClient,ctx,err := getClientUpload()
	if err != nil {
		return err
	}

	//loop para determinar a quantidade de processos simultaneos fazendo upload exemplo de 20 em 20
	//quantidade goroutines depende de concurrency
	//ira realizar a leitura de in que é o canal que contem a posicao das paths a ser processada o upload
	for proccess := 0; proccess < concurrency; proccess++ {
		go vu.uploadWorker(in,returnChannel,uploadClient,ctx)
	}

	//funcao responsavel por ler os caminhos de cada arquivo e repassar para o canal in
	//os arquivos serao repassados para o canal in que sera utilizado por uploadWorker para realizar o processo de upload 
	go func() {
		//percorre todos caminhos com os arquivos frag para upload
		//envia para in o numero das linhas de paths dos arquivos para upload
		for x :=0; x < len(vu.Paths); x++ {
			in <- x
		}
		close(in) 
	}()

	//leitura do returnChan
	for r := range returnChannel {
		if r != "" {
			//indica que ocorreu erro ou o upload foi completado
			//se o ocorrer error termina o processo instantaneamente
			doneUpload <- r
			break
		}
	}

	return nil
 }

 // inicia os processos de upload, realizar os uploads
 //faz leitura do canal in
 func (vu *VideoUpload) uploadWorker(in chan int, returnChan chan string, uploadClient *storage.Client, ctx context.Context) {
	//leitura dos item recebidos em in
	//x é a posicao do path do arquivo
	for x := range in {
		//faz o upload do arquivo para a cloud
		err := vu.UploadObject(vu.Paths[x],uploadClient,ctx)
		if err != nil {
			//envia o arquivo que deu error para armazenar em Errors
			vu.Errors = append(vu.Errors, vu.Paths[x])
			log.Printf("error in upload: %v. Error: %v", vu.Paths[x], err)
			returnChan <- err.Error() //recebe o erro
		}

		returnChan <- "" //indica para returnChan que nao obteve erros
	}

	returnChan <- "uploaded completed"
 }

//pegar o client para fazer o upload para a bucket
func getClientUpload() (*storage.Client, context.Context, error) {
	ctx := context.Background()
	client,err := storage.NewClient(ctx)
	if err != nil {
		return nil,nil, err
	}
	return client, ctx, nil
}