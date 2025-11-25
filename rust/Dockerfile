
FROM rust:slim
WORKDIR /usr/src/agenticrtbframework
ENV BASE_DIR=/usr/src/agenticrtbframework
ENV DATA_DIR=${BASE_DIR}/data

RUN mkdir -p ${BASE_DIR}/proto

COPY proto/*.proto ${BASE_DIR}/proto/
COPY cargo.toml ./Cargo.toml

RUN cargo install --path .
RUN apt-get update && apt-get install -y protobuf-compiler

ENV OUT_DIR=${BASE_DIR}/proto
ENV ARTF_SERVER_PORT=22001
ENV ARTF_MAX_CONNS=256

EXPOSE ${ARTF_SERVER_PORT}

ENTRYPOINT ["agentic-rtb-framework-service"]