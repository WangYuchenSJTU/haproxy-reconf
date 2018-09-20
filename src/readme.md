
## Setup
```bash
go get github.com/gin-gonic/gin
go build restserver.go
sudo env GOPATH=/home/lucas/gocode go run restserver.go

```

## Access
```bash
curl -H "Content-Type:application/x-www-form-urlencoded" -X POST -d 'table=http-in&ip=192.168.3.54&role=0' http://localhost:8080/table 
```

## Backend
```bash
docker run -it --name httpd1 -v /home/lucas/webdata/data1:/usr/local/apache2/htdocs httpd
docker run -it --name httpd2 -v /home/lucas/webdata/data2:/usr/local/apache2/htdocs httpd
service haproxy start
```
