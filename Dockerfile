FROM ubuntu:latest

RUN apt-get update
RUN apt-get install -y ca-certificates
COPY AWS-SSO-VIEWER /bin/
ADD staticfiles ./staticfiles
COPY config.yml /etc/aws-sso-viewer.yml

ENTRYPOINT ["AWS-SSO-VIEWER"]
