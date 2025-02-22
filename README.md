# smol-crypto-api
grpc-rest-crypto-api


# Redis
redis-cli > 

KEYS * (list all keys)
GET keyname (get value for a specific key)
INFO (get Redis server information)

z docker-compose:
http://localhost:8001/


# Buf
buf mod prune
buf mod update
buf generate

# Curl
curl "http://localhost:8080/api/v1/crypto?page=2&limit=20" 


Do zrobienia jako kolejne:
- upewnic sie, ze postgres dziala gdy redis nie ma danych, albo gdy request jest o dane historyczne
- następne endpointy
- telemetry (porównanie szybkości redis vs postgres)
- frontend dla http, moze też dla grpc?? research
- dodatkowe endpointy z fajnymi metryczkami 