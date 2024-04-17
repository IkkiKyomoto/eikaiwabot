FROM golang:1.18-alpine3.15 AS go
WORKDIR /eikaiwabot
COPY go.mod go.sum main.go ./
CMD ["/eikaiwabot/main"]
EXPOSE 3000
