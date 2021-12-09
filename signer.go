package main

import (
	"fmt"
	"sort"
	"strings"
)

func ExecutePipeline(job ...job) {
	var in, out chan interface{}
	for _, fnc := range job {
		fnc(in, out)
	}
}

func SingleHash(in, out chan interface{}) {
	rawData := <-in

	switch typeData := rawData.(type) {
	case string:
		out <- DataSignerCrc32(typeData) + "~" + DataSignerCrc32(DataSignerMd5(typeData))
	default:
		out <- fmt.Errorf("Unexpected data type. Expected string, received %T", typeData)
	}
	return
}

func MultiHash(in, out chan interface{}) {
	rawData := <-in
	hashResult := strings.Builder{}
	hashResult.Grow(6)

	switch typeData := rawData.(type) {
	case string:
		for i := 0; i <= 5; i++ {
			hashResult.WriteString(DataSignerCrc32(fmt.Sprintf("%d%s", i, typeData)))
		}
		out <- hashResult.String()
	default:
		out <- fmt.Errorf("Unexpected data type. Expected string, received %T", typeData)
	}
	return
}

func CombineResults(in, out chan interface{}) {
	strArray := make([]string, 0, 2)

loop:
	select {
	case data := <-in:
		str, _ := data.(string)
		strArray = append(strArray, str)
	default:
		sort.Strings(strArray)
		out <- strings.Join(strArray, "_")
		break loop
	}

	// sort.Strings(hashes)
	// return strings.Join(hashes, "_")
}
