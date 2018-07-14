FROM alpine:3.7

RUN adduser -D launcher

COPY ./bin/launcher /home/launcher/launcher 

RUN chown -R launcher /home/launcher && \
    chmod o+x /home/launcher/launcher

ENV PORT 8080
USER launcher
ENTRYPOINT /home/launcher/launcher
