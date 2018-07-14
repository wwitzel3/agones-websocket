FROM alpine:3.7

RUN adduser -D server    

COPY ./bin/server /home/server/server
COPY ./site/index.html /home/server/site/index.html

RUN chown -R server /home/server && \
    chmod o+x /home/server/server

USER server
ENTRYPOINT /home/server/server
