FROM golang:1.20

COPY armbian-release /etc/

# NB: golang image already has git.
#RUN apt-get update && \
#    apt-get install -y git && \

RUN apt-get update && apt-get install -y sudo

WORKDIR /app

RUN \
    git clone https://github.com/orangepi-xunlong/wiringOP.git && \
    cd wiringOP && \
    ./build clean && \
    ./build

COPY go.mod *.go ./

RUN go build -o /opi-fan

CMD ["/opi-fan"]

# Note that because we require access to /sys/devices and /dev/mem, we need to run this image with --privileged
