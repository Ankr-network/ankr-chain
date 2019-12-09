# Compile
FROM golang:1.13-alpine AS compiler
ARG PRIV_PATH

RUN apk add --no-cache git make dep openssh-client

COPY id_rsa /root/.ssh/
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN chmod go-w /root
RUN chmod 700 /root/.ssh
RUN chmod 600 /root/.ssh/id_rsa

WORKDIR /ankr-chain
COPY . .

RUN make linux

# Build image
FROM alpine:3.7 AS public

# Ankrchain will be looking for the genesis file in /ankrchain/config/genesis.json
# (unless you change `genesis_file` in config.toml). You can put your config.toml and
# private validator file into /ankrchain/config.
#
# The /ankrchain/data dir is used by ankrchain to store state.
ENV TMHOME /ankrchain

# OS environment setup
# Set user right away for determinism, create directory for persistence and give our user ownership
# jq and curl used for extracting `pub_key` from private validator while
# deploying ankrchain with Kubernetes. It is nice to have bash so the users
# could execute bash commands.
RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq bash && \
    addgroup tmuser && \
    adduser -S -G tmuser tmuser -h "$TMHOME"

ARG BRANCH_NAME
ENV BRANCH_NAME=${BRANCH_NAME}
COPY --from=compiler /ankr-chain/build/ankrchain-linux-amd64/ankrchain /usr/bin/ankrchain
COPY DOCKER/tmhome/config/config."${BRANCH_NAME}".toml "$TMHOME"/config/config.toml
COPY DOCKER/tmhome/config/genesis."${BRANCH_NAME}".json "$TMHOME"/config/genesis.json
RUN mkdir "$TMHOME"/data \
    && chown -R tmuser:tmuser "$TMHOME"/config "$TMHOME"/data

# Run the container with tmuser by default. (UID=100, GID=1000)
USER root

# Expose the data directory as a volume since there's mutable state in there
VOLUME [ $TMHOME ]

WORKDIR $TMHOME

# p2p and rpc port
EXPOSE 26656 26657 26658

ENTRYPOINT ["/bin/sh"]
CMD ["-c", "ankrchain init && ankrchain start --log_level=info --moniker=`hostname`"]
STOPSIGNAL SIGTERM

FROM public AS hub
USER root

COPY DOCKER/tmhome/config/node_key.*.json "$TMHOME"/config/
COPY DOCKER/tmhome/config/priv_validator_key.*.json "$TMHOME"/config/
COPY DOCKER/tmhome/config/priv_validator_state.json "$TMHOME"/config/
RUN chown -R tmuser:tmuser "$TMHOME"/config "$TMHOME"/config

USER root
CMD ["ankrchain","-h"]