# Dockerfile - alpine
# https://github.com/openresty/docker-openresty

FROM alpine:latest

MAINTAINER Evan Wies <evan@neomantra.net>

# Docker Build Arguments
ARG RESTY_VERSION="1.11.2.5"
ARG RESTY_OPENSSL_VERSION="1.0.2j"
ARG RESTY_PCRE_VERSION="8.39"
ARG RESTY_J="1"
ARG LUAROCKS_VERSION="2.4.2"
ARG RESTY_CONFIG_OPTIONS="\
    --with-file-aio \
    --with-http_addition_module \
    --with-http_auth_request_module \
    --with-http_dav_module \
    --with-http_flv_module \
    --with-http_geoip_module=dynamic \
    --with-http_gunzip_module \
    --with-http_gzip_static_module \
    --with-http_image_filter_module=dynamic \
    --with-http_mp4_module \
    --with-http_random_index_module \
    --with-http_realip_module \
    --with-http_secure_link_module \
    --with-http_slice_module \
    --with-http_ssl_module \
    --with-http_stub_status_module \
    --with-http_sub_module \
    --with-http_v2_module \
    --with-http_xslt_module=dynamic \
    --with-ipv6 \
    --with-mail \
    --with-mail_ssl_module \
    --with-md5-asm \
    --with-pcre-jit \
    --with-sha1-asm \
    --with-stream \
    --with-stream_ssl_module \
    --with-threads \
    "

# These are not intended to be user-specified
ARG _RESTY_CONFIG_DEPS="--with-openssl=/tmp/openssl-${RESTY_OPENSSL_VERSION} --with-pcre=/tmp/pcre-${RESTY_PCRE_VERSION} --with-pcre-jit"


# 1) Install apk dependencies
# 2) Download and untar OpenSSL, PCRE, and OpenResty
# 3) Build OpenResty
# 4) Cleanup

RUN \
    apk add --no-cache --virtual .build-deps \
        build-base \
        curl \
        gd-dev \
        geoip-dev \
        libxslt-dev \
        linux-headers \
        make \
        perl-dev \
        readline-dev \
        zlib-dev \
        pcre-dev \
        g++ \
        make \
        python \
        bash
RUN \ 
        apk add --no-cache \
        supervisor \
        gd \
        geoip \
        libgcc \
        libxslt \
        zlib \
        curl \
        pcre \
        libstdc++
RUN \
    cd /tmp \
    && curl -k -fSL https://www.openssl.org/source/openssl-${RESTY_OPENSSL_VERSION}.tar.gz -o openssl-${RESTY_OPENSSL_VERSION}.tar.gz \
    && tar xzf openssl-${RESTY_OPENSSL_VERSION}.tar.gz \
    && curl -k -fSL https://ftp.csx.cam.ac.uk/pub/software/programming/pcre/pcre-${RESTY_PCRE_VERSION}.tar.gz -o pcre-${RESTY_PCRE_VERSION}.tar.gz \
    && tar xzf pcre-${RESTY_PCRE_VERSION}.tar.gz \
    && curl -k -fSL https://openresty.org/download/openresty-${RESTY_VERSION}.tar.gz -o openresty-${RESTY_VERSION}.tar.gz \
    && tar xzf openresty-${RESTY_VERSION}.tar.gz \
    && cd /tmp/openresty-${RESTY_VERSION} \
    && ./configure -j${RESTY_J} ${_RESTY_CONFIG_DEPS} ${RESTY_CONFIG_OPTIONS} \
    && make -j${RESTY_J} \
    && make -j${RESTY_J} install

# Install luarocks
RUN \
    cd /tmp \
    && curl -k -fSL http://luarocks.github.io/luarocks/releases/luarocks-${LUAROCKS_VERSION}.tar.gz -o luarocks-${LUAROCKS_VERSION}.tar.gz \
    && tar xzf luarocks-${LUAROCKS_VERSION}.tar.gz \
    && cd luarocks-${LUAROCKS_VERSION} \
    && ./configure --prefix=/usr/local/openresty/luajit \
    --with-lua=/usr/local/openresty/luajit/ \
    --lua-suffix=jit \
    --with-lua-include=/usr/local/openresty/luajit/include/luajit-2.1 \
    && make -j${RESTY_J} \
    && make install

# Install lua-resty-waf
RUN \
    mkdir /tmp/lua-resty-waf

ADD addition_package/lua-resty-waf /tmp/lua-resty-waf
    
## Prepare for install lua-resty-waf
RUN \
    cd /tmp/lua-resty-waf/libinjection \
    && make all

RUN \
    cp /usr/local/openresty/luajit/bin/luajit /usr/local/openresty/luajit/bin/lua

RUN \
    cd /tmp \
    && cd lua-resty-waf \
    && bash \
    && export PATH=/usr/local/openresty/luajit/bin:/usr/local/openresty/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    && make LUA_INCLUDE_DIR=/usr/local/openresty/luajit/include/luajit-2.1 \
    && make LUA_INCLUDE_DIR=/usr/local/openresty/luajit/include/luajit-2.1/ install

RUN \
    cd /tmp \
    && rm -rf \
        openssl-${RESTY_OPENSSL_VERSION} \
        openssl-${RESTY_OPENSSL_VERSION}.tar.gz \
        openresty-${RESTY_VERSION}.tar.gz openresty-${RESTY_VERSION} \
        pcre-${RESTY_PCRE_VERSION}.tar.gz pcre-${RESTY_PCRE_VERSION} \
        luarocks-${LUAROCKS_VERSION}.tar.gz luarocks-${LUAROCKS_VERSION} \
        lua-resty-waf \
    && apk del .build-deps \
    && ln -sf /dev/stdout /usr/local/openresty/nginx/logs/access.log \
    && ln -sf /dev/stderr /usr/local/openresty/nginx/logs/error.log

RUN mkdir -p /etc/nginx/certs
COPY certs/default.pem /etc/nginx/certs/

COPY nginx/lualib/prometheus.lua /usr/local/openresty/lualib/
COPY nginx/nginx_config/nginx.conf /etc/nginx/

COPY postStart.sh /
RUN chmod +x /postStart.sh

## nginx-config-watcher binary
COPY nginx-config-watcher /
RUN chmod +x /nginx-config-watcher

## nginx.tmpl
RUN mkdir -p /etc/nginx/template/
COPY nginx/template/nginx.tmpl /etc/nginx/template/nginx.tmpl

## supervisord config
COPY supervisord/supervisord.conf /etc/supervisord.conf

ENTRYPOINT ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]