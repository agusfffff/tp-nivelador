package protocol

/* HEADER (3 bytes)
  tipo:          1 byte   (pocos valores)
  largo payload: 2 bytes

PAYLOAD (variable, tamaño = largo payload)
  agency:        1 byte
  largo nombre:  1 byte
  nombre:        N bytes   (UTF-8, variable)
  largo apellido:1 byte
  apellido:      M bytes   (UTF-8, variable)
  documento:     4 bytes
  cumpleaños:    4 bytes  (AAAAMMDD como un solo entero)
  number:        4 bytes

header type:
1 = BET       (cliente → servidor, una apuesta)
4 = END       (cliente → servidor, servidor → cliente, "ya mandé todo")
2 = WINNER   (servidor → cliente, un winner)
3 = BET_ACK   (servidor → cliente, confirmación de éxito X cada BET recibido) */

import (
	"encoding/binary"
	"net"
	"strconv"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const Bet = 1
const Winner = 2
const Ack = 3
const End = 4
const Batch = 5
const headerSize = 3

type WinnerMessage struct {
	Nombre     string
	Apellido   string
	Documento  uint32
	Cumpleanos string
	Number     uint32
}

type BetMessage struct {
	Agency     byte
	Nombre     string
	Apellido   string
	Documento  uint32
	Cumpleanos string
	Number     uint32
}

func ReadMessage(sock net.Conn) (int, []byte, error) {
	header, err := safe_socket.RecvAll(sock, headerSize)

	if err != nil {
		return 0, nil, err
	}

	tipo := int(header[0])

	largo_payload := int(binary.BigEndian.Uint16(header[1:3]))

	payload, err := safe_socket.RecvAll(sock, largo_payload)

	if err != nil {
		return 0, nil, err
	}

	return tipo, payload, nil
}

func EncodeEnd() []byte {
	return newMessage(End, 0)
}

func EncodeAck(number uint32) []byte {
	message := newMessage(Ack, 4)
	binary.BigEndian.PutUint32(message[headerSize:], number)
	return message
}

func AppendBatch(buf []byte, bets []BetMessage) []byte {
	headerPos := len(buf)
	buf = append(buf, Batch, 0, 0)

	payloadStart := len(buf)
	for _, bet := range bets {
		buf = appendBetPayload(buf, bet)
	}

	payloadSize := len(buf) - payloadStart
	binary.BigEndian.PutUint16(buf[headerPos+1:headerPos+3], uint16(payloadSize))

	return buf
}

func newMessage(tipo byte, payloadSize int) []byte {
	message := make([]byte, headerSize+payloadSize)
	message[0] = tipo
	binary.BigEndian.PutUint16(message[1:3], uint16(payloadSize))
	return message
}

func appendBetPayload(buf []byte, bet BetMessage) []byte {
	buf = append(buf, bet.Agency)
	buf = appendPrefixedField(buf, bet.Nombre)
	buf = appendPrefixedField(buf, bet.Apellido)
	buf = binary.BigEndian.AppendUint32(buf, bet.Documento)
	buf = binary.BigEndian.AppendUint32(buf, encodeBirthdate(bet.Cumpleanos))
	buf = binary.BigEndian.AppendUint32(buf, bet.Number)
	return buf
}

// encodeBirthdate arma el entero AAAAMMDD a partir de "AAAA-MM-DD",
// recorriendo el string dígito a dígito e ignorando los guiones. Evita
// depender de strings.ReplaceAll, que asignaría un string nuevo por cada
// apuesta procesada.
func encodeBirthdate(cumpleanos string) uint32 {
	var value uint32
	for _, char := range cumpleanos {
		if char == '-' {
			continue
		}
		value = value*10 + uint32(char-'0')
	}
	return value
}

func DecodeWinner(payload []byte) WinnerMessage {
	nombre, pos := readPrefixedField(payload, 0)
	apellido, pos := readPrefixedField(payload, pos)

	documento := binary.BigEndian.Uint32(payload[pos : pos+4])
	pos += 4

	cumpleanos := decodeBirthdate(payload[pos : pos+4])
	pos += 4

	number := binary.BigEndian.Uint32(payload[pos : pos+4])

	return WinnerMessage{nombre, apellido, documento, cumpleanos, number}
}

func decodeBirthdate(birthdayBytes []byte) string {
	value := binary.BigEndian.Uint32(birthdayBytes)
	str := strconv.FormatUint(uint64(value), 10)
	return str[0:4] + "-" + str[4:6] + "-" + str[6:8]
}

func DecodeAck(payload []byte) uint16 {
	return binary.BigEndian.Uint16(payload)
}

func appendPrefixedField(buf []byte, value string) []byte {
	buf = append(buf, byte(len(value)))
	buf = append(buf, value...)
	return buf
}

func readPrefixedField(payload []byte, pos int) (string, int) {
	length := int(payload[pos])
	start := pos + 1
	end := start + length
	return string(payload[start:end]), end
}
