FROM ubuntu:22.04@sha256:27cb6e6ccef575a4698b66f5de06c7ecd61589132d5a91d098f7f3f9285415a9 AS build
RUN apt-get update && apt-get install -y \
  build-essential \
  libcurl4-openssl-dev \
  wget
RUN wget -q https://packages.microsoft.com/repos/azurecore/pool/main/a/azguestattestation1/azguestattestation1_1.0.2_amd64.deb \
  && apt-get install /azguestattestation1_1.0.2_amd64.deb
RUN wget -q https://github.com/Azure/confidential-computing-cvm-guest-attestation/raw/4bd89d2808912fbaa319e8853e6f5e1e245d45ca/cvm-guest-attestation-linux-app/main.cpp \
  && sed -i s/test.attest.azure.net/attest.azure.net/ main.cpp \
  && touch Utils.h \
  && g++ -Os -I/usr/include/azguestattestation1 -oclient main.cpp -lazguestattestation

FROM ubuntu:22.04@sha256:27cb6e6ccef575a4698b66f5de06c7ecd61589132d5a91d098f7f3f9285415a9 AS release
COPY --from=build client azguestattestation1_1.0.2_amd64.deb /
RUN apt-get update && apt-get install -y /azguestattestation1_1.0.2_amd64.deb
ENTRYPOINT ["/client"]
