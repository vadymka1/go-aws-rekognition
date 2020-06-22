package services

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/joho/godotenv"
	"log"
	"os"
	"upload_images/models"
)

type SQSMessage struct {
	Body string `json:"Body"`
}

func init () {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func SendMessage (message string, s *session.Session, image *models.Image) {
	fmt.Print("start sqs Service \n")
	svc := sqs.New(s)

	SendParams := &sqs.SendMessageInput{
		DelaySeconds:            aws.Int64(3),
		MessageBody:             aws.String(message),
		QueueUrl:                aws.String(os.Getenv("AWS_SQS_URL")),
	}

	sendMessage, err := svc.SendMessage(SendParams)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("[Send message] \n%v \n\n", sendMessage)

	RecieveParam := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(os.Getenv("AWS_SQS_URL")),
		MaxNumberOfMessages: aws.Int64(3),
		VisibilityTimeout:   aws.Int64(30),
		WaitTimeSeconds:     aws.Int64(20),
	}
	RecieveMessage, err := svc.ReceiveMessage(RecieveParam)
	if err != nil {
		log.Println(err)
	}

	s3url := RecieveMessage.Messages[0].Body
	fmt.Println("Start Unmarshalling ")
	fmt.Println("url : ", *s3url)
	fmt.Println("End of Unmarshalling")

	GetImageData(*s3url, s, image)

	deleteMessages(RecieveMessage.Messages, svc)

}

func deleteMessages (messages []*sqs.Message, svc *sqs.SQS) {
	for _, message := range messages {
		deleteParams := &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(os.Getenv("AWS_SQS_URL")),
			ReceiptHandle: message.ReceiptHandle,
		}
		_, err := svc.DeleteMessage(deleteParams) // No response returned when successed.
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("[Delete message] \nMessage ID: %s has beed deleted.\n\n", *message.MessageId)
	}
}