FROM golang:1.9.2-alpine
RUN apk add --update git
WORKDIR /go/src/cadrspace_status
COPY . $WORKDIR
RUN go get -u github.com/golang/dep/...
RUN dep ensure
RUN go build -o ./cadrspace_status_client -i ./client/main.go
RUN go build -o ./cadrspace_status_server -i ./server/main.go

FROM alpine:latest  
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /
COPY --from=0 /go/src/cadrspace_status/cadrspace_status_client .
COPY --from=0 /go/src/cadrspace_status/cadrspace_status_server .