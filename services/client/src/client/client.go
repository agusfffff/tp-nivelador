package client

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"bufio"
	"os"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 500

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
}

type Client struct {
	conn   net.Conn
	config ClientConfig
	agency byte
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	agencyId, err := strconv.Atoi(config.AgencyId)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn, config: config, agency: byte(agencyId)}, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Run() error {
	defer client.conn.Close()

	if err := client.sendBets(); err != nil {
		return err
	}

	if err := safe_socket.SendAll(client.conn, protocol.EncodeEnd()); err != nil {
		logger.Error("send-end", logger.Fail)
		return err
	}

	if err := client.receiveWinners(); err != nil {
		return err
	}

	logger.Info("client-run", logger.Success, "agency-id", client.config.AgencyId)
	return nil
}

func (client *Client) sendBets() error {
	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error("open-input-file", logger.Fail)
		return err
	}
	defer inputFile.Close()

	reader := bufio.NewReader(inputFile)

	for {
		line, readErr := reader.ReadString('\n')
		trimmed := strings.TrimRight(line, "\r\n")

		if trimmed != "" {
			row := strings.Split(trimmed, ",")

			documento, err := strconv.ParseUint(row[2], 10, 32)
			if err != nil {
				return err
			}
			numberValue, err := strconv.ParseUint(row[4], 10, 16)
			if err != nil {
				return err
			}

			message := protocol.EncodeBet(protocol.BetMessage{
				Agency:     client.agency,
				Nombre:     row[0],
				Apellido:   row[1],
				Documento:  uint32(documento),
				Cumpleanos: row[3],
				Number:     uint16(numberValue),
			})
			if err := safe_socket.SendAll(client.conn, message); err != nil {
				logger.Error("send-bet", logger.Fail)
				return err
			}

			tipo, payload, err := protocol.ReadMessage(client.conn)
			if err != nil {
				logger.Error("recv-bet-ack", logger.Fail)
				return err
			}
			if tipo != protocol.BetAck {
				return fmt.Errorf("mensaje inesperado del servidor: %d", tipo)
			}
			protocol.DecodeBetAck(payload)
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			logger.Error("read-input", logger.Fail)
			return readErr
		}
	}

	return nil
}

func (client *Client) receiveWinners() error {
	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Error("create-output-file", logger.Fail)
		return err
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)

	for {
		tipo, payload, err := protocol.ReadMessage(client.conn)
		if err != nil {
			logger.Error("recv-message", logger.Fail)
			return err
		}

		if tipo == protocol.End {
			break
		}

		if tipo != protocol.Winner {
			return fmt.Errorf("mensaje inesperado del servidor: %d", tipo)
		}

		winner := protocol.DecodeWinner(payload)
		row := []string{
			winner.Nombre,
			winner.Apellido,
			strconv.FormatUint(uint64(winner.Documento), 10),
			winner.Cumpleanos,
			strconv.FormatUint(uint64(winner.Number), 10),
		}

		if _, err := writer.WriteString(strings.Join(row, ",") + "\n"); err != nil {
			logger.Error("write-output", logger.Fail)
			return err
		}
		if err := writer.Flush(); err != nil {
			logger.Error("flush-output", logger.Fail)
			return err
		}

		if err := safe_socket.SendAll(client.conn, protocol.EncodeBetAck(winner.Number)); err != nil {
			logger.Error("send-winner-ack", logger.Fail)
			return err
		}
	}

	return nil
}
