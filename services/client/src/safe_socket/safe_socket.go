package safe_socket

import "io"

func SendAll(socket io.Writer, bytes []byte) error {
	total := 0

	for {
		n, err := socket.Write(bytes[total:])

		total += n

		if total == len(bytes) {
			return nil
		}

		if err != nil {
			return err
		}

	}

}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	total := 0
	for {
		n, err := socket.Read(buff[total:])

		total += n

		if total == size {
			return buff, nil
		}

		if err != nil {
			return buff[:total], err
		}

	}

}
