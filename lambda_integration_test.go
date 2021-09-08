package opendevopslambda

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
  "flag"
)

var aws_region string

func init() {
  flag.StringVar(&aws_region, "aws_region", "", "AWS Region")
}

func getStackOutputValue(output *cloudformation.DescribeStacksOutput, key string) (string, error) {
	for _, x := range output.Stacks[0].Outputs {
		if *x.OutputKey == key {
			return *x.OutputValue, nil
		}
	}
	return "", errors.New(fmt.Sprintf("unable to find key: %s", key))
}

func getSubmitImageResponse(cf *cloudformation.CloudFormation, imageUrl string) (string, error) {
	submitImageStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String("OpenDevOpsSubmitImage"),
	}

	submitImageStackOutput, err := cf.DescribeStacks(submitImageStackInput)
	if err != nil {
		return "", err
	}

	submitImageEndpoint, err := getStackOutputValue(submitImageStackOutput, "SubmitImageAPI")
	if err != nil {
		return "", err
	}

	getUrl := fmt.Sprintf("%s?url=%s", submitImageEndpoint, imageUrl)

	resp, err := http.Get(getUrl)
	if err != nil {
		return "", err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	respString := string(respBytes)

	return respString, nil
}

func getGetImageLabelResponse(cf * cloudformation.CloudFormation, imageId string) (string, error) {
	getImageLabelStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String("OpenDevOpsGetImageLabel"),
	}

	getImageLabelStackOutput, err := cf.DescribeStacks(getImageLabelStackInput)
	if err != nil {
		return "", err
	}

	getImageLabelEndpoint, err := getStackOutputValue(getImageLabelStackOutput, "GetImageLabelAPI")
	if err != nil {
		return "", err
	}

	getUrl := fmt.Sprintf("%s?imageId=%s", getImageLabelEndpoint, imageId)

	resp, err := http.Get(getUrl)
	if err != nil {
		return "", err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	respString := string(respBytes)

	return respString, nil
}

func TestImageClassificationSystem(t *testing.T) {
	t.Run("Successful Request", func(t *testing.T) {
		awsConfig := aws.Config{
			Region: aws.String(aws_region),
		}

		sess := session.Must(session.NewSession(&awsConfig))
		cf := cloudformation.New(sess)

		imageUrl := "https://i.ytimg.com/vi/iVZYAhzxG4Y/maxresdefault.jpg"

		submitImageResponse, err := getSubmitImageResponse(cf, imageUrl)
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Logf("submitImageResponse: %s", submitImageResponse)

		time.Sleep(time.Second*30)

		imageId := strings.Split(submitImageResponse, ":")[1]
		imageId = strings.Trim(imageId, "\"")
		t.Logf("imageId: %s", imageId)
		getImageLabelResponse, err := getGetImageLabelResponse(cf, imageId)
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Logf("getImageLabelResponse: %s", getImageLabelResponse)

		if !strings.Contains(getImageLabelResponse, "French bulldog") {
			t.Fatalf(fmt.Sprintf("getImageLabelResponse does not container \"%s\"", "French bulldog"))
    }
		if !strings.Contains(getImageLabelResponse, "pug") {
			t.Fatalf(fmt.Sprintf("getImageLabelResponse does not container \"%s\"", "pug"))
		}
		if !strings.Contains(getImageLabelResponse, "bull mastiff") {
			t.Fatalf(fmt.Sprintf("getImageLabelResponse does not container \"%s\"", "bull mastiff"))
		}
	})
}
