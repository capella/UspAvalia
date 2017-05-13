<?php
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
  return $theValue;
}
}

mysql_select_db($database_CapellaResumo, $CapellaResumo);
$query_Melhores = "SELECT * FROM Melhores LIMIT 10";
$Melhores = mysql_query($query_Melhores, $CapellaResumo) or die(mysql_error());
$row_Melhores = mysql_fetch_assoc($Melhores);
$totalRows_Melhores = mysql_num_rows($Melhores);
?>
<h2>10 Melhores</h2>
<hr />
<h4>Cursos Avaliados</h4>
<table class="table table-bordered">
  <tr>
    <td><div align="left"><strong>Média</strong></div></td>
    <td><div align="left"><strong>Votos</strong></div></td>
    <td><div align="left"><strong>Disciplina</strong></div></td>
    <td><div align="left"><strong>Unidade</strong></div></td>
    <td><div align="left"><strong>Professor</strong></div></td>
    <td><div align="left"></div></td>
  </tr>
  <?php do { ?>
    <tr>
      <td><div align="center"><?php echo $row_Melhores['media']; ?></div></td>
      <td><div align="center"><?php echo $row_Melhores['votos']; ?></div></td>
      <td><?php echo $row_Melhores['codigo']; ?>-<?php echo $row_Melhores['materia']; ?></td>
      <td><?php echo $row_Melhores['unidade']; ?></td>
      <td><?php echo $row_Melhores['professor']; ?></td>
      <td>
        <div align="center"><a href="<?= $url_full; ?>/?p=ver&id=<?php echo $row_Melhores['id']; ?>" class="btn btn-success  btn-sm">
          Avaliar
          </a>
      </div></td>
    </tr>
    <?php } while ($row_Melhores = mysql_fetch_assoc($Melhores)); ?>
</table>
<?
mysql_free_result($Melhores);

mysql_select_db($database_CapellaResumo, $CapellaResumo);
$query_Melhores = "SELECT
  professores.nome,
  SubQuery.expr1 * 2 AS nota,
  SubQuery.expr2 AS votos,
  unidades.NOME AS unidade,
  professores.id as id
FROM (SELECT
    votos.id,
    aulaprofessor.idprofessor,
    AVG(votos.nota) AS expr1,
    COUNT(*) AS expr2
  FROM votos
    INNER JOIN aulaprofessor
      ON votos.APid = aulaprofessor.id
  WHERE votos.tipo <> 5
  GROUP BY aulaprofessor.idprofessor) SubQuery
  INNER JOIN professores
    ON SubQuery.idprofessor = professores.id
  INNER JOIN unidades
    ON professores.idunidade = unidades.id
WHERE SubQuery.expr2 >= 15
ORDER BY SubQuery.expr1 DESC, SubQuery.expr2 DESC
LIMIT 10";
$Melhores = mysql_query($query_Melhores, $CapellaResumo) or die(mysql_error());
$row_Melhores = mysql_fetch_assoc($Melhores);
$totalRows_Melhores = mysql_num_rows($Melhores);
?>
<h4>Professores Avaliados</h4>
<table class="table table-bordered">
  <tr>
    <td><div align="left"><strong>Média</strong></div></td>
    <td><div align="left"><strong>Votos</strong></div></td>
    <td><div align="left"><strong>Professor</strong></div></td>
    <td><div align="left"><strong>Unidade</strong></div></td>
    <td><div align="left"></div></td>
  </tr>
  <?php do { ?>
    <tr>
      <td><div align="center"><?php echo $row_Melhores['nota']; ?></div></td>
      <td><div align="center"><?php echo $row_Melhores['votos']; ?></div></td>
      <td><?php echo $row_Melhores['nome']; ?></td>
      <td><?php echo $row_Melhores['unidade']; ?></td>
      <td>
        <div align="center"><a href="<?= $url_full; ?>/?p=pesquisa2&id=<?php echo $row_Melhores['id']; ?>&t=2" class="btn btn-success  btn-sm">
          Avaliar
          </a>
      </div></td>
    </tr>
    <?php } while ($row_Melhores = mysql_fetch_assoc($Melhores)); ?>
</table>
<small>Só foram avaliados professores com 15 ou mais votos.</small> 
<br /><br />
<?
mysql_free_result($Melhores);
?>
