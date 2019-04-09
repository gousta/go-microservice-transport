FROM golang:latest

ENV path /go/src/app

# Setup timezone
RUN ln -sf /usr/share/zoneinfo/Europe/Athens /etc/localtime

# Install govendor
RUN go get -u github.com/kardianos/govendor

# Add current path to the container, switch to it and install dependencies
ADD . $path
WORKDIR $path
RUN govendor sync

# Build to generate executable
RUN go get github.com/pilu/fresh

# Instruct container to 
# ENTRYPOINT $path/microservice-sms
RUN fresh

EXPOSE 9002