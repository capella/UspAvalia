<?php

if (!$user) {
    header('Location: '.$loginUrl);
    exit;
}

$arr = array('Avaliação Geral', 'Didática', 'Empenho/Dedicação', 'Relação com os alunos', 'Dificuldade');
reset($arr);
while (list($key, $value) = each($arr)) {
    $chave = $key + 1;
    if ((isset($_GET["nota".$chave])) && ($_GET["nota".$chave] != "")) {
        $insertSQL = sprintf(
            "INSERT INTO votos (APid, iduso, `time`, nota, tipo) VALUES (%s, %s, %s, %s, %s) ON DUPLICATE KEY UPDATE `time`=%s, nota=%s",
            GetSQLValueString($_GET['id'], "int"),
            GetSQLValueString(hash($hash, $secret_key.$user_profile['id']), "text"),
            GetSQLValueString(time(), "int"),
            GetSQLValueString($_GET['nota'.$chave], "int"),
            GetSQLValueString($chave, "int"),
            GetSQLValueString(time(), "int"),
            GetSQLValueString($_GET['nota'.$chave], "int")
        );

        $result = $connection->query($insertSQL);
    }
}

$insertGoTo = "?p=ver&id=".$_GET['id'];
header(sprintf("Location: %s", $insertGoTo));
