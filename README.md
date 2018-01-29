## defaults ports

```
server status api enpoint port 2000
```

```
# docker build
docker build . -t cadrspace_status

# docker run on server
docker run -d --restart=always -p 2000:2000 --name=cadrspace_status_server cadrspace_status /cadrspace_status_server

# build for Raspberry Pi
GOOS=linux GOARCH=arm GOARM=6 go build -o ./cadrspace_status_client -i ./client/main.go

# go run
go run ./client/main.go
go run ./server/main.go
```