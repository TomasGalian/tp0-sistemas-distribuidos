package common

type Bet struct {
	nombre      string
	apellido    string
	documento   string
	nacimiento  string
	numero      string
}

func NewBet(nombre string, apellido string, documento string, nacimiento string, numero string) *Bet {
	bet := &Bet{
		nombre:     nombre,
		apellido:   apellido,
		documento:  documento,
		nacimiento: nacimiento,
		numero:     numero,
	}
	return bet
}

func serializeBet(agencyID string, bet *Bet) []byte {
	var serializedBet []byte

	// Soprtomos campos de longitud maxima 255
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