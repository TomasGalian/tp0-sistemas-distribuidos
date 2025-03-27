package common

type Bet struct {
	nombre      string
	apellido    string
	documento   string
	nacimiento  string
	numero      string
}

// Birthadte format yyyy-mm-dd: max 10 bytes
// Document format: xx.xxx.xxx: max 8
func NewBet(nombre string, apellido string, documento string, nacimiento string, numero string) *Bet {
	if len(nombre) > 30 || len(apellido) >= 15 || len(documento) > 8 || len(nacimiento) > 10 || len(numero) > 10 {
		return nil
	}

	bet := &Bet{
		nombre:     nombre,
		apellido:   apellido,
		documento:  documento,
		nacimiento: nacimiento,
		numero:     numero,
	}
	return bet
}

func (bet *Bet) serializeBet(agencyID string) []byte {
	var serializedBet []byte

	// Use 1 byte for action
	// 0x01 for send bet
	// 0x11 for send ACK
	// serializedBet = append(serializedBet, 0x01)

	// The rest of the fields are serialized as: length + data
	serializedBet = append(serializedBet, byte(len(agencyID))) // 1 byte
	serializedBet = append(serializedBet, []byte(agencyID)...) // 1 bytes

	serializedBet = append(serializedBet, byte(len(bet.nombre))) // 1 byte
	serializedBet = append(serializedBet, []byte(bet.nombre)...) // 30 byte max

	serializedBet = append(serializedBet, byte(len(bet.apellido))) // 1 byte
	serializedBet = append(serializedBet, []byte(bet.apellido)...) // 15 byte max

	serializedBet = append(serializedBet, byte(len(bet.documento))) // 1 byte
	serializedBet = append(serializedBet, []byte(bet.documento)...) // 8 byte max

	serializedBet = append(serializedBet, byte(len(bet.nacimiento))) // 1 byte
	serializedBet = append(serializedBet, []byte(bet.nacimiento)...) // 10 byte max

	serializedBet = append(serializedBet, byte(len(bet.numero))) // 1 byte
	serializedBet = append(serializedBet, []byte(bet.numero)...) // 10 byte max

	//Max bytes 1 + 1 + 30 + 1 + 15 + 1 + 8 + 1 + 10 + 1 + 10 = 78 + 1 Byte for the action = 79
	// 8kB / 79 = 101 bets 

	return serializedBet
}