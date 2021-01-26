FROM busybox

COPY bin/prometheus-zenaiop /bin/prometheus-zenaiop
RUN mkdir -p /prometheus-zenaiop

USER        nobody
ENV         GIN_MODE=release
EXPOSE      9299
WORKDIR     /prometheus-zenaiop
ENTRYPOINT [ "/bin/prometheus-zenaiop" ]
CMD        [ "" ]