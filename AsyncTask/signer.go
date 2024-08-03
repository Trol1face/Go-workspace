package main

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
)

func CalculateCrc32(crc32 *string, input string, wg *sync.WaitGroup) {
	*crc32 = DataSignerCrc32(input)
	wg.Done()
}

func SingleHash(in, out chan interface{}) {
	wg1 := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for input := range in {
		wg1.Add(1)
		go func(input interface{}) {
			wg2 := &sync.WaitGroup{}
			crc32, md5, crc32Md5 := "", "", ""

			dataInt, ok := input.(int)
			if !ok {
				fmt.Errorf("can't convert result data to string")
			}
			data := strconv.Itoa(dataInt)

			wg2.Add(1)
			go CalculateCrc32(&crc32, data, wg2)

			mu.Lock()
			md5 = DataSignerMd5(data)
			mu.Unlock()

			wg2.Add(1)
			go CalculateCrc32(&crc32Md5, md5, wg2)

			wg2.Wait()

			result := crc32 + "~" + crc32Md5
			out <- result

			wg1.Done()
		}(input)

	}
	wg1.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg1 := &sync.WaitGroup{}
	for input := range in {
		wg1.Add(1)
		go func() {
			wg2 := &sync.WaitGroup{}
			calculations := make([]string, 6)
			result := ""

			data, ok := input.(string)
			if !ok {
				fmt.Errorf("can't convert result data to string")
			}

			for i := 0; i <= 5; i++ {
				wg2.Add(1)
				go func(i int) {
					calculations[i] = DataSignerCrc32(strconv.Itoa(i) + data)
					wg2.Done()
				}(i)
			}

			wg2.Wait()

			for _, c := range calculations {
				result += c
			}

			out <- result
			wg1.Done()
		}()

	}
	wg1.Wait()
}

func CombineResults(in, out chan interface{}) {
	var result string
	parts := make([]string, 0)

	for input := range in {
		data, ok := input.(string)
		if !ok {
			fmt.Errorf("can't convert result data to string")
		}
		parts = append(parts, data)
	}

	slices.Sort(parts)

	lastPartNum := len(parts) - 1
	for i, part := range parts {
		result += part
		if i != lastPartNum {
			result += "_"
		}
	}

	out <- result
}

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	jobsCount := len(jobs)
	channels := make([]chan interface{}, jobsCount+1)

	for i, jb := range jobs {
		channels[i+1] = make(chan interface{})

		wg.Add(1)
		go ExecuteJob(i, jb, channels[i], channels[i+1], wg)
	}
	wg.Wait()
}

func ExecuteJob(i int, job job, in, out chan interface{}, wg *sync.WaitGroup) {
	job(in, out)
	if out != nil {
		close(out)
	}
	wg.Done()
}
