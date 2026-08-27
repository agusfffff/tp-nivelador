package safe_socket

import "io"

func SendAll(socket io.Writer, bytes []byte) error {
	total := 0

	for {
		n, err := socket.Write(bytes[total:])

		if err != nil {
			return err
		}

		total += n

		if total == len(bytes) {
			return nil
		}
	}

}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	total := 0
	for {
		n, err := socket.Read(buff[total:])

		if err != nil && err != io.EOF {
			return buff, err
		}

		if err == io.EOF && n == 0 {
			return buff, nil
		}

		total += n

		if total == size {
			return buff, nil
		}

	}

}
