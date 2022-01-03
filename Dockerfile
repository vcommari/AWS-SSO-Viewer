FROM ubuntu:latest
COPY AWS-SSO-VIEWER /bin/
ADD staticfiles ./staticfiles
COPY config.yml /etc/aws-sso-viewer.yml

ENTRYPOINT ["AWS-SSO-VIEWER"]
