<?php

require __DIR__ . '/../config.php';
$hostname_connection = $hostname;
$database_connection = $database;
$username_connection = $username;
$password_connection = $password;
$connection = mysql_pconnect($hostname_connection, $username_connection, $password_connection) or trigger_error(mysql_error(),E_USER_ERROR);
mysql_set_charset('utf8',$connection);

?>