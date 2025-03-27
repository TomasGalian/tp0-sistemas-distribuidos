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
}

func NewAgency(serverAddress string, agencyID string) *Agency {
	agency := &Agency{
		serverAddress: serverAddress,
		agencyID:      agencyID,
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
	if ack[0] != 0x11 {
		err := errors.New("invalid ACK received")
		return err
	}

	return nil
}

func (a *Agency) SendBet(bet *Bet) error {
	// Serialize Bet
	betSerialized := bet.serializeBet(a.agencyID)
	log.Debugf("action: serialize_bet | result: success | bet: %v", betSerialized)

	// Send bet
	err := a.sendData(betSerialized)
	if err != nil {
		return err
	}
	log.Debugf("action: send_bet | result: success | bet: %v", betSerialized)

	// Wait for ACK
	err_ack := a.waitACK()
	if err_ack != nil {
		return err
	}

	log.Infof("action: apuesta_enviada | result: success | dni: %v, | numero: %v", bet.documento, bet.numero)
	return nil
}

func (a *Agency) StartLottery() {
	// Create signal channel and assign them to handleSigterm
	sigChan := make(chan os.Signal, 1)
	a.handleSigterm(sigChan)

	// Create connection to server
	a.createAgencySocket()

	// Create bet
	bet := NewBet(
		os.Getenv("NOMBRE"),
		os.Getenv("APELLIDO"),
		os.Getenv("DOCUMENTO"),
		os.Getenv("NACIMIENTO"),
		os.Getenv("NUMERO"),
	)

	// Handle invalid bet
	if bet == nil {
		log.Errorf("action: create_bet | result: fail | error: invalid bet fields")
		return
	}

	// Send bet
	if err := a.SendBet(bet); err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | agency_id: %v | error: %v", a.agencyID, err)
		a.conn.Close()
		return
	}

	// Close connection
	a.conn.Close()
}
