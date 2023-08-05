<?php

/* configiracoes do banco de dados */
$hostname = $_ENV["HOSTNAME"];
$database = $_ENV["MYSQL_DATABASE"];
$username = $_ENV["MYSQL_USER"];
$password = $_ENV["MYSQL_PASSWORD"];

/* chave de seguranca, usada para hash */
$secret_key = $_ENV["SECRET_KEY"];

/* tipo do hash */
/* http://php.net/manual/en/function.hash.php */
$hash = "sha256";

/* chaves facebook */
$appId_facebook = $_ENV["APPID_FACEBOOK"];
$secret_facebook = $_ENV["SECRET_FACEBOOK"];

/* sitename */
$url = "uspavalia.com";
$url_full = "https://uspavalia.com";


/* smtp mail */
$smtp_host = $_ENV["SMTP_HOST"];
$smtp_username = $_ENV["SMTP_USERNAME"];
$smtp_password = $_ENV["SMTP_PASSWORD"];
$smtp_password = $_ENV["SMTP_PASSWORD"];
