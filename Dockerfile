FROM alpine:3.20.0

RUN apk update && apk add --no-cache \
    openssh=9.7_p1-r3 \
    rclone=1.66.0-r3 \
    python3=3.12.3-r1

COPY sshd_config /etc/ssh/sshd_config
COPY entrypoint.py /entrypoint.py
COPY sync.py /sync.py
RUN chmod +x /entrypoint.py /sync.py

EXPOSE 22
CMD ["/entrypoint.py"]
