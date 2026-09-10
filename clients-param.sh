#!/bin/bash

set -euo pipefail

if [ $# -lt 1 ] || [ $# -gt 2 ]; then
  echo "Uso: $0 <cantidad_de_clientes> [agency_quorum_min]" >&2
    exit 1
fi

CLIENT_AMOUNT=$1 
AGENCY_QUORUM_MIN=${2:-1}

if ! [[ "$CLIENT_AMOUNT" =~ ^[0-9]+$ ]] || [ "$CLIENT_AMOUNT" -le 0 ]; then
    echo "Error: la cantidad de clientes debe ser un numero entero positivo" >&2
    exit 1
fi

if ! [[ "$AGENCY_QUORUM_MIN" =~ ^[0-9]+$ ]] || [ "$AGENCY_QUORUM_MIN" -le 0 ] || [ "$AGENCY_QUORUM_MIN" -gt "$CLIENT_AMOUNT" ]; then
  echo "Error: agency_quorum_min debe ser mayor que 0 y menor o igual que la cantidad de clientes" >&2
  exit 1
fi

LAST_CLIENT_INDEX=$(($CLIENT_AMOUNT - 1))

OUTPUT_FILE="docker-compose.yaml"

cat > "$OUTPUT_FILE" <<EOF
services:
  server:
    build:
      context: ./services/server
      dockerfile: Dockerfile
    container_name: server
    ports: 
      - "5678:5678"
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - AGENCY_QUORUM_MIN=$AGENCY_QUORUM_MIN
EOF

    for i in $(seq 0 "$LAST_CLIENT_INDEX"); do
    cat >> "$OUTPUT_FILE" <<EOF

  client_$i:
    build:
      context: ./services/client
      dockerfile: Dockerfile
    container_name: client_$i
    depends_on:
      - server
    environment:
      - AGENCY_ID=$i
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - INPUT_FILE=/input/input-$i.csv
      - OUTPUT_FILE=/output/output-$i.csv
      - BATCH_SIZE=2
    volumes: 
      - ./output:/output
      - ./input/input-$i.csv:/input/input-$i.csv

EOF
done

echo "Se genero $OUTPUT_FILE con $CLIENT_AMOUNT cliente(s) y quorum minimo $AGENCY_QUORUM_MIN."
