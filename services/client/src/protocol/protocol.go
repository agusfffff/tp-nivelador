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
  number:        2 bytes

header type:
1 = BET       (cliente → servidor, una apuesta)
4 = END       (cliente → servidor, servidor → cliente, "ya mandé todo")
2 = WINNER   (servidor → cliente, un winner)
3 = BET_ACK   (servidor → cliente, confirmación de éxito X cada BET recibido) */

import (
	"encoding/binary"
	"net"
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const Bet = 1
const Winner = 2
const BetAck = 3
const End = 4
const headerSize = 3

type WinnerMessage struct {
	Nombre     string
	Apellido   string
	Documento  uint32
	Cumpleanos string
	Number     uint16
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

func packMessage(tipo byte, payload []byte) []byte {
	header := make([]byte, headerSize)

	header[0] = tipo

	binary.BigEndian.PutUint16(header[1:3], uint16(len(payload)))

	return append(header, payload...)
}

func EncodeEnd() []byte {
	return packMessage(End, []byte{})
}

func EncodeBetAck(number uint16) []byte {
	numberBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(numberBytes, number)
	return packMessage(BetAck, numberBytes)
}

func EncodeBet(agency byte, nombre, apellido string, documento uint32, cumpleanos string, number uint16) []byte {
	nombreBytes := []byte(nombre)
	apellidoBytes := []byte(apellido)

	payload := make([]byte, 0, 3+len(nombreBytes)+len(apellidoBytes)+10)

	payload = append(payload, agency)

	payload = append(payload, byte(len(nombreBytes)))
	payload = append(payload, nombreBytes...)

	payload = append(payload, byte(len(apellidoBytes)))
	payload = append(payload, apellidoBytes...)

	documentoBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(documentoBytes, documento)
	payload = append(payload, documentoBytes...)

	payload = append(payload, encodeBirthdate(cumpleanos)...)

	numberBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(numberBytes, number)

	payload = append(payload, numberBytes...)

	return packMessage(Bet, payload)
}

func encodeBirthdate(cumpleanos string) []byte {
	digits := strings.ReplaceAll(cumpleanos, "-", "")
	value, _ := strconv.ParseUint(digits, 10, 32)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(value))
	return buf
}

func DecodeWinner(payload []byte) WinnerMessage {
	largoNombre := int(payload[0])
	nombre := string(payload[1 : 1+largoNombre])

	pos := 1 + largoNombre
	largoApellido := int(payload[pos])
	pos++
	apellido := string(payload[pos : pos+largoApellido])
	pos += largoApellido

	documento := binary.BigEndian.Uint32(payload[pos : pos+4])
	pos += 4

	cumpleanos := decodeBirthdate(payload[pos : pos+4])
	pos += 4

	number := binary.BigEndian.Uint16(payload[pos : pos+2])

	return WinnerMessage{nombre, apellido, documento, cumpleanos, number}
}

func decodeBirthdate(birthdayBytes []byte) string {
	value := binary.BigEndian.Uint32(birthdayBytes)
	str := strconv.FormatUint(uint64(value), 10)
	return str[0:4] + "-" + str[4:6] + "-" + str[6:8]
}

func DecodeBetAck(payload []byte) uint16 {
	return binary.BigEndian.Uint16(payload)
}
