package main

import (
	"sort"
	"strconv"
	"sync"
)

func (t job) functionCall( in, out chan interface{}, wg *sync.WaitGroup){
	t(in,out)
	close(out)
	wg.Done()
}

func crc32HashValue(ch chan string, value string){
	ch <- DataSignerCrc32(value)
}

func concurrentSingleHashFunction(output chan interface{}, value string, hashedValue string, wg *sync.WaitGroup){
	hashedChannel := make(chan string)

	go crc32HashValue(hashedChannel, hashedValue)
	result := DataSignerCrc32(value)
	result += "~" + <- hashedChannel

	output <- result

	close(hashedChannel)
	wg.Done()
}

func SingleHash(in, out chan interface{}){
	wg := sync.WaitGroup{}

	for val, ok := (<- in).(int); ok;{
		value := strconv.Itoa(val)
		md5 := DataSignerMd5(value)
		wg.Add(1)
		go concurrentSingleHashFunction(out, value, md5, &wg)
		val, ok = (<- in).(int)
	}

	wg.Wait()
}

func concurrentMultiHashFunction(output chan interface{}, value string, wg *sync.WaitGroup){
	resultString := ""
	resultChannel := make([]chan string, 6)
	for i := 0; i < 6; i ++{
		resultChannel[i] = make(chan string)
		go crc32HashValue(resultChannel[i], strconv.Itoa(i) + value)
	}
	for i := 0; i <= 5; i++{
		val := <-resultChannel[i]
		resultString += val

		close(resultChannel[i])
	}

	output <- resultString
	wg.Done()
}

func MultiHash(in, out chan interface{}){
	wg := sync.WaitGroup{}
	for value, ok := (<-in).(string); ok;{
		wg.Add(1)
		go concurrentMultiHashFunction(out, value, &wg)
		value, ok = (<-in).(string)
	}

	wg.Wait()
}

func CombineResults(in, out chan interface{}){
	values := make([]string, 0)
	for val, ok := (<-in).(string); ok;{
		values = append(values, val)
		val,ok = (<-in).(string)
	}

	resultString := ""
	sort.Strings(values)
	for i,val := range values{
		if i == 0{
			resultString += val
			continue
		}
		resultString += "_" + val
	}

	out <- resultString
}

func ExecutePipeline(jobs ...job){
	length := len(jobs)

	channels := make([]chan interface{},length + 1)
	for i := 0; i < length + 1; i++{
		channels[i] = make(chan interface{})
	}
	wg := &sync.WaitGroup{}

	for index, val := range jobs{
		wg.Add(1)
		go val.functionCall(channels[index], channels[index+1], wg)
	}
	close(channels[0])
	wg.Wait()
}

func main(){

}