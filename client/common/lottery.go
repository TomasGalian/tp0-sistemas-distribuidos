package common

import (
	"encoding/csv"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type Lottery struct {
	serverAddress string
	conn          net.Conn
	agencyID      string
	maxAmount	  int
}


func NewLottery(serverAddress string, agencyID string, maxAmount int) *Lottery {
	lottery := &Lottery{
		serverAddress: serverAddress,
		agencyID:      agencyID,
		maxAmount:     maxAmount,
	}
	return lottery
}

func (l *Lottery) handleSigterm(sigChan chan os.Signal) {
	// Handle SIGTERM
	signal.Notify(sigChan, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Infof("Received SIGTERM, shutting down gracefully...")
		if l.conn != nil {
			err := l.conn.Close()
			if err == nil {
				log.Info("action: close_connection | result: success | agency_id: %v", l.agencyID)
			}
		}
		os.Exit(0)
	}()
}

func (l *Lottery) createConnection() {
	// Connect to server
	conn, err := net.Dial("tcp", l.serverAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | agency_id: %v | error: %v",
			l.agencyID,
			err,
		)
		os.Exit(1)
	}
	l.conn = conn
}

func (l *Lottery) getReader() (*csv.Reader, *os.File) {
	file, err := os.Open("./agency.csv")
	if err != nil {
		log.Criticalf("action: open_file | result: fail | agency_id: %v | error: %v", l.agencyID, err)
		return nil, nil
	}

	betsReader := csv.NewReader(file)
	betsReader.FieldsPerRecord = 5 

	return betsReader, file
}

func (l *Lottery) SendBet() {
	// Create signal channel and assign them to handleSigterm
	sigChan := make(chan os.Signal, 1)
	l.handleSigterm(sigChan)

	log.Debugf("action: connect | result: in_progress | agency_id: %v ", l.agencyID)
	// Create connection to server
	l.createConnection()
	log.Debugf("action: connect | result: success | agency_id: %v ", l.agencyID)

	csvReader, file := l.getReader()
	if csvReader == nil {
		defer file.Close()
		os.Exit(1)
	}
	
	// Read bets from csv
	finish := false
	for !finish {
		// Read bets from csv
		betsSerialized := make([]byte, 0)
		numBets := uint8(0)

		for i := 0; i < l.maxAmount; i++ {
			// Read the bet
			csvBet, err := csvReader.Read()
			if err == io.EOF {
				finish = true
				break
			}
			if err != nil {
				log.Errorf("action: read_csv | result: fail | error: %v", err)
				return
			}

			// Create a bet
			bet := NewBet(csvBet[0], csvBet[1], csvBet[2], csvBet[3], csvBet[4])
			
			// Serialize Bet
			betSerialized := serializeBet(l.agencyID, bet)
			betsSerialized = append(betsSerialized, betSerialized...)
			numBets++
		}

		// No enviar paquetes vacíos
		if numBets == 0 {
			break
		}

		// Send bets to server
		betsSerialized = append([]byte{numBets}, betsSerialized...)
		lengthBets := len(betsSerialized)

		for lengthBets > 0 {
			n, err := l.conn.Write(betsSerialized)
			if err != nil {
				log.Errorf("action: apuesta_enviada | result: fail | error: %v", err)
				return
			}
			if n == 0 {
				log.Errorf("action: apuesta_enviada | result: fail | reason: zero bytes written")
				return
			}
			lengthBets -= n
			betsSerialized = betsSerialized[n:]
		}

		// Recibir un byte como ACK
		ack := make([]byte, 1) // Un solo byte
		n, err := l.conn.Read(ack)
		if err != nil || n != 1 {
			log.Errorf("action: apuesta_enviada | result: fail | error: %v", err)
			return
		}

		// Verificar si recibimos el ACK
		if ack[0] == 0x01 {
			log.Infof("action: apuesta_enviada | result: success")
		} else {
			log.Errorf("action: apuesta_enviada | result: fail | error: unexpected ACK value")
		}
	}

	// Create an empty bet to finish the connection
	bet := NewBet("", "", "", "", "")
		
	// Serialize Bet
	betSerialized := serializeBet(l.agencyID, bet)
	betSerialized = append([]byte{0}, betSerialized...)
	lengthBet := len(betSerialized)

	// Send empty bet to server
	for lengthBet > 0 {
		n, err := l.conn.Write(betSerialized)
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | error: %v", err)
			return
		}
		if n == 0 {
			log.Errorf("action: apuesta_enviada | result: fail | reason: zero bytes written")
			return
		}
		lengthBet -= n
		betSerialized = betSerialized[n:]
	}

	// Create an empty bet to get winners
	winner_msg := NewBet("", "", "", "", "")
	
	// Serialize Bet
	winner_Msg_Serialized := serializeBet(l.agencyID, winner_msg)
	winner_Msg_Serialized = append([]byte{255}, winner_Msg_Serialized...)
	lengthWinnerMsg := len(winner_Msg_Serialized)

	// Send empty bet to server
	for lengthWinnerMsg > 0 {
		n, err := l.conn.Write(winner_Msg_Serialized)
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | error: %v", err)
			return
		}
		if n == 0 {
			log.Errorf("action: apuesta_enviada | result: fail | reason: zero bytes written")
			return
		}
		lengthWinnerMsg -= n
		winner_Msg_Serialized = winner_Msg_Serialized[n:]
	}

	winners := 0

	for {
		// Read the length of the document (1 byte)
		docLength := make([]byte, 1)
		n, err := l.conn.Read(docLength)
		if err != nil || n != 1 {
			log.Errorf("action: recibir_documento_len | result: fail | error: %v", err)
			return
		}

		// Read the document based on the received length
		length := int(docLength[0])

		if length == 0 {
			break
		}

		document := make([]byte, length)
		n, err = l.conn.Read(document)
		if err != nil || n != length {
			log.Errorf("action: recibir_documento | result: fail | error: %v", err)
			return
		}
		winners++
	}

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", winners)

	// Close connection
	l.conn.Close()
}
