<?php 


$pesquisa = '';

if (isset($_GET['pesquisa'])) {
  $pesquisa = GetSQLValueString($_GET['pesquisa'], "text2");
}
$startRow_Pesquisa = $pageNum_Pesquisa * $maxRows_Pesquisa;

mysql_select_db($database_connection, $connection);
$query_Pesquisa = "
(
SELECT id, idunidade, nome, codigo,  '1' AS  'tipo'
FROM  `disciplinas` 
WHERE  `nome` LIKE  '%".$pesquisa."%'
OR  `codigo` LIKE  '%".$pesquisa."%'
)
UNION (
SELECT id, idunidade, nome,  '' AS  'codigo',  '2' AS  'tipo'
FROM professores
WHERE  `nome` LIKE  '%".$pesquisa."%'
)
ORDER BY `nome` ASC LIMIT 0, 100;";
$Pesquisa = mysql_query($query_Pesquisa, $connection) or die(mysql_error());
$row_Pesquisa = mysql_fetch_assoc($Pesquisa);
$totalRows_Pesquisa = mysql_num_rows($Pesquisa);


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
  <? if($totalRows_Pesquisa>0): ?>
  <table class="table table-striped">
    <thead><tr><th>Nome</th><th>Tipo</th></tr></thead>
    <?php do { ?>
      <?
        if($row_Pesquisa['tipo']==1)
            echo '<tr><td><a href="?p=pesquisa2&id='.$row_Pesquisa['id'].'&t=1">'.$row_Pesquisa['nome'].' - '.$row_Pesquisa['codigo'].'<a/></td><td>Disciplina</td></tr>';  
        else {
            echo '<tr><td><a href="?p=pesquisa2&id='.$row_Pesquisa['id'].'&t=2">'.$row_Pesquisa['nome'].'<a/></td><td>Professor(a)</td></tr>';   
        }
      ?>
    <?php } while ($row_Pesquisa = mysql_fetch_assoc($Pesquisa)); ?>
  </table>
  <? else: ?>
  	<div class="alert alert-danger">Não encontramos nada com <?=$pesquisa;?>. <a href="?p=add"> Deseja criar adicionar nova disciplina ou professor? Clique aqui. </a></div>
  <? endif; ?>
</p>
<hr />
<small>Foram encontrados <?php echo $totalRows_Pesquisa ?> registros.</small>


<?php
mysql_free_result($Pesquisa);
?>
