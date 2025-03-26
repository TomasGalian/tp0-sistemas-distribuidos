package common

import (
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
	bet           Bet
}


func NewLottery(serverAddress string, bet Bet, agencyID string) *Lottery {
	lottery := &Lottery{
		serverAddress: serverAddress,
		bet:           bet,
		agencyID:      agencyID,
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
	}
	l.conn = conn
}

func (l *Lottery) SendBet() {
	// Create signal channel and assign them to handleSigterm
	sigChan := make(chan os.Signal, 1)
	l.handleSigterm(sigChan)

	// Create connection to server
	l.createConnection()

	// Serialize Bet
	betSerialized := serializeBet(l.agencyID, &l.bet)
	lengthBet := len(betSerialized)

	// Send bet
	for lengthBet > 0 {
		n, err := l.conn.Write(betSerialized)
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: %v",
				l.bet.documento,
				l.bet.numero,
				err,
			)
			return
		}
		lengthBet -= n
		betSerialized = betSerialized[n:]
	}

	// Recibir un byte como ACK
	ack := make([]byte, 1) // Un solo byte
	_, err := l.conn.Read(ack)
	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: %v",
				l.bet.documento,
				l.bet.numero,
				err,
			)
			return
	}

	// Verificar si recibimos el ACK
	if ack[0] == 0x01 {
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			l.bet.documento,
			l.bet.numero,
		)
	} else {
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: %v",
				l.bet.documento,
				l.bet.numero,
				err,
			)
	}

	// Close connection
	l.conn.Close()
}
