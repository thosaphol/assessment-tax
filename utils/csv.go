package utils

import (
	"encoding/csv"
	"errors"
	"io"
)

type CsvReader struct {
	fReader io.Reader
	cReader *csv.Reader
	record  []string
	err     error
}

func NewCsvReader(r io.Reader) CsvReader {
	reader := csv.NewReader(r)
	return CsvReader{fReader: r, cReader: reader, record: nil}
}

func (csvReader *CsvReader) ReadLine() bool {
	record, err := csvReader.cReader.Read()
	csvReader.record = record
	csvReader.err = err

	if err != nil {
		return false
	}

	if len(record) == 0 {
		err = errors.New("End of CSV File")
	}

	csvReader.record = record
	csvReader.err = err
	return err == nil
}

func (csvReader *CsvReader) GetLine() ([]string, error) {
	return csvReader.record, csvReader.err
}
