FROM golang:1.21.7
 
WORKDIR /monitor
COPY . .
RUN go mod download
 
RUN go build ./cmd/monitor
 
EXPOSE 8989
 
ENTRYPOINT [ "./monitor" ]
