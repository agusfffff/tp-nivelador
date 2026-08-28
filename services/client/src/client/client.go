package client

import (
	"io"
	"net"
	"time"

	"bufio"
	"os"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 500

const ECHO_CLIENT_BUFFER_SIZE = 512

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
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
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

func get_input_file(inputFilePath string) (*os.File, error) {
	file, err := os.Open(inputFilePath)

	if err != nil {
		logger.Error("Error opening file", logger.Fail)
		return nil, err
	}

	defer file.Close()

	return file, nil
}

func get_output_file(outputFilePath string) (*os.File, error) {
	file, err := os.Create(outputFilePath)

	if err != nil {
		logger.Error("Error creating file", logger.Fail)
		return nil, err
	}

	defer file.Close()

	return file, nil
}

func (client *Client) Run() error {
	const mainAction = "test-echo-server"
	defer client.conn.Close()

	inputFile, err := get_input_file(client.config.InputFile)

	if err != nil {
		return err
	}

	outputFile, err := get_output_file(client.config.OutputFile)

	if err != nil {
		return err
	}

	reader := bufio.NewReader(inputFile)

	writer := bufio.NewWriter(outputFile)

	for {

		line, readErr := client.read_line(reader)

		if len(line) == 0 {
			break
		}

		response, procErr := client.processLine(line)
		if procErr != nil {
			return procErr
		}

		if err := client.write_line(response, writer); err != nil {
			return err
		}

		if readErr == io.EOF {
			break
		}
	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}

func (client *Client) processLine(line string) (string, error) {
	err := safe_socket.SendAll(client.conn, []byte(line))

	if err != nil {
		logger.Error("send-message", logger.Fail)
		return "", err
	}

	responseBuffer, err := safe_socket.RecvAll(client.conn, ECHO_CLIENT_BUFFER_SIZE)

	responseString := string(responseBuffer)

	if err != nil {
		logger.Error("recv-response", logger.Fail)
		return responseString, err
	}

	return responseString, nil
}

func (client *Client) write_line(responseBuffer string, writer *bufio.Writer) error {
	_, errWrite := writer.WriteString(string(responseBuffer))

	if errWrite != nil {
		logger.Error("write-output", logger.Fail)
		return errWrite
	}

	errFlush := writer.Flush()

	if errFlush != nil {
		logger.Error("flush-output", logger.Fail)
		return errFlush
	}
	return nil
}

func (client *Client) read_line(reader *bufio.Reader) (string, error) {

	line, readErr := reader.ReadString('\n')

	if readErr != nil && readErr != io.EOF {
		logger.Error("read-input", logger.Fail)
		return "", readErr
	}

	return line, nil
}
