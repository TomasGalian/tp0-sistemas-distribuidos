#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"

bets=(
  "Juan,Perez,12345678,1990-01-01,1234"
  "Maria,Gomez,87654321,1985-05-10,5678"
  "Carlos,Lopez,56781234,1992-08-20,9012"
  "Ana,Martinez,34567812,1988-12-15,3456"
  "Luis,Diaz,67890123,1995-06-30,7890"
)

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
    entrypoint: /client
    environment:
      - CLI_ID=$i
      - NOMBRE=$nombre
      - APELLIDO=$apellido
      - DOCUMENTO=$documento
      - NACIMIENTO=$nacimiento
      - NUMERO=$numero
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