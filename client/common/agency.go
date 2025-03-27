package common

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")
 
type Agency struct {
	agencyID      string
	serverAddress string
	conn          net.Conn
	reader     	 *Reader
}

func NewAgency(serverAddress string, agencyID string, maxAmount int) *Agency {
	reader, err := NewReader("./agency.csv", maxAmount)
	
	if err != nil {
		log.Criticalf("action: open_file | result: fail | agency_id: %v | error: %v", agencyID, err)
		return nil
	}

	agency := &Agency{
		serverAddress: serverAddress,
		agencyID:      agencyID,
		reader:        reader,
	}
	return agency
}

func (a *Agency) handleSigterm(sigChan chan os.Signal) {
	// Handle SIGTERM
	signal.Notify(sigChan, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Debugf("Received SIGTERM, shutting down gracefully...")
		// If there is a connection, close it
		if a.conn != nil {
			err := a.conn.Close()
			if err == nil {
				log.Infof("action: close_connection | result: success | agency_id: %v", a.agencyID)
			}
		}
		os.Exit(0)
	}()
}

func (a *Agency) createAgencySocket() {
	// Connect to server
	conn, err := net.Dial("tcp", a.serverAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | agency_id: %v | error: %v",
			a.agencyID,
			err,
		)
	}
	a.conn = conn
}

func (a *Agency) sendData(data []byte) error {
	lengthData := len(data)
	for lengthData > 0 {
		n, err := a.conn.Write(data)
		if err != nil {
			log.Errorf("action: send_data | result: fail | agency_id: %v | error: %v",
				a.agencyID,
				err,
			)
			return err
		}
		lengthData -= n
		data = data[n:]
	}
	return nil
}

func (a *Agency) waitACK() error {
	ack := make([]byte, 1)
	if _, err := a.conn.Read(ack); err != nil {
		return err
	}
	log.Debugf("action: wait_ack | result: success | ack: %v", ack)
	if ack[0] != 0xFF {
		err := errors.New("invalid ACK received")
		return err
	}

	return nil
}

func (a *Agency) SendBets(bets []Bet) error {
	// Serialize Bets
	betsSerialized := make([]byte, 0)
	for _, bet := range bets {
		betSerialized := bet.serializeBet(a.agencyID)
		betsSerialized = append(betsSerialized, betSerialized...)
	}
	// Use the byte of action to send the number of bets
	// If the number of bets is 1 the server understands like ej5
	// The max number of bets is 254 because of the ACK
	betsSerialized = append([]byte{uint8(len(bets))}, betsSerialized...)

	// Send bet
	err := a.sendData(betsSerialized)
	if err != nil {
		return err
	}

	// Wait for ACK
	err_ack := a.waitACK()
	if err_ack != nil {
		return err
	}

	log.Infof("action: apuesta_enviada | result: success | cantidad: %v", len(bets))
	return nil
}

func (a *Agency) receiveWinners(clientSock net.Conn, totalBets int) ([]Bet, error) {
	var bets []Bet
	fields := []string{"Agencia", "Nombre", "Apellido", "Documento", "Fecha de nacimiento", "Numero"}

	for i := 0; i < totalBets; i++ {
		bet := make(map[string]string)
		for _, field := range fields {
			lenField := make([]byte, 1)
			_, err := clientSock.Read(lenField)
			if err != nil {
				return nil, err
			}

			fieldLength := int(lenField[0])
			buffer := make([]byte, fieldLength)
			_, err = clientSock.Read(buffer)
			if err != nil {
				return nil, err
			}

			bet[field] = string(buffer)
		}

		bets = append(bets, Bet{
			nombre:      	bet["Nombre"],
			apellido:       bet["Apellido"],
			documento:      bet["Documento"],
			nacimiento:     bet["Fecha de nacimiento"],
			numero:         bet["Numero"],
		})
	}

	return bets, nil
}

func (a *Agency) StartLottery() {
	// Create signal channel and assign them to handleSigterm
	sigChan := make(chan os.Signal, 1)
	a.handleSigterm(sigChan)

	// Create connection to server
	a.createAgencySocket()

	for bets := a.reader.getBets(); len(bets) > 0; bets = a.reader.getBets() {
		// Send bets
		// Handle invalid bet
		if bets == nil {
			log.Errorf("action: create_bets | result: fail | error: invalid bet")
			return
		}
	
		// Send bets
		if err := a.SendBets(bets); err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | agency_id: %v | error: %v", a.agencyID, err)
			a.conn.Close()
			return
		}
	}

	// Send end of bets
	err := a.sendData([]byte{0x00})
	if err != nil {
		log.Errorf("action: send_end_bets | result: fail | agency_id: %v | error: %v", a.agencyID, err)
		return
	}

	// Send winners ask to server: we use 0xF0 to ask for winners
	err = a.sendData([]byte{0xF0})
	if err != nil {
		log.Errorf("action: send_winners | result: fail | agency_id: %v | error: %v", a.agencyID, err)
		return
	}

	winners_len := make([]byte, 1)
	if _, err_winners := a.conn.Read(winners_len); err_winners != nil {
		return
	}

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", winners_len[0])

	// Close reader 
	a.reader.Close()	

	// Close connection
	a.conn.Close()
}
