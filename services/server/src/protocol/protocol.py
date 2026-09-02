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
from typing import NamedTuple

import safe_socket

BET = 1
WINNER = 2
BET_ACK = 3
END = 4
BATCH = 5
_HEADER_SIZE = 3


class BetMessage(NamedTuple):
    agency: int
    nombre: str
    apellido: str
    documento: int
    cumpleanos: str
    number: int

class Winner(NamedTuple):
    nombre: str
    apellido: str
    documento: int
    cumpleanos: str
    number: int


def read_message(sock) -> BetMessage | int | None:
    header = safe_socket.recv_all(sock, _HEADER_SIZE)
    tipo = int.from_bytes(header[0:1], 'big')
    largo_payload = int.from_bytes(header[1:3], 'big')
    payload = safe_socket.recv_all(sock, largo_payload)
    return decode_message(tipo, payload)


def encode_message(tipo: int, payload: bytes) -> bytes:
    return tipo.to_bytes(1, 'big') + len(payload).to_bytes(2, 'big') + payload


def encode_end() -> bytes:
    return encode_message(END, b'')

def encode_bet_ack(number) -> bytes:
    return encode_message(BET_ACK, number.to_bytes(2, 'big'))

def encode_winner(winner: Winner) -> bytes:
    payload = (
        _encode_prefixed_field(winner.nombre) +
        _encode_prefixed_field(winner.apellido) +
        winner.documento.to_bytes(4, 'big') +
        encode_birthdate(winner.cumpleanos) +
        winner.number.to_bytes(2, 'big')
    )

    return encode_message(WINNER, payload)

def _encode_prefixed_field(value: str) -> bytes:
    value_bytes = value.encode('utf-8')
    return len(value_bytes).to_bytes(1, 'big') + value_bytes

def encode_birthdate(cumpleanos) -> bytes:
    " de AAAA-MM-DD a AAAAMMDD bytes"
    cumpleanos_int = int(cumpleanos.replace("-", ""))
    return cumpleanos_int.to_bytes(4, 'big')

def decode_bet(payload) -> BetMessage:
    agency = int.from_bytes(payload[0:1], 'big')
    nombre, pos = _read_prefixed_field(payload, 1)
    apellido, pos = _read_prefixed_field(payload, pos)
    documento = int.from_bytes(payload[pos:pos + 4], 'big')
    pos += 4
    cumpleanos = decode_birthdate(payload[pos:pos + 4])
    pos += 4
    number = int.from_bytes(payload[pos:pos + 2], 'big')

    return BetMessage(agency,nombre,apellido,documento,cumpleanos,number)


def _decode_bet_at(payload, pos) -> tuple[BetMessage, int]:
    agency = int.from_bytes(payload[pos:pos + 1], 'big')
    pos += 1
    nombre, pos = _read_prefixed_field(payload, pos)
    apellido, pos = _read_prefixed_field(payload, pos)
    documento = int.from_bytes(payload[pos:pos + 4], 'big')
    pos += 4
    cumpleanos = decode_birthdate(payload[pos:pos + 4])
    pos += 4
    number = int.from_bytes(payload[pos:pos + 2], 'big')
    pos += 2

    return BetMessage(agency, nombre, apellido, documento, cumpleanos, number), pos


def decode_batch(payload) -> list[BetMessage]:
    bets = []
    pos = 0
    while pos < len(payload):
        bet, pos = _decode_bet_at(payload, pos)
        bets.append(bet)
    return bets

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


def decode_message(tipo, payload) -> BetMessage | int | None:
    if tipo == BATCH:
        return decode_batch(payload)
    elif tipo == BET_ACK:
        return decode_bet_ack(payload)
    elif tipo == END:
        return None
    else:
        raise ValueError(f"Tipo de mensaje desconocido: {tipo}")


def _read_prefixed_field(payload, pos):
    length = payload[pos]
    start = pos + 1
    end = start + length
    return payload[start:end].decode('utf-8'), end
