FROM golang:1.18
ENV GOPRIVATE=github.com/edgelesssys/*

RUN apt-get update && apt-get install cmake iproute2 iputils-ping libssl-dev -y

RUN git clone -b v1.0.20210914 --depth=1 https://git.zx2c4.com/wireguard-tools && make -C wireguard-tools/src -j`nproc` && make -C wireguard-tools/src install
RUN git clone -b 0.0.20211016 --depth=1 https://git.zx2c4.com/wireguard-go && cd wireguard-go && make -j`nproc` && make install

COPY . /constellation
RUN cd /constellation && rm -rf build && mkdir build && cd build && cmake .. &&  make debug_coordinator -j`nproc` && cd /

CMD wireguard-go wg0 && /constellation/build/debug_coordinator -debug
