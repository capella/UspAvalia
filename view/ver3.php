<?php 


$pesquisa = '';

if (isset($_GET['pesquisa'])) {
  $pesquisa = $_GET['pesquisa'];
}
$startRow_Pesquisa = $pageNum_Pesquisa * $maxRows_Pesquisa;

mysql_select_db($database_CapellaResumo, $CapellaResumo);
$query_Pesquisa;
	$query_Pesquisa = "SELECT * FROM (SELECT AP.id, AP.idaula, AP.idprofessor, DIS.nome as 'Dnome', PRO.nome as 'Pnome', codigo FROM aulaprofessor AP INNER JOIN disciplinas DIS ON AP.idaula = DIS.id INNER JOIN professores PRO ON AP.idprofessor = PRO.id) S WHERE id = ".$_GET['id'];
	
$Pesquisa = mysql_query($query_Pesquisa, $CapellaResumo) or die(mysql_error());
$row_Pesquisa = mysql_fetch_assoc($Pesquisa);

if (isset($_GET['totalRows_Pesquisa'])) {
  $totalRows_Pesquisa = $_GET['totalRows_Pesquisa'];
} else {
  $all_Pesquisa = mysql_query($query_Pesquisa);
  $totalRows_Pesquisa = mysql_num_rows($all_Pesquisa);
}

function mediap($PAid, $tipo){
	global $CapellaResumo;
	$query_Media = "SELECT AVG(nota) as m, COUNT(*) as j FROM votos WHERE APid = ".$PAid." AND tipo = ".$tipo." GROUP BY APid;";	
	$Media= mysql_query($query_Media, $CapellaResumo) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	$data;
	if(mysql_num_rows($Media)==1){
		$data[0] = number_format($row_Media['m']*2, 3, ',', ' ');
		$data[1] = number_format($row_Media['m']*2, 5, '.', ' ');
		$data[2] = $row_Media['j'];
	}else{
		$data[0] = 'XX';
		$data[1] = 'XX';
	}
	return $data;
}

function desvio($tipo){
	global $CapellaResumo;
	$query_Media = "SELECT STDDEV_POP(nota) g FROM votos WHERE tipo = ".$tipo;	
	$Media= mysql_query($query_Media, $CapellaResumo) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	if(mysql_num_rows($Media)==1)
		return $row_Media['g']*2;
	else
		return '0';
}

function media($tipo){
	global $CapellaResumo;
	$query_Media = "SELECT AVG(nota) as m  FROM votos WHERE tipo = ".$tipo.";";	
	$Media= mysql_query($query_Media, $CapellaResumo) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	if(mysql_num_rows($Media)==1)
		return $row_Media['m']*2;
	else
		return 'XX';
}

function mediageral($PAid){
	global $CapellaResumo;
	$query_Media = "SELECT AVG(nota) as m FROM votos WHERE APid = ".$PAid." AND tipo <> 5 GROUP BY APid;";	
	$Media= mysql_query($query_Media, $CapellaResumo) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	if(mysql_num_rows($Media)==1)
		return number_format($row_Media['m']*2, 1, '.', '');
	else
		return '';
}
?>

<h2><span itemprop="title"><?=$row_Pesquisa['Pnome'];?> </span><small>  <span itemprop="affiliation"><?=$row_Pesquisa['Dnome'];?> - <?=$row_Pesquisa['codigo'];?></span></small></h2>
<title><?=$row_Pesquisa['Pnome'];?> - <?=$row_Pesquisa['codigo'];?></title>

    <ul id="myTab" class="nav nav-tabs">
      <li role="presentation" class=""><a href="?p=ver&id=<?=$_GET['id']?>" role="tab">Avaliações/Notas</a></li>
      <li role="presentation" class="active"><a href="?p=ver3&id=<?=$_GET['id']?>" role="tab">Comentários</a></li>
    </ul>
  
  <? include('view/ver4.php'); ?>
  
<?php
mysql_free_result($Pesquisa);
?>
