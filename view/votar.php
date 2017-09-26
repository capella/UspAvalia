<?php

if (!$user) {
  header('Location: '.$loginUrl);
  exit;
}

$arr = array('Avaliação Geral', 'Didática', 'Empenho/Dedicação', 'Relação com os alunos', 'Dificuldade');
reset($arr);
while (list($key, $value) = each($arr)) {
	$chave = $key+1;
	if ((isset($_POST["nota".$chave])) && ($_POST["nota".$chave] != "")) {
	$insertSQL = sprintf("INSERT INTO votos (APid, iduso, `time`, nota, tipo) VALUES (%s, %s, %s, %s, %s) ON DUPLICATE KEY UPDATE `time`=%s, nota=%s",
						   GetSQLValueString($_POST['id'], "int"),
						   GetSQLValueString(hash($hash , $secret_key.$user_profile['id']), "text"),
						   GetSQLValueString(time(), "int"),
						   GetSQLValueString($_POST['nota'.$chave], "int"),
						   GetSQLValueString($chave, "int"),
						   GetSQLValueString(time(), "int"),
						   GetSQLValueString($_POST['nota'.$chave], "int"));
	
	  mysql_select_db($database_connection, $connection);
	  $Result1 = mysql_query($insertSQL, $connection) or die(mysql_error());
	
	}
}

$insertGoTo = "?p=ver&id=".$_POST['id'];
header(sprintf("Location: %s", $insertGoTo));

?>