package services

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/joho/godotenv"
	"log"
	"os"
	"upload_images/models"
)

func init () {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func GetImageData (s3ImageUrl string, s *session.Session, image *models.Image) {
	fmt.Print("start rekognition Service \n")
	fmt.Println("File name :", s3ImageUrl)

	svc := rekognition.New(s)

	input := &rekognition.DetectLabelsInput{
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket:  aws.String(os.Getenv("S3_BUCKET")),
				Name:    aws.String(s3ImageUrl),
			},
		},
		MaxLabels: aws.Int64(123),
		MinConfidence: aws.Float64(70.0000),
	}


	result, err := svc.DetectLabels(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rekognition.ErrCodeInvalidS3ObjectException:
				fmt.Println(rekognition.ErrCodeInvalidS3ObjectException, aerr.Error())
			case rekognition.ErrCodeInvalidParameterException:
				fmt.Println(rekognition.ErrCodeInvalidParameterException, aerr.Error())
			case rekognition.ErrCodeImageTooLargeException:
				fmt.Println(rekognition.ErrCodeImageTooLargeException, aerr.Error())
			case rekognition.ErrCodeAccessDeniedException:
				fmt.Println(rekognition.ErrCodeAccessDeniedException, aerr.Error())
			case rekognition.ErrCodeInternalServerError:
				fmt.Println(rekognition.ErrCodeInternalServerError, aerr.Error())
			case rekognition.ErrCodeThrottlingException:
				fmt.Println(rekognition.ErrCodeThrottlingException, aerr.Error())
			case rekognition.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(rekognition.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case rekognition.ErrCodeInvalidImageFormatException:
				fmt.Println(rekognition.ErrCodeInvalidImageFormatException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return
	}
	imageLabel := ImageLabels{
		ID:    image.ID,
		Name:  image.Name,
		Label: result.Labels,
	}

	for _, label := range result.Labels {

		imageLabel.Label = append(imageLabel.Label, label)

		fmt.Printf("Name:%s Confidence:%f\n", *label.Name, *label.Confidence)
	}

	SaveLabels(imageLabel)
	fmt.Println("End of rekognition")
}