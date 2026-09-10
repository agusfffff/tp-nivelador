#!/bin/bash

NETWORK=$(docker network ls --format '{{.Name}}' | grep -m1 '^tp-nivelador_default$')
if [ -z "$NETWORK" ]; then
    NETWORK=$(docker network ls --format '{{.Name}}' | head -n1)
fi

echo "Usando red: $NETWORK"

MSG="Works well"

RSP=$(echo "$MSG" | docker run --rm -i --network "$NETWORK" busybox nc -w2 server 5678)

if [ "$RSP" == "$MSG" ]; then
    echo "Works well"
    exit 0
else 
    echo "Doesnt work well" 
    echo "enviado: '$MSG'" 
    echo "recibido: '$RSP'" 

    exit 1
fi