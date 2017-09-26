<?
if (!$user) {
  header('Location: '.$loginUrl);
  exit;
}

if ((isset($_GET["MM_insert"])) && ($_GET["MM_insert"] == "form")) {

  $insertSQL = sprintf("INSERT INTO cometario (comantario, iduso, aulaprofessorid, time) VALUES (%s,%s,%s,%s)",
             GetSQLValueString($_GET['comentario'], "text"),
					   GetSQLValueString(hash($hash , $secret_key.$user_profile['id']), "text"),
					   GetSQLValueString($_GET['id'], "int"),
					   GetSQLValueString(time(), "int"));

      $result = $connection->query($insertSQL);
}

$insertGoTo = "?p=ver&id=".$_GET['id'];
header(sprintf("Location: %s", $insertGoTo));
?>