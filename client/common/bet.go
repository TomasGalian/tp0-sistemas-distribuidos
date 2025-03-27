package common

type Bet struct {
	nombre      string
	apellido    string
	documento   string
	nacimiento  string
	numero      string
}

func NewBet(nombre string, apellido string, documento string, nacimiento string, numero string) *Bet {
	if len(nombre) >= 255 || len(apellido) >= 255 || len(documento) >= 255 || len(nacimiento) >= 255 || len(numero) >= 255 {
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
	serializedBet = append(serializedBet, 0x01)

	// The rest of the fields are serialized as: length + data
	serializedBet = append(serializedBet, byte(len(agencyID)))
	serializedBet = append(serializedBet, []byte(agencyID)...)

	serializedBet = append(serializedBet, byte(len(bet.nombre)))
	serializedBet = append(serializedBet, []byte(bet.nombre)...)

	serializedBet = append(serializedBet, byte(len(bet.apellido)))
	serializedBet = append(serializedBet, []byte(bet.apellido)...)

	serializedBet = append(serializedBet, byte(len(bet.documento)))
	serializedBet = append(serializedBet, []byte(bet.documento)...)

	serializedBet = append(serializedBet, byte(len(bet.nacimiento)))
	serializedBet = append(serializedBet, []byte(bet.nacimiento)...)

	serializedBet = append(serializedBet, byte(len(bet.numero)))
	serializedBet = append(serializedBet, []byte(bet.numero)...)

	return serializedBet
}