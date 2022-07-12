# SPDX-License-Identifier: Apache-2.0
# Copyright (c) Arm Limited and Contributors
# Copyright (c) Intel Corporation
FROM fedora:36

ARG TAG=v22.05
ARG ARCH=native

WORKDIR /root
RUN dnf install -y git diffutils procps-ng && dnf clean all

# hadolint ignore=DL3003
RUN git clone https://github.com/spdk/spdk --branch ${TAG} --depth 1 && \
    cd spdk && git submodule update --init --depth 1 && scripts/pkgdep.sh --rdma

# hadolint ignore=DL3003
RUN cd spdk && \
    ./configure --disable-tests --without-vhost --without-virtio \
                --with-rdma --target-arch=${ARCH} && \
    make
