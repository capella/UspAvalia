<?
if ($user) {
  $logoutUrl = $facebook->getLogoutUrl();
} else {
  //$statusUrl = $facebook->getLoginStatusUrl();
  $loginUrl = $facebook->getLoginUrl();
  header('Location: '.$loginUrl);
}

if (isset($_GET["id"])&&isset($user_profile['id'])&&($_GET['voto']==1||$_GET['voto']==-1)) {

  $insertSQL = sprintf("INSERT INTO votoscomentario (idcomentario, iduso, voto, time) VALUES (%s,%s,%s,%s) ON DUPLICATE KEY UPDATE time=VALUES(time), voto=VALUES(voto)",
                       GetSQLValueString($_GET['idcomantario'], "int"),
					   GetSQLValueString(hash($hash , $secret_key.$user_profile['id']), "text"),
					   GetSQLValueString($_GET['voto'], "int"),
					   GetSQLValueString(time(), "int"));

  mysql_select_db($database_CapellaResumo, $CapellaResumo);
  $Result1 = mysql_query($insertSQL, $CapellaResumo) or die(mysql_error());
}

$insertGoTo = "?p=ver3&id=".$_GET['id'];
header(sprintf("Location: %s", $insertGoTo));
?>