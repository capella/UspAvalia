<?php 

if (isset($_GET['pesquisa'])) {
  $data = GetSQLValueString($_GET['pesquisa'], "text2");
  $sql = "
  SELECT * FROM (
     (SELECT id, CONCAT(codigo, ' - ', nome) as nome, 1 as type from disciplinas
     WHERE MATCH(nome, codigo) AGAINST('*".$data."*' IN BOOLEAN MODE)
     ORDER BY MATCH(nome, codigo) AGAINST('*".$data."*' IN BOOLEAN MODE) DESC
     LIMIT 40)
     UNION
     (SELECT id, nome, 0 as type FROM professores
     WHERE MATCH(nome) AGAINST('*".$data."*' IN BOOLEAN MODE)
     ORDER BY MATCH(nome) AGAINST('*".$data."*' IN BOOLEAN MODE) DESC
     LIMIT 40)
  ) NAMES
  LIMIT 80"; 

  $result = $connection->query($sql);
}

?>
<h2>Pesquisa</h2>
<form  method="get" action="/">
  <div class="input-group">
    <input type="text" class="form-control typeahead"  name="pesquisa" value="<?=$data;?>" autocomplete="off">
    <span class="input-group-btn">
      <input type="hidden" name="p" value="pesquisa" />	
      <button class="btn btn-default" type="submit">Pesquisar!</button>
    </span>
  </div><!-- /input-group -->
</form>

<p>
  <? if($result && $result->num_rows > 0): ?>
  <table class="table table-striped">
    <thead><tr><th>Nome</th><th>Tipo</th></tr></thead>
    <?php while ($row = $result->fetch_assoc()) { ?>
      <?
        if($row['type']==1)
            echo '<tr><td><a href="?p=disciplina&id='.$row['id'].'">'.$row['nome'].'<a/></td><td>Disciplina</td></tr>';  
        else {
            echo '<tr><td><a href="?p=professor&id='.$row['id'].'">'.$row['nome'].'<a/></td><td>Professor(a)</td></tr>';   
        }
      ?>
    <?php } ?>
  </table>
  <? else: ?>
  	<div class="alert alert-danger">Não encontramos nada com <?=$pesquisa;?>. <a href="?p=add"> Deseja criar adicionar nova disciplina ou professor? Clique aqui. </a></div>
  <? endif; ?>
</p>
<hr />
<small>Foram encontrados <?= $result->num_rows ?> registros.</small>

<?php if ($result) $result->close(); ?>
