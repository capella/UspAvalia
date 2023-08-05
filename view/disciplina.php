<?php

if (isset($_GET['id'])) {
    $data = GetSQLValueString($_GET['id'], "int");
    $sql = "
      SELECT AP.id, DIS.nome as 'Dnome', PRO.nome as 'Pnome', codigo, media
      FROM aulaprofessor AP
      INNER JOIN disciplinas DIS ON AP.idaula = DIS.id
      INNER JOIN professores PRO ON AP.idprofessor = PRO.id
      LEFT JOIN (
          SELECT APid, AVG(nota) as media FROM votos WHERE tipo <> 5 GROUP BY APid
      ) MED ON MED.APid = AP.id
      WHERE DIS.id =".$data." ORDER by media DESC, PRO.nome ASC";
    $result = $connection->query($sql);
}

?>
<h2>Pesquisa</h2>
<form  method="get" action="/">
  <div class="input-group">
    <input type="text" class="form-control typeahead"  name="pesquisa" autocomplete="off">
    <span class="input-group-btn">
      <input type="hidden" name="p" value="pesquisa" />  
      <button class="btn btn-default" type="submit">Pesquisar!</button>
    </span>
  </div><!-- /input-group -->
</form>

<br>
<hr>

<?php if ($result && $result->num_rows > 0) { ?>
<div class="table-responsive">
   <table class="table table-striped">
      <thead>
         <tr>
            <th>Nome Professor</th>
            <th>Aula</th>
            <th>Nota (0-10)</th>
            <th>#</th>
            <th>#</th>
         </tr>
      </thead>
      <?php while ($row = $result->fetch_assoc()) { ?>
      <?php $nome = $row['Pnome']; ?>
      <tr>
         <td><a href="?p=ver&id=<?= $row['id'];?>"><?=$row['Pnome'];?></a></td>
         <td><a href="?p=ver&id=<?= $row['id'];?>"><?=$row['Dnome'];?> - <?=$row['codigo'];?></a></td>
         <td><?= $row['media'] ? number_format($row['media']*2, 2, ',', ' ') : "Sem avaliações";?></td>
         <td>
            <button class="btn btn-success" data-toggle="modal" data-target="#modal<?=$row['id'];?>">
               Avaliar
            </button>
            <!-- Modal -->
            <div class="modal fade" id="modal<?=$row['id'];?>" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
               <div class="modal-dialog">
                  <div class="modal-content">
                     <?php include('view/modal.php'); ?>
                  </div>
               </div>
            </div>
         </td>
         <td><a href="?p=ver&id=<?= $row['id'];?>" class="btn btn-info">Comentar</a></td>
      </tr>
      <?php } ?>
   </table>
</div>
<?php } else { ?>
<p>Não encontramos ninguem que ministre essa disciplina :(.</p>
<?php } ?>
<hr />

<p><small>Foram encontrados <?php echo $result->num_rows ?> registros.</small></p>

<?php if ($result) {
    $result->close();
} ?>
