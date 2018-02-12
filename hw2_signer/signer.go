package main

import (
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})
	out := make(chan interface{})

	wg := &sync.WaitGroup{}
	for _, job := range jobs {
		wg.Add(1)
		go worker(wg, job, in, out)
		in = out
		out = make(chan interface{})
	}

	wg.Wait()
}

func worker(waiter *sync.WaitGroup, j job, in, out chan interface{}) {
	defer waiter.Done()
	defer close(out)
	j(in, out)
	runtime.Gosched()
}

func SingleHash(in, out chan interface{}) {
	for input := range in {
		value, ok := input.(int)
		if !ok {
			panic("sh: failed type assertion")
		}
		data := strconv.Itoa(value)
		result := DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
		out <- result
	}
}

func MultiHash(in, out chan interface{}) {
	for input := range in {
		data, ok := input.(string)
		if !ok {
			panic("mh: failed type assertion")
		}
		var result string
		for th := 0; th < 6; th++ {
			result += DataSignerCrc32(strconv.Itoa(th) + data)
		}
		out <- result
	}
}

func CombineResults(in, out chan interface{}) {
	var hashes []string
	for input := range in {
		data, ok := input.(string)
		if !ok {
			panic("cr: failed type assertion")
		}

		hashes = append(hashes, data)
	}
	sort.Strings(hashes)

	result := strings.Join(hashes, "_")
	out <- result
}
