## defaults ports

```
server status api enpoint port   2000
server tunnel port               3000
server shared port               4000
local forwarded port             5000
```

```
# docker build
docker build . -t cadrspace_status

# docker run
docker run -d --restart=always --network=host --name=cadrspace_status_client cadrspace_status /cadrspace_status_client

docker run -d --restart=always -p 2000:2000 -p 3000:3000 -p 4000:4000 --name=cadrspace_status_server cadrspace_status /cadrspace_status_server

# build for Raspberry Pi 2 Model B
GOOS=linux GOARCH=arm GOARM=7 go build -o ./cadrspace_status_client -i ./client/main.go
GOOS=linux GOARCH=arm GOARM=7 go build -o ./cadrspace_status_server -i ./server/main.go

# go run
go run ./client/main.go
go run ./server/main.go
```