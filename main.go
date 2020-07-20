package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func main() {

	var wg sync.WaitGroup

	urls := [...]string{
		"https://golang.org/src/net/http/response.go",
		"https://golang.org/src/net/http/cookie.go",
		"https://bravelancer.com",
	}

	for _, url := range urls {
		wg.Add(1)
		go func(url string, wg *sync.WaitGroup) {

			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("URL Down!")
				sendCloudWatchMetric(url, 0)
				wg.Done()
				return
			}

			if resp.StatusCode == http.StatusOK {
				fmt.Println("URL Ok!")
				sendCloudWatchMetric(url, 1)
			} else {
				fmt.Println("URL Down!")
				sendCloudWatchMetric(url, 0)
			}
			wg.Done()
			return

		}(url, &wg)
	}

	wg.Wait()
}

func sendCloudWatchMetric(url string, status float64) bool {

	input := cloudwatch.PutMetricDataInput{
		Namespace: aws.String("Website/Status"),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String("Status"),
				Unit:       aws.String("Count"),
				Value:      aws.Float64(status),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("WebsiteURL"),
						Value: aws.String(url),
					},
				},
			},
		},
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		fmt.Println(err)
	}

	svc := cloudwatch.New(sess)

	_, err = svc.PutMetricData(&input)
	if err != nil {
		fmt.Println("Error adding metrics:", err.Error())
		return false
	}

	return true
}
