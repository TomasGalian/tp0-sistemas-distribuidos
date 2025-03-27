package common

import (
	"encoding/csv"
	"io"
	"os"
)

type Reader struct {
	filepath 	string
	maxAmount 	int
	betsReader 	*csv.Reader
	file 		*os.File // Agregamos una referencia al archivo
}

func NewReader(filepath string, maxAmount int) (*Reader, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = 5 
	
	reader := &Reader	{
		filepath: 	filepath,
		maxAmount: 	maxAmount,
		betsReader: csvReader,
		file: 		file, // Guardamos la referencia al archivo
	}
	return reader, nil
}

func (r *Reader) Close() error {
	if r.file != nil {
		return r.file.Close() // Cerramos el archivo
	}
	return nil
}

func (r *Reader) getBets() []Bet {
	bets := make([]Bet, 0)
	for i := 0; i < r.maxAmount; i++ {
		// Read the bet
		csvBet, err := r.betsReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("action: read_csv | result: fail | error: %v", err)
			return nil
		}
		// Create a bet
		bet := NewBet(csvBet[0], csvBet[1], csvBet[2], csvBet[3], csvBet[4])
		if bet == nil {
			log.Errorf("action: create_bets | result: fail | error: invalid bets fields")
			return nil
		}
		bets = append(bets, *bet)
	}
	return bets
}