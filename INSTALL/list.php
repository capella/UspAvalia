<?php

require __DIR__ . '/../vendor/autoload.php';
require __DIR__ . '/../helpers/connection.php';
require __DIR__ . '/../helpers/sanitizer.php';

$json = [];
$sql = "SELECT * FROM disciplinas";

$result = $connection->query($sql);

while($row = $result->fetch_assoc()) {
    $json[] = $row['id'];
}
$result->close();

echo json_encode($json, JSON_UNESCAPED_UNICODE);
