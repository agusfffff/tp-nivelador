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

def encode_end() ->bytes: 
    payload = 0
    largo_payload = len(payload).to_bytes(2, 'big')

    tipo = (4).to_bytes(1, 'big')
    return tipo + largo_payload 

def encode_bet_ack(number) -> bytes:
    payload = number.to_bytes(2, 'big')
    largo_payload = len(payload).to_bytes(2, 'big')

    tipo = (3).to_bytes(1, 'big')
    return tipo + largo_payload + payload

def encode_winner(nombre, apellido, documento, cumpleanos, number) -> bytes:
    nombre_bytes = nombre.encode('utf-8')
    apellido_bytes = apellido.encode('utf-8')

    payload = (
        encode_client(nombre_bytes, apellido_bytes, documento, cumpleanos, number)
    )

    tipo = (2).to_bytes(1, 'big')
    largo_payload = len(payload).to_bytes(2, 'big')

    return tipo + largo_payload + payload


def encode_client(nombre, apellido, documento, cumpleanos, number)->bytes: 
    payload = (
        len(nombre).to_bytes(1, 'big') +
        nombre +
        len(apellido).to_bytes(1, 'big') +
        apellido +
        documento.to_bytes(4, 'big') +
        encode_birthdate(cumpleanos) +
        number.to_bytes(2, 'big')
    ) 
    return payload

def encode_bet(agency, nombre, apellido, documento, cumpleanos, number) -> bytes:
    nombre_bytes = nombre.encode('utf-8')
    apellido_bytes = apellido.encode('utf-8')

    payload = (
        agency.to_bytes(1, 'big') +
        encode_client(nombre_bytes, apellido_bytes, documento, cumpleanos, number)
    )

    largo_payload = len(payload).to_bytes(2, 'big')

    tipo = (1).to_bytes(1, 'big')

    return tipo + largo_payload + payload

def encode_birthdate(cumpleanos) -> bytes:
    " de AAAA-MM-DD a AAAAMMDD bytes"
    cumpleanos_int = int(cumpleanos.replace("-", ""))
    return cumpleanos_int.to_bytes(4, 'big')

def decode_header(bytes_data) -> tuple:
    tipo_bytes = bytes_data[0:1]
    tipo = int.from_bytes(tipo_bytes, 'big')

    largo_bytes= bytes_data[1:3]
    largo_payload = int.from_bytes(largo_bytes, 'big')

    return tipo, largo_payload

def decode_bet(largo_payload, bytes_data) -> tuple:
    payload = bytes_data[3:3+largo_payload]  

    agency_bytes = payload[0:1]
    agency = int.from_bytes(agency_bytes, 'big')

    largo_nombre_bytes = payload[1:2]
    largo_nombre = int.from_bytes(largo_nombre_bytes, 'big')

    end_nombre= 2+largo_nombre
    nombre_bytes = payload[2:end_nombre]
    nombre = nombre_bytes.decode('utf-8')

    end_largo_apellido = end_nombre+1
    largo_apellido_bytes = payload[end_nombre:end_largo_apellido]
    largo_apellido = int.from_bytes(largo_apellido_bytes, 'big')

    end_apellido = end_largo_apellido + largo_apellido
    apellido_bytes = payload[end_largo_apellido:end_apellido]
    apellido = apellido_bytes.decode('utf-8')

    end_documento= end_apellido+4
    documento_bytes = payload[end_apellido:end_documento]
    documento = int.from_bytes(documento_bytes, 'big')

    end_birthday= end_documento+4
    birthday_bytes = payload [end_documento:end_birthday]
    birthday = decode_birthdate(birthday_bytes)

    number_bytes = payload[end_birthday:end_birthday+2]
    number = int.from_bytes(number_bytes, 'big')

    return (agency, nombre, apellido, documento, birthday, number)

def decode_birthdate(birthday_bytes) -> str:
    "de AAAAMMDD a AAAA-MM-DD"
    birthday =int.from_bytes(birthday_bytes, 'big')
    birthday_str = str(birthday)
    year = birthday_str[:4]
    month = birthday_str[4:6] 
    day = birthday_str[6:8]
    return f"{year}-{month}-{day}"