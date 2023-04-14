FROM golang:1.20 as builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn make build

FROM debian

RUN apt-get update && apt-get install -y git curl && rm -rf /var/lib/apt/lists/ && apt-get autoremove -y && apt-get autoclean -y

RUN curl -sL https://deb.nodesource.com/setup_16.x | bash - && apt-get install -y nodejs && npm install -g yarn

RUN mkdir -p /app/configs
RUN mkdir -p /root/yoo/projects

COPY deployments/ssh_key/id_ed25519 /root/.ssh/id_ed25519

RUN chmod 600 /root/.ssh/id_ed25519

RUN touch /root/.ssh/config \
    && echo "StrictHostKeyChecking no" >> /root/.ssh/config \
    && echo "IdentityFile /root/.ssh/id_ed25519" >> /root/.ssh/config

COPY --from=builder /src/_output/yoo /app
COPY --from=builder /src/configs/yoo.yaml /app/configs

WORKDIR /app

EXPOSE 8080

VOLUME ["/app/configs"]

CMD ["./yoo", "-c", "/app/configs/yoo.yaml"]