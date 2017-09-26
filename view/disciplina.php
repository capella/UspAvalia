<?php 


$pesquisa = '';

if (isset($_GET['pesquisa'])) {
  $pesquisa = $_GET['pesquisa'];
}
$startRow_Pesquisa = $pageNum_Pesquisa * $maxRows_Pesquisa;

mysql_select_db($database_connection, $connection);
$query_Pesquisa;
if($_GET['t']==1)
	$query_Pesquisa = "SELECT * FROM (SELECT AP.id, AP.idaula, AP.idprofessor, DIS.nome as 'Dnome', PRO.nome as 'Pnome', codigo FROM aulaprofessor AP INNER JOIN disciplinas DIS ON AP.idaula = DIS.id INNER JOIN professores PRO ON AP.idprofessor = PRO.id) S WHERE idaula =".GetSQLValueString($_GET['id'], "text2");
else
	$query_Pesquisa = "SELECT * FROM (SELECT AP.id, AP.idaula, AP.idprofessor, DIS.nome as 'Dnome', PRO.nome as 'Pnome', codigo FROM aulaprofessor AP INNER JOIN disciplinas DIS ON AP.idaula = DIS.id INNER JOIN professores PRO ON AP.idprofessor = PRO.id) S WHERE idprofessor =".GetSQLValueString($_GET['id'], "text2");
	
	
$Pesquisa = mysql_query($query_Pesquisa, $connection) or die(mysql_error());
$row_Pesquisa = mysql_fetch_assoc($Pesquisa);

if (isset($_GET['totalRows_Pesquisa'])) {
  $totalRows_Pesquisa = $_GET['totalRows_Pesquisa'];
} else {
  $all_Pesquisa = mysql_query($query_Pesquisa);
  $totalRows_Pesquisa = mysql_num_rows($all_Pesquisa);
}

function media($PAid){
	global $connection;
	$query_Media = "SELECT AVG(nota) as m FROM votos WHERE APid = ".$PAid." AND tipo <> 5 GROUP BY APid;";	
	$Media= mysql_query($query_Media, $connection) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	if(mysql_num_rows($Media)==1)
		return number_format($row_Media['m']*2, 2, ',', ' ');
	else
		return 'Sem avaliações';
}
$nome;

?>

<h2>Pesquisa</h2>
<form  method="get" action="/">
  <div class="input-group">
    <input type="text" class="form-control"  name="pesquisa" value="<?=$pesquisa;?>">
    <span class="input-group-btn">
      <input type="hidden" name="p" value="pesquisa" />	
      <button class="btn btn-default" type="submit">Pesquisar!</button>
    </span>
  </div><!-- /input-group -->
</form>

<p>

  <div class="table-responsive">
  <table class="table table-striped">
    <thead><tr><th>Nome Professor</th><th>Aula</th><th>Nota (0-10)</th><th>#</th><th>#</th></tr></thead>
    <?php do { 
	$nome = $row_Pesquisa['Pnome'];
	?>
    <tr>
    	<td><a href="?p=ver&id=<?= $row_Pesquisa['id'];?>"><?=$row_Pesquisa['Pnome'];?></a></td>
        <td><a href="?p=ver&id=<?= $row_Pesquisa['id'];?>"><?=$row_Pesquisa['Dnome'];?> - <?=$row_Pesquisa['codigo'];?></a></td>
        <td><?=media($row_Pesquisa['id']);?></td>
        <td>
        <button class="btn btn-success" data-toggle="modal" data-target="#modal<?=$row_Pesquisa['id'];?>">
          Avaliar
        </button>
        <!-- Modal -->
        <div class="modal fade" id="modal<?=$row_Pesquisa['id'];?>" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
          <div class="modal-dialog">
            <div class="modal-content">
				<? include('view/modal.php'); ?>
            </div>
          </div>
        </div>
        </td>
        <td><a href="?p=ver3&id=<?= $row_Pesquisa['id'];?>" class="btn btn-info">Comentar</a></td>
    </tr>
    <?php } while ($row_Pesquisa = mysql_fetch_assoc($Pesquisa)); ?>
  </table>
  </div>
</p>
<hr />

<? if($_GET['t']==2): ?>
  	<div class="label label-info"> Não encontrou a disciplina com esse professor? <?=$pesquisa;?>. <a href="?p=add&prf=<?=$nome;?>"> Clique aqui para adicionar. </a></div>
<? endif; ?>

<p><small>Foram encontrados <?php echo $totalRows_Pesquisa ?> registros.</small></p>

<?php
mysql_free_result($Pesquisa);
?>
