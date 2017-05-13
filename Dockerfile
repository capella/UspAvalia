FROM php:5.6-apache

LABEL maintainer="gabriel@capella.pro"

# Packages
RUN apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -y curl zlib1g-dev git

RUN docker-php-ext-install mysql
RUN cp /etc/apache2/mods-available/rewrite.load /etc/apache2/mods-enabled/rewrite.load

# Setup the Composer installer
RUN curl -o /tmp/composer-setup.php https://getcomposer.org/installer \
  && curl -o /tmp/composer-setup.sig https://composer.github.io/installer.sig \
  && php -r "if (hash('SHA384', file_get_contents('/tmp/composer-setup.php')) !== trim(file_get_contents('/tmp/composer-setup.sig'))) { unlink('/tmp/composer-setup.php'); echo 'Invalid installer' . PHP_EOL; exit(1); }"

RUN php /tmp/composer-setup.php --no-ansi --install-dir=/usr/local/bin --filename=composer && rm -rf /tmp/composer-setup.php
COPY composer.json /var/www/html/
WORKDIR /var/www/html/
RUN php /usr/local/bin/composer install  --no-dev
WORKDIR /

ADD . /var/www/html/
