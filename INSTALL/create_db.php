<?php

require __DIR__ . '/../vendor/autoload.php';
require __DIR__ . '/../helpers/connection.php';
require __DIR__ . '/../helpers/sanitizer.php';

$json = ["ok"];
$templine = '';

// Read in entire file
$lines = file("struct.sql");

// Loop through each line
foreach ($lines as $line) {
    // Skip it if it's a comment
    if (substr($line, 0, 2) == '--' || $line == '') {
        continue;
    }
    // Add this line to the current segment
    $templine .= $line;
    // If it has a semicolon at the end, it's the end of the query
    if (substr(trim($line), -1, 1) == ';') {
        // Perform the query
        $result = $connection->query($templine);
        if (!$result) {
            $json = array('error' => $connection->error);
            break;
        }
        // Reset temp variable to empty
        $templine = '';
    }
}
echo json_encode($json, JSON_UNESCAPED_UNICODE);
