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
3 = ACK   (cliente → servidor, servidor → cliente, confirmación de éxito, sin payload) */

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const Bet = 1
const Winner = 2
const Ack = 3
const End = 4
const Batch = 5
const headerSize = 3
const MaxPayloadSize = 0xFFFF
const MaxFieldLength = 0xFF

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

func EncodeAck() []byte {
	return newMessage(Ack, 0)
}

func EncodeBet(bet BetMessage) ([]byte, error) {
	var buf []byte
	buf = append(buf, bet.Agency)

	buf, err := appendPrefixedField(buf, bet.Nombre)
	if err != nil {
		return nil, fmt.Errorf("nombre: %w", err)
	}
	buf, err = appendPrefixedField(buf, bet.Apellido)
	if err != nil {
		return nil, fmt.Errorf("apellido: %w", err)
	}

	buf = binary.BigEndian.AppendUint32(buf, bet.Documento)
	buf = binary.BigEndian.AppendUint32(buf, encodeBirthdate(bet.Cumpleanos))
	buf = binary.BigEndian.AppendUint32(buf, bet.Number)
	return buf, nil
}

func AppendBatch(buf []byte, encodedBets [][]byte) ([]byte, error) {
	headerPos := len(buf)
	buf = append(buf, Batch, 0, 0)

	payloadStart := len(buf)
	for _, encodedBet := range encodedBets {
		buf = append(buf, encodedBet...)
	}

	payloadSize := len(buf) - payloadStart
	if payloadSize > MaxPayloadSize {
		return nil, fmt.Errorf("batch payload de %d bytes excede el máximo de %d", payloadSize, MaxPayloadSize)
	}
	binary.BigEndian.PutUint16(buf[headerPos+1:headerPos+3], uint16(payloadSize))

	return buf, nil
}

func newMessage(tipo byte, payloadSize int) []byte {
	message := make([]byte, headerSize+payloadSize)
	message[0] = tipo
	binary.BigEndian.PutUint16(message[1:3], uint16(payloadSize))
	return message
}

func encodeBirthdate(cumpleanos string) uint32 {
	digits := strings.ReplaceAll(cumpleanos, "-", "")
	value, _ := strconv.ParseUint(digits, 10, 32)
	return uint32(value)
}

func DecodeWinner(payload []byte) (WinnerMessage, error) {
	nombre, pos, err := readPrefixedField(payload, 0)
	if err != nil {
		return WinnerMessage{}, fmt.Errorf("nombre: %w", err)
	}
	apellido, pos, err := readPrefixedField(payload, pos)
	if err != nil {
		return WinnerMessage{}, fmt.Errorf("apellido: %w", err)
	}

	documentoBytes, pos, err := take(payload, pos, 4)
	if err != nil {
		return WinnerMessage{}, fmt.Errorf("documento: %w", err)
	}
	documento := binary.BigEndian.Uint32(documentoBytes)

	cumpleanosBytes, pos, err := take(payload, pos, 4)
	if err != nil {
		return WinnerMessage{}, fmt.Errorf("cumpleanos: %w", err)
	}
	cumpleanos := decodeBirthdate(cumpleanosBytes)

	numberBytes, _, err := take(payload, pos, 4)
	if err != nil {
		return WinnerMessage{}, fmt.Errorf("number: %w", err)
	}
	number := binary.BigEndian.Uint32(numberBytes)

	return WinnerMessage{nombre, apellido, documento, cumpleanos, number}, nil
}

func decodeBirthdate(birthdayBytes []byte) string {
	value := binary.BigEndian.Uint32(birthdayBytes)
	str := strconv.FormatUint(uint64(value), 10)
	return str[0:4] + "-" + str[4:6] + "-" + str[6:8]
}

func appendPrefixedField(buf []byte, value string) ([]byte, error) {
	if len(value) > MaxFieldLength {
		return nil, fmt.Errorf("campo de %d bytes excede el máximo de %d", len(value), MaxFieldLength)
	}
	buf = append(buf, byte(len(value)))
	buf = append(buf, value...)
	return buf, nil
}

func readPrefixedField(payload []byte, pos int) (string, int, error) {
	lengthByte, pos, err := take(payload, pos, 1)
	if err != nil {
		return "", 0, err
	}
	length := int(lengthByte[0])

	value, pos, err := take(payload, pos, length)
	if err != nil {
		return "", 0, err
	}
	return string(value), pos, nil
}

func take(payload []byte, pos int, n int) ([]byte, int, error) {
	end := pos + n
	if end > len(payload) {
		return nil, 0, fmt.Errorf("payload truncado: se esperaban %d bytes en la posición %d, quedan %d", n, pos, len(payload)-pos)
	}
	return payload[pos:end], end, nil
}
