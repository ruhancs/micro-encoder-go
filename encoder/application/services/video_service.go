package services

import (
	"context"
	"encoder/application/repositories"
	"encoder/domain"
	"io"
	"os/exec"

	//"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/storage"
)

type VideoService struct {
	Video *domain.Video
	VideoRepository *repositories.VideoRepository //interface do repositorio
}

//nao usa ponteiro para pois sera acessado de qualquer lugar
func NewVideoService() VideoService {
	return VideoService{}
}

func (v *VideoService) Download(bucketName string) error {
	//ctx para desviar o contexto no meio do caminho e realizar outra tarefa
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	
	bkt := client.Bucket(bucketName)
	obj := bkt.Object(v.Video.FilePath)//obj que se quer fazer o download caminho do video
	r, err := obj.NewReader(ctx) //leitor do arquivo
	if err != nil {
		return err
	}
	defer r.Close()

	body,err := io.ReadAll(r)//ler o arquivo de video
	if err != nil {
		return err
	}

	//copiar o conteudo lido em r para o novo arquivo
	//busca na variavel de ambiente localstoragePath o local onde o video sera salvo
	f, err := os.Create(os.Getenv("localstoragePath") + "/" + v.Video.ID + ".mp4")//criar arquivo no caminho nome do arqivo é o id do video
	if err != nil {
		return err
	}

	//verificar se tudo deu certo
	_,err = f.Write(body)//criar arquivo igual ao lido
	if err != nil {
		return err
	}

	defer f.Close()

	log.Printf("video %v has been storage", v.Video.ID)

	return nil
}

//executar comando no sytema op para executar o bento4 e fragmentar o video
func (v *VideoService) Fragment() error {
	//criar pasta para enviar o video fragmentado
	err := os.Mkdir(os.Getenv("localstoragePath") + "/" + v.Video.ID, os.ModePerm)
	if err != nil {
		return err
	}

	//de onde vem o video e para onde sera enviado
	source := os.Getenv("localstoragePath") + "/" + v.Video.ID + ".mp4"
	target := os.Getenv("localstoragePath") + "/" + v.Video.ID + ".frag"

	//executar o bento4 para fragmentar o video
	cmd := exec.Command("mp4fragment",source,target)
	output,err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(output)
	return nil
}

func printOutput(output []byte) {
	if len(output) > 0 {
		log.Printf("=====> Output: %s\n", string(output))
	}
}