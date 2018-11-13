FROM alpine:3.6
LABEL AUTHOR <shadowbq@gmail.com>

# tested with v0.1.0 release - hardcoded -pull precompiled binares OR compile in dockerfile (< image)
ARG VERSION

RUN addgroup -g 9999 nomnomlog
RUN adduser -u 9999 -D -G nomnomlog -k /etc/nomnomlog -H nomnomlog

RUN apk update \
  && apk add ca-certificates wget \
  && update-ca-certificates 

##build arg used to pull desired versions - docker cli 
RUN wget https://github.com/shadowbq/nomnomlog/releases/download/$VERSION/nomnomlog_linux_amd64.tar.gz
RUN tar -xvf nomnomlog_linux_amd64.tar.gz

#package/tarball path changes will break dockerfile at this point
RUN mv nomnomlog/nomnomlog /usr/local/bin/nomnomlog
RUN mv /nomnomlog/example_config.yml /etc/nomnomlog-config.yml

#Define the directories for -v flag
#VOLUME ["/var/log/foobar/"]
RUN touch /locallog.txt
RUN chown nomnomlog:nomnomlog /locallog.txt

#run nomnomlog as user in production
RUN chown nomnomlog:nomnomlog /usr/local/bin/nomnomlog
USER nomnomlog
CMD ["nomnomlog", "-D" ]

## uncomment to debug container
#CMD ["/usr/bin/tail", "-f", "/tmp/foo"]
