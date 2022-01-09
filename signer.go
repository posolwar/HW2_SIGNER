package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	maxTh int = 6
)

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	for _, soloJob := range jobs {
		wg.Add(1)
		out := make(chan interface{})
		go doJob(in, out, wg, soloJob)
		in = out
	}
}

func doJob(in, out chan interface{}, wg *sync.WaitGroup, soloJob job) {
	defer wg.Done()
	defer close(out)
	soloJob(in, out)
}

func goroutCrc32(data string, outStrChan chan string) {
	outStrChan <- DataSignerCrc32(data)
}

func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	for v := range in {
		wg.Add(1)
		go singleHashJob(v.(int), wg, mu, out)
	}
}

func singleHashJob(value int, wg *sync.WaitGroup, mu *sync.Mutex, out chan interface{}) {
	defer wg.Done()

	crc32one := make(chan string)
	crc32two := make(chan string)
	data := strconv.Itoa(value)

	go goroutCrc32(data, crc32one)

	mu.Lock()
	md5 := DataSignerMd5(data)
	mu.Unlock()

	go goroutCrc32(md5, crc32two)

	crc32 := <-crc32one
	md5crc32 := <-crc32two

	out <- crc32 + "~" + md5crc32
}

func MultiHash(in, out chan interface{}) {
	//	start := time.Now()
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	for v := range in {
		wg.Add(1)
		go multiHashJob(v.(string), wg, out)
	}

	//	fmt.Printf("time from start MultiHash: %v \n", time.Since(start))
}

func multiHashJob(value string, wg *sync.WaitGroup, out chan interface{}) {
	defer wg.Done()

	muInJob := &sync.Mutex{}
	wgInJob := &sync.WaitGroup{}
	hashResult := make([]string, maxTh)

	wgInJob.Add(maxTh)

	for i := 0; i < maxTh; i++ {
		data := strconv.Itoa(i) + value
		go func(data string, numerator int, results []string, wgInJob *sync.WaitGroup, muInJob *sync.Mutex) {
			results[numerator] = DataSignerCrc32(data)
			wgInJob.Done()
		}(data, i, hashResult, wgInJob, muInJob)
	}
	wgInJob.Wait()
	out <- strings.Join(hashResult[:], "")
}

func CombineResults(in, out chan interface{}) {
	//	start := time.Now()

	resSlice := make([]string, 0, 6)

	for v := range in {
		resSlice = append(resSlice, v.(string))
	}

	sort.Strings(resSlice)
	out <- strings.Join(resSlice[:], "_")

	//	fmt.Printf("time from start CombineResults: %v \n ", time.Since(start))
}
