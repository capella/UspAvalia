<?php

function GetSQLValueString($theValue, $theType, $theDefinedValue = "", $theNotDefinedValue = "")  {
  if (PHP_VERSION < 6) {
    $theValue = get_magic_quotes_gpc() ? stripslashes($theValue) : $theValue;
  }

  $theValue = function_exists("mysql_real_escape_string") ? mysql_real_escape_string($theValue) : mysql_escape_string($theValue);

  switch ($theType) {
    case "text":
      if ($theValue != "") {
        $theValue = strip_tags($theValue, '<br>');
        $theValue = "'" . $theValue . "'";
      } else {
        $theValue = "NULL";
      }
      break; 
    case "text2":
      if ($theValue != "") {
        $theValue = strip_tags($theValue, '<br>');
        $theValue = $theValue;
      } else {
        $theValue = "NULL";
      }
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

?>