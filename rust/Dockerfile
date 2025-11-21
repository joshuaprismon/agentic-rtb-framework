
FROM rust:slim
WORKDIR /usr/src/chalice_hages_service
ENV BASE_DIR=/usr/src/chalice_hages_service
ENV DATA_DIR=${BASE_DIR}/data

RUN mkdir -p ${DATA_DIR}/models
RUN mkdir -p ${DATA_DIR}/logs
RUN mkdir -p ${BASE_DIR}/proto

COPY src/*.rs ./src/
COPY proto/*.proto ${BASE_DIR}/proto/
COPY src/models_and_segments/*.json ${DATA_DIR}/models/
COPY src/models_and_segments/*.yaml ${DATA_DIR}/models/
COPY cargo.toml ./Cargo.toml

RUN cargo install --path .
RUN apt-get update && apt-get install -y protobuf-compiler

ENV OUT_DIR=${BASE_DIR}/proto
ENV CHALICE_MODEL_CONFIG_FILE=hages_model_config.yaml
ENV CHALICE_MODEL_FILE_PATH=${DATA_DIR}/models/
ENV CHALICE_LOG_FILE_PATH=${DATA_DIR}/logs/
ENV CHALICE_METRICS_LOG_FILENAME=chalice_hages_metrics.log
ENV CHALICE_SERVER_PORT=22001
ENV CHALICE_MAX_CONNS=256
ENV CHALICE_KEEPALIVE_TIMEOUT=300

EXPOSE ${CHALICE_SERVER_PORT}

ENTRYPOINT ["chalice_hages_service"]