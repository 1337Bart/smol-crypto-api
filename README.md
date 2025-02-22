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
- następny endpoint: do danych historycznych z datą jako parametr
- telemetry (porównanie szybkości redis vs postgres)
- dodatkowe endpointy z fajnymi metryczkami (Np. zmiana w ciągu ostatniego tygodnia, rsi, macd itp)