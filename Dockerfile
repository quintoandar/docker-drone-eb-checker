FROM golang:1.7.3
WORKDIR /go/src/github.com/quintoandar/drone-eb-checker
ADD . .
RUN go get ./... 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o drone-eb-checker .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /usr/local/bin
COPY --from=0 /go/src/github.com/quintoandar/drone-eb-checker/drone-eb-checker .
ENTRYPOINT ["drone-eb-checker"]  
