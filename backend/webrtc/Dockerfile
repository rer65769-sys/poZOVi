FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
ENV LANG=C.UTF-8

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    curl \
    wget \
    ca-certificates \
    ninja-build \
    automake \
    autoconf \
    libtool \
    tar \
    unzip \
    libssl-dev \
    libsrtp2-dev \
    libuv1-dev \
    libglib2.0-dev \
    libjsoncpp-dev \
    pkg-config \
    zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

# Ensure /usr/local/bin is on PATH
ENV PATH=/usr/local/bin:$PATH
ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:$PKG_CONFIG_PATH

WORKDIR /tmp

# Build CMake 4.2.1 from source
RUN wget https://github.com/Kitware/CMake/releases/download/v4.2.1/cmake-4.2.1.tar.gz && \
    tar -xzf cmake-4.2.1.tar.gz && \
    cd cmake-4.2.1 && \
    ./bootstrap \
        --prefix=/usr/local \
        --parallel=$(nproc) && \
    make -j$(nproc) && \
    make install && \
    cd / && \
    rm -rf /tmp/cmake-4.2.1*

# Build gRPC
RUN git clone --recursive https://github.com/grpc/grpc && \
    cd grpc && \
    mkdir -p build && cd build && \
    cmake .. \
      -DgRPC_INSTALL=ON \
      -DgRPC_BUILD_TESTS=OFF \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_INSTALL_PREFIX=/usr/local && \
    cmake --build . --parallel $(nproc) && \
    cmake --install . && \
    cd /tmp && rm -rf grpc

# Build libdatachannel
RUN git clone --recursive https://github.com/paullouisageneau/libdatachannel.git && \
    cd libdatachannel && \
    mkdir -p build && cd build && \
    cmake .. \
      -DUSE_NICE=OFF \
      -DUSE_SYSTEM_SRTP=ON \
      -DCMAKE_CXX_STANDARD=23 \
      -DCMAKE_INSTALL_PREFIX=/usr/local \
      -DCMAKE_BUILD_TYPE=Release && \
    cmake --build . --parallel $(nproc) && \
    cmake --install . && \
    cd /tmp && rm -rf libdatachannel

WORKDIR /workspace
CMD ["bash"]