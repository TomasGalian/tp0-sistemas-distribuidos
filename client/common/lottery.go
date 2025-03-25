package common

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type Bet struct {
	nombre 		string
	apellido 	string
	documento 	string
	nacimiento	string
	numero		string
}

type Lottery struct {
	ServerAddress 	string
	conn 			net.Conn
	agencyId 		string
	bet 			Bet
}

func NewBet(nombre string, apellido string, documento string, nacimiento string, numero string) *Bet {
	bet := &Bet{
		nombre: nombre,
		apellido: apellido,
		documento: documento,
		nacimiento: nacimiento,
		numero: numero,
	}
	return bet
}

func NewLottery(ServerAddress string, bet Bet, agencyId string) *Lottery {
	lottery := &Lottery{
		ServerAddress: ServerAddress,
		bet: bet,
		agencyId: agencyId,
	}
	return lottery
}

func SerializeBet(agencyId string, bet *Bet) []byte {
	var serializeBet []byte

	// Soprtomos campos de longitud maxima 255
	serializeBet = append(serializeBet, byte(len(agencyId)))
	serializeBet = append(serializeBet, []byte(agencyId)...)
	
	serializeBet = append(serializeBet, byte(len(bet.nombre)))
	serializeBet = append(serializeBet, []byte(bet.nombre)...)

	serializeBet = append(serializeBet, byte(len(bet.apellido)))
	serializeBet = append(serializeBet, []byte(bet.apellido)...)

	serializeBet = append(serializeBet, byte(len(bet.documento)))
	serializeBet = append(serializeBet, []byte(bet.documento)...)

	serializeBet = append(serializeBet, byte(len(bet.nacimiento)))
	serializeBet = append(serializeBet, []byte(bet.nacimiento)...)

	serializeBet = append(serializeBet, byte(len(bet.numero)))
	serializeBet = append(serializeBet, []byte(bet.numero)...)

	
	return serializeBet
}

func (l *Lottery) SendBet() {
	// Handle SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		log.Infof("Received SIGTERM, shutting down gracefully...")
		// Acá debería cerrar el socket que creo tiene abierto el cliente si es que tiene
		if l.conn != nil {
			error := l.conn.Close()
			if error == nil {
				log.Info("action: close_connection | result: success | agency_id: %v", l.agencyId)
			}
		}
		os.Exit(0)
	}()

	// Connect to server
	conn, err := net.Dial("tcp", l.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | agency_id: %v | error: %v",
			l.agencyId,
			err,
		)
	}
	l.conn = conn

	// Serialize Bet
	betSerialized := SerializeBet(l.agencyId, &l.bet)
	lengthBet := len(betSerialized)
	
	// Send bet
	for lengthBet > 0 {
		n, err := l.conn.Write(betSerialized)
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v",
				l.bet.documento,
				l.bet.numero,
				err,
			)
			return
		}
		lengthBet -= n
		betSerialized = betSerialized[n:]
	}

	// CLose conection
	l.conn.Close()
	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
		l.bet.documento,
		l.bet.numero,
	)
}