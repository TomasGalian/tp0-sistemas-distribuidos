#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"

touch $1

cat <<EOF > $1
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    volumes: 
      - ./server/config.ini:/config.ini
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
    networks:
      - testing_net

EOF

for i in $(seq 1 $2) 
do

  IFS=',' read -r nombre apellido documento nacimiento numero <<< "${bets[$i-1]}"

  cat <<EOF >> $1
  client$i:
    container_name: client$i
    image: client:latest
    volumes:
      - ./client/config.yaml:/config.yaml
      - ./.data/agency-$i.csv:/agency.csv
    entrypoint: /client
    environment:
      - CLI_ID=$i
    networks:
      - testing_net
    depends_on:
      - server

EOF
done

cat <<EOF >> $1
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOF