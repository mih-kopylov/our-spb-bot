FROM alpine

WORKDIR /bot

COPY ./dist/our-spb-bot_linux_amd64_v1/bot ./bot

ENTRYPOINT ["sh"]