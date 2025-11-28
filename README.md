```
GOEXPERIMENT=greenteagc go build -o so.gt ./cmd/so/...
go build ./cmd/so/... -o so

openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out crt.pem -days 365 \
 -subj "/CN=localhost.localdomain" \
 -addext "subjectAltName=DNS:localhost.localdomain,IP:127.0.1.1"

GOGC=800 GOMEMLIMIT=4GiB GOMAXPROCS=4 taskset -c 0-3 ./so.gt
echo "GET https://localhost.localdomain:8443/cache" | GOMAXPROCS=4 taskset -c 4-7 vegeta attack -rate 2000/s -duration 120s -timeout 1s -root-certs crt.pem| vegeta report
echo "GET https://localhost.localdomain:8443/source" | GOMAXPROCS=4 taskset -c 4-7 vegeta attack -rate 2000/s -duration 120s -timeout 1s -root-certs crt.pem| vegeta report
echo "GET https://localhost.localdomain:8443/cache" | GOMAXPROCS=4 taskset -c 4-7 vegeta attack -rate 2000/s -duration 120s -timeout 1s -root-certs crt.pem| vegeta report

GOGC=800 GOMEMLIMIT=4GiB GOMAXPROCS=4 taskset -c 0-3 ./so
echo "GET https://localhost.localdomain:8443/cache" | GOMAXPROCS=4 taskset -c 4-7 vegeta attack -rate 2000/s -duration 120s -timeout 1s -root-certs crt.pem| vegeta report
echo "GET https://localhost.localdomain:8443/source" | GOMAXPROCS=4 taskset -c 4-7 vegeta attack -rate 2000/s -duration 120s -timeout 1s -root-certs crt.pem| vegeta report
echo "GET https://localhost.localdomain:8443/cache" | GOMAXPROCS=4 taskset -c 4-7 vegeta attack -rate 2000/s -duration 120s -timeout 1s -root-certs crt.pem| vegeta report


 GOGC=800 GOMEMLIMIT=4GiB GOMAXPROCS=4 taskset -c 0-3 ./so.gt -maxsize 10000 -unique 3000000 -expiry 0s -refresh 1000s -loadfactor 25
 echo "GET https://localhost.localdomain:8443/cache" | GOMAXPROCS=4 taskset -c 4-7 vegeta attack -rate 2000/s -duration 120s -timeout 1s -root-certs crt.pem| vegeta report --every 1s
```
