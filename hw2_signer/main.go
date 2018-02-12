package main

import (
	"fmt"
	"time"
)

func main() {
	var testResult string
	inputData := []int{0, 1}
	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				panic("cant convert result data to string")
			}
			testResult = data
		}),
	}

	start := time.Now()
	ExecutePipeline(hashSignJobs...)
	end := time.Now().Sub(start)

	fmt.Println(testResult)
	fmt.Println(end)
}
