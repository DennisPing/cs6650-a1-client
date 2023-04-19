package log

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func LogThroughput(throughputSamples []float64) error {
	fmt.Println(len(throughputSamples))
	csvFile, err := os.Create("avg_throughput.csv")
	if err != nil {
		return err
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// Write CSV header
	err = csvWriter.Write([]string{"time", "throughput"})
	if err != nil {
		return err
	}

	// Write the throughput samples to the CSV file
	for i, throughput := range throughputSamples {
		line := make([]string, 2) // 2 columns
		line[0] = strconv.Itoa(i + 1)
		line[1] = strconv.FormatFloat(throughput, 'f', 2, 64)
		err = csvWriter.Write(line)
		if err != nil {
			return err
		}
	}
	return nil
}
