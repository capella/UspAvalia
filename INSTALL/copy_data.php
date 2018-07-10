<?php

$json = ["ok"];

$options = array(
    "ssl" => array(
        "verify_peer" => false,
        "verify_peer_name" => false,
    ),
);  


try {
	$data = file_get_contents("http://bcc.ime.usp.br/matrusp/db/db_usp.txt", false, stream_context_create($options));
	file_put_contents("db_usp.txt", $data);
} catch (Exception $e) {
    $json = array('error' =>  $e->getMessage());
}

echo json_encode($json, JSON_UNESCAPED_UNICODE);
?>