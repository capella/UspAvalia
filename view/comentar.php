<?
if ($user) {
  $logoutUrl = $facebook->getLogoutUrl();
} else {
  //$statusUrl = $facebook->getLoginStatusUrl();
  $loginUrl = $facebook->getLoginUrl();
  header('Location: '.$loginUrl);
}

if ((isset($_GET["MM_insert"])) && ($_GET["MM_insert"] == "form")) {

  $insertSQL = sprintf("INSERT INTO cometario (comantario, iduso, aulaprofessorid, time) VALUES (%s,%s,%s,%s)",
             GetSQLValueString($_GET['comentario'], "text"),
					   GetSQLValueString(hash($hash , $secret_key.$user_profile['id']), "text"),
					   GetSQLValueString($_GET['id'], "int"),
					   GetSQLValueString(time(), "int"));

  mysql_select_db($database_CapellaResumo, $CapellaResumo);
  $Result1 = mysql_query($insertSQL, $CapellaResumo) or die(mysql_error());
}

$insertGoTo = "?p=ver3&id=".$_GET['id'];
header(sprintf("Location: %s", $insertGoTo));
?>