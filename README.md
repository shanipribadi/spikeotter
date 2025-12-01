```
GOEXPERIMENT=greenteagc go build -o so.gt ./cmd/so/...
go build ./cmd/so/... -o so

openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out crt.pem -days 365 \
 -subj "/CN=localhost.localdomain" \
 -addext "subjectAltName=DNS:localhost.localdomain,IP:127.0.0.1"

# Otter
GOGC=800 GOMEMLIMIT=8GiB GOMAXPROCS=2 taskset -c 0-1 ./so.gt -maxsize 60000 -unique 3000000 -expiry 30s -refresh 10s -loadfactor 25
echo "GET https://localhost.localdomain:8443/bcache" | GOMAXPROCS=2 taskset -c 2-3 vegeta attack -rate 1000/s -duration 120s -timeout 1s -root-certs crt.pem| vegeta report --every 1s

# BigCache
GOGC=800 GOMEMLIMIT=8GiB GOMAXPROCS=2 taskset -c 0-1 ./so.gt -maxsize 25000 -unique 3000000 -expiry 30s -refresh 1s -loadfactor 25
echo "GET https://localhost.localdomain:8443/bcache" | GOMAXPROCS=2 taskset -c 2-3 vegeta attack -rate 1000/s -duration 120s -timeout 1s -root-certs crt.pem| vegeta report --every 1s

# Source
GOGC=800 GOMEMLIMIT=8GiB GOMAXPROCS=2 taskset -c 0-1 ./so.gt -maxsize 25000 -unique 3000000 -expiry 30s -refresh 1s -loadfactor 25
echo "GET https://localhost.localdomain:8443/source" | GOMAXPROCS=2 taskset -c 2-3 vegeta attack -rate 1000/s -duration 120s -timeout 1s -root-certs crt.pem| vegeta report --every 1s
```
