package wcpeaks

import (
	"encoding/csv"
	"io"
	"os"

	"strconv"
)

func Read(inputFile string) ([]*Peak, error) {
	f, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = 3
	r.Comma = ';'
	peaks := make([]*Peak, 0, 1000)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		p := NewPeak(record)
		peaks = append(peaks, p)
	}
	return peaks, nil
}

func NewPeak(record []string) *Peak {
	ele, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		panic(err)
	}
	return &Peak{
		Name:  record[0],
		Range: record[1],
		Ele:   ele,
	}
}

type Peak struct {
	Name  string
	Range string
	Ele   float64
}
