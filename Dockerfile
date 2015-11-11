# Docker image running maud (https://github.com/hamcha/maud)
# /// Howto: ///
#
# -- Non-dockerized DB --
# docker run -d --net=host silverweed/maud:latest -dburl <db_url>
#
# -- Dockerized DB --
# docker run -d --name db mongo:latest
# docker run -d --link db:db silverweed/maud:latest -dburl db
#
# To change conf files:
# mkdir my_maud_conf && cd my_maud_conf
# (create & edit conf files)
# Create a Dockerfile:
# 	FROM silverweed/maud:latest
#	COPY info.json admin.conf blacklist.conf captcha.conf /maud/
#	ENTRYPOINT [ "./maud" ]
# docker build -t my_maud_conf .
# (launch as descripted above)
FROM debian:jessie
MAINTAINER silverweed <silverweed1991@gmail.com>
EXPOSE 8080
RUN mkdir -p /maud/template /maud/errors /maud/static /maud/stiki
COPY maud /maud/
COPY template /maud/template
COPY errors /maud/errors
COPY static /maud/static
COPY stiki /maud/stiki
COPY info.json admin.conf blacklist.conf captcha.conf /maud/
WORKDIR /maud/
ENTRYPOINT [ "./maud" ]
