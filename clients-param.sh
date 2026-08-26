#!/bin/bash

set -euo pipefail

if [ $# -ne 1 ]; then
    echo "Uso: $0 <cantidad_de_clientes>" >&2
    exit 1
fi

CLIENT_AMOUNT=$1

if ! [[ "$CLIENT_AMOUNT" =~ ^[0-9]+$ ]] || [ "$CLIENT_AMOUNT" -le 0 ]; then
    echo "Error: la cantidad de clientes debe ser un numero entero positivo" >&2
    exit 1
fi

OUTPUT_FILE="docker-compose.yaml"

cat > "$OUTPUT_FILE" <<EOF
services:
  server:
    build:
      context: ./services/server
      dockerfile: Dockerfile
    container_name: server
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_HOST=server
      - SERVER_PORT=5678
EOF

for i in $(seq 1 "$CLIENT_AMOUNT"); do
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
EOF
done

echo "Se genero $OUTPUT_FILE con $CLIENT_AMOUNT cliente(s)."
