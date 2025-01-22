FROM golang:1.23.5

LABEL version="v0.0.1"

COPY . .
RUN go build

EXPOSE 9101

ENTRYPOINT ["./ses_exporter"]