FROM golang:1.18
ENV GOPRIVATE=github.com/edgelesssys/*

RUN apt-get update && apt-get install  pkg-config cmake iproute2 iputils-ping libssl-dev libcryptsetup12 libcryptsetup-dev -y

WORKDIR /constellation
COPY go.mod .
COPY go.sum .
RUN go mod download

RUN git clone -b v1.0.20210914 --depth=1 https://git.zx2c4.com/wireguard-tools && make -C wireguard-tools/src -j`nproc` && make -C wireguard-tools/src install
RUN git clone -b 0.0.20211016 --depth=1 https://git.zx2c4.com/wireguard-go && cd wireguard-go && make -j`nproc` && make install

COPY . /constellation
RUN rm -rf build && mkdir build && cd build && cmake .. &&  make debug_bootstrapper -j`nproc` && cd /

CMD wireguard-go wg0 && /constellation/build/debug_bootstrapper -debug
