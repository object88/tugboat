FROM python:3.9-slim-buster

RUN apt-get update && \
  apt-get install -y --no-install-recommends graphviz

RUN pip install diagrams

ADD build.sh /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/build.sh"]
CMD [""]