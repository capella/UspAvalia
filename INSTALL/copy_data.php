<?php

$json = ["ok"];
$templine = '';

$data = fopen("http://bcc.ime.usp.br/matrusp/db/db_usp.txt", 'r');
file_put_contents("db_usp.txt", $data);

echo json_encode($json, JSON_UNESCAPED_UNICODE);
?>