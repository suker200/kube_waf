FROM python:2.7-alpine3.7

RUN apk update && \
		apk add libffi-dev \
				musl-dev \
				openssl-dev 

RUN mkdir /app

ADD requirements.txt /

RUN pip install -r /requirements.txt

ADD . /app

WORKDIR /app

RUN cp -rpf /app/elastalert_modules /usr/local/lib/python2.7/

ENTRYPOINT ["python", "-m", "elastalert.elastalert", "--verbose"]


