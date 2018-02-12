package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 100)
	out := make(chan interface{}, 100)

	wg := &sync.WaitGroup{}
	for _, job := range jobs {
		wg.Add(1)
		go worker(wg, job, in, out)
		in = out
		out = make(chan interface{}, 100)
	}

	wg.Wait()
}

func worker(waiter *sync.WaitGroup, j job, in, out chan interface{}) {
	defer waiter.Done()
	defer close(out)
	j(in, out)
}

func SingleHash(in, out chan interface{}) {
	var mu sync.Mutex
	wgSingleHash := &sync.WaitGroup{}

	for input := range in {
		wgSingleHash.Add(1)
		go func(in interface{}) {
			defer wgSingleHash.Done()
			value, ok := in.(int)
			if !ok {
				panic("sh: failed type assertion")
			}
			data := strconv.Itoa(value)

			mu.Lock()
			md5hash := DataSignerMd5(data)
			mu.Unlock()

			mData := map[string]string{
				"data":    data,
				"md5hash": md5hash,
			}
			nmData := make(map[string]string, 2)
			wg := &sync.WaitGroup{}
			for k := range mData {
				wg.Add(1)
				go func(key string) {
					defer wg.Done()
					hash := DataSignerCrc32(mData[key])
					mu.Lock()
					nmData[key] = hash
					mu.Unlock()
				}(k)
			}
			wg.Wait()

			result := nmData["data"] + "~" + nmData["md5hash"]
			out <- result
		}(input)
	}
	wgSingleHash.Wait()
}

func MultiHash(in, out chan interface{}) {
	wgMultiHash := &sync.WaitGroup{}
	for input := range in {
		wgMultiHash.Add(1)
		go func(in interface{}) {
			defer wgMultiHash.Done()
			data, ok := in.(string)
			if !ok {
				panic("mh: failed type assertion")
			}
			wg := &sync.WaitGroup{}
			mu := &sync.Mutex{}

			mData := make(map[int]string, 6)
			for th := 0; th < 6; th++ {
				wg.Add(1)
				go func(mData map[int]string, th int, data string) {
					defer wg.Done()
					hash := DataSignerCrc32(strconv.Itoa(th) + data)
					mu.Lock()
					mData[th] = hash
					mu.Unlock()
				}(mData, th, data)
			}
			wg.Wait()

			keys := make([]int, 0, len(mData))
			for k, _ := range mData {
				keys = append(keys, k)
			}
			sort.Ints(keys)

			var result string
			for k := range keys {
				result += mData[k]
			}

			out <- result
		}(input)
	}
	wgMultiHash.Wait()
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
