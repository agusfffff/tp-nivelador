""" 
HEADER (3 bytes)
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
3 = BET_ACK   (servidor → cliente, confirmación de éxito X cada BET recibido)

"""
import safe_socket

BET = 1 
WINNER = 2
BET_ACK = 3
END = 4
_HEADER_SIZE = 3



def read_message(sock) -> tuple:
    header = safe_socket.recv_all(sock, _HEADER_SIZE)
    tipo = decode_tipo_header(header[0:1])
    largo_payload = decode_largo_payload(header[1:3])
    payload = safe_socket.recv_all(sock, largo_payload)
    data = decode_message(tipo, payload) 
    return data


def pack_message(tipo: int, payload: bytes) -> bytes:
    return tipo.to_bytes(1, 'big') + len(payload).to_bytes(2, 'big') + payload


def encode_end() ->bytes:
    return pack_message(4, b'')

def encode_bet_ack(number) -> bytes:
    return pack_message(3, number.to_bytes(2, 'big'))

def encode_winner(nombre, apellido, documento, cumpleanos, number) -> bytes:

    payload = encode_client(nombre, apellido, documento, cumpleanos, number)

    return pack_message(2, payload)


def encode_client(nombre, apellido, documento, cumpleanos, number)->bytes: 
    nombre_bytes = nombre.encode('utf-8')
    apellido_bytes = apellido.encode('utf-8')

    payload = (
        len(nombre_bytes).to_bytes(1, 'big') +
        nombre_bytes +
        len(apellido_bytes).to_bytes(1, 'big') +
        apellido_bytes +
        documento.to_bytes(4, 'big') +
        encode_birthdate(cumpleanos) +
        number.to_bytes(2, 'big')
    ) 
    return payload

def encode_bet(agency, nombre, apellido, documento, cumpleanos, number) -> bytes:
    payload = (
        agency.to_bytes(1, 'big') +
        encode_client(nombre, apellido, documento, cumpleanos, number)
    )

    return pack_message(1, payload)

def encode_birthdate(cumpleanos) -> bytes:
    " de AAAA-MM-DD a AAAAMMDD bytes"
    cumpleanos_int = int(cumpleanos.replace("-", ""))
    return cumpleanos_int.to_bytes(4, 'big')

def decode_tipo_header(bytes_data) -> int: 
    tipo_bytes = bytes_data
    tipo = int.from_bytes(tipo_bytes, 'big')

    return tipo

def decode_largo_payload(bytes_data) -> int:
    largo_payload_bytes = bytes_data
    largo_payload = int.from_bytes(largo_payload_bytes, 'big')

    return largo_payload

def decode_winner(payload) -> tuple:
    return decode_client(payload)

def decode_bet(payload) -> tuple:
    agency_bytes = payload[0:1]
    agency = int.from_bytes(agency_bytes, 'big')

    nombre, apellido, documento, birthday, number = decode_client(payload[1:])

    return (agency, nombre, apellido, documento, birthday, number)



def decode_client(bytes) -> tuple: 
    largo_nombre_bytes = bytes[0:1]
    largo_nombre = int.from_bytes(largo_nombre_bytes, 'big')

    end_nombre= 1+largo_nombre
    nombre_bytes = bytes[1:end_nombre]
    nombre = nombre_bytes.decode('utf-8')

    end_largo_apellido = end_nombre+1
    largo_apellido_bytes = bytes[end_nombre:end_largo_apellido]
    largo_apellido = int.from_bytes(largo_apellido_bytes, 'big')

    end_apellido = end_largo_apellido + largo_apellido
    apellido_bytes = bytes[end_largo_apellido:end_apellido]
    apellido = apellido_bytes.decode('utf-8')

    end_documento= end_apellido+4
    documento_bytes = bytes[end_apellido:end_documento]
    documento = int.from_bytes(documento_bytes, 'big')

    end_birthday= end_documento+4
    birthday_bytes = bytes [end_documento:end_birthday]
    birthday = decode_birthdate(birthday_bytes)

    number_bytes = bytes[end_birthday:end_birthday+2]
    number = int.from_bytes(number_bytes, 'big')

    return (nombre, apellido, documento, birthday, number)

def decode_birthdate(birthday_bytes) -> str:
    "de AAAAMMDD a AAAA-MM-DD"
    birthday =int.from_bytes(birthday_bytes, 'big')
    birthday_str = str(birthday)
    year = birthday_str[:4]
    month = birthday_str[4:6] 
    day = birthday_str[6:8]
    return f"{year}-{month}-{day}"


def decode_bet_ack(bytes_data) -> int:
    number = int.from_bytes(bytes_data, 'big')
    return number


def decode_message(tipo, payload) -> tuple:
    if tipo == BET:
        return decode_bet(payload)
    elif tipo == BET_ACK: 
        return decode_bet_ack(payload)
    elif tipo == END:
        return None
    else:
        raise ValueError(f"Tipo de mensaje desconocido: {tipo}")
