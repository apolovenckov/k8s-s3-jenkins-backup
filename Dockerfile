# Set the base image
FROM alpine:3.6

# Install dependencies
RUN apk -v --update add \
        python \
        py-pip \
        groff \
        less \
        mailcap \
        curl \
        && \
    pip install --upgrade awscli s3cmd python-magic && \
    apk -v --purge del py-pip && \
    rm /var/cache/apk/*

# Install kubectl
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl

ENV JENKINS_LABEL='app.kubernetes.io/name=jenkins'
ENV JENKINS_NAMESPACE='jenkins'
ENV JENKINS_HOME='/var/jenkins_home'

# Copy backup script and execute
COPY resources/jenkins-backup.sh /
RUN chmod +x /jenkins-backup.sh
CMD ["sh", "/jenkins-backup.sh"]
