<?php

require __DIR__ . '/../config.php';
$hostname_connection = $hostname;
$database_connection = $database;
$username_connection = $username;
$password_connection = $password;
$connection = mysqli_connect($hostname_connection, $username_connection, $password_connection);

if (!$connection) {
    echo "Error: Unable to connect to MySQL." . PHP_EOL;
    echo "Debugging errno: " . mysqli_connect_errno() . PHP_EOL;
    echo "Debugging error: " . mysqli_connect_error() . PHP_EOL;
    exit;
}

if (!$connection->set_charset("utf8")) {
    printf("Error loading character set utf8: %s\n", $mysqli->error);
    exit;
}

$connection->select_db($database_connection);
