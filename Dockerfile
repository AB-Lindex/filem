# syntax=docker/dockerfile:1

##
## STEP 1 - BUILD
##

# specify the base image to  be used for the application, alpine or ubuntu
FROM golang:1.21-alpine AS build

# create a working directory inside the image
WORKDIR /app

# copy Go modules and dependencies to image
COPY go.* ./

# download Go modules and dependencies
RUN go mod download

# copy directory files i.e all files ending with .go
COPY ./src ./src

# compile application
#ENV CGO_ENABLED=0
RUN go build -o /filem ./src

##
## STEP 2 - DEPLOY
##
FROM alpine:latest

RUN apk upgrade --no-cache --force-refresh ca-certificates

WORKDIR /app

COPY --from=build /filem filem

ENTRYPOINT ["/app/filem"]