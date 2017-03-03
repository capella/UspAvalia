<?php
# FileName="Connection_php_mysql.htm"
# Type="MYSQL"
# HTTP="true"
require 'config.php';
$hostname_CapellaResumo = $hostname;
$database_CapellaResumo = $database;
$username_CapellaResumo = $username;
$password_CapellaResumo = $password;
$CapellaResumo = mysql_pconnect($hostname_CapellaResumo, $username_CapellaResumo, $password_CapellaResumo) or trigger_error(mysql_error(),E_USER_ERROR);
mysql_set_charset('utf8',$CapellaResumo);

if (!function_exists("GetSQLValueString")) {
function GetSQLValueString($theValue, $theType, $theDefinedValue = "", $theNotDefinedValue = "") 
{
  if (PHP_VERSION < 6) {
    $theValue = get_magic_quotes_gpc() ? stripslashes($theValue) : $theValue;
  }

  $theValue = function_exists("mysql_real_escape_string") ? mysql_real_escape_string($theValue) : mysql_escape_string($theValue);

  switch ($theType) {
    case "text":
      $theValue = ($theValue != "") ? "'" . $theValue . "'" : "NULL";
      break;    
    case "long":
    case "int":
      $theValue = ($theValue != "") ? intval($theValue) : "NULL";
      break;
    case "double":
      $theValue = ($theValue != "") ? doubleval($theValue) : "NULL";
      break;
    case "date":
      $theValue = ($theValue != "") ? "'" . $theValue . "'" : "NULL";
      break;
    case "defined":
      $theValue = ($theValue != "") ? $theDefinedValue : $theNotDefinedValue;
      break;
  }
  return trim(preg_replace("/\r|\n/", "", trim(preg_replace('!\s+!', ' ', $theValue))));
}
}

?>