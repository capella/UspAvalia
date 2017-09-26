<?php 
$arr = array('Avaliação Geral', 'Didática', 'Empenho/Dedicação', 'Relação com os alunos', 'Dificuldade');

if (isset($_GET['id'])) {
   $id = GetSQLValueString($_GET['id'], "int");
   $sql = "
      SELECT AP.id, DIS.nome as 'Dnome', PRO.nome as 'Pnome', codigo, media
      FROM aulaprofessor AP
      INNER JOIN disciplinas DIS ON AP.idaula = DIS.id
      INNER JOIN professores PRO ON AP.idprofessor = PRO.id
      LEFT JOIN (
          SELECT APid, AVG(nota) as media FROM votos WHERE tipo <> 5 GROUP BY APid
      ) MED ON MED.APid = AP.id
      WHERE AP.id =".$id;
   $info_result = $connection->query($sql);
   $info = $info_result->fetch_assoc();

   $sql = "
      SELECT tipo, COUNT(*) as count, STDDEV_POP(nota)*2 as std, AVG(nota)*2 as avg
      FROM votos
      WHERE APid =".$id."
      GROUP BY APid, tipo";
   $medias_result = $connection->query($sql);

   $sql = "
      SELECT cometario.*, IFNULL(k.negativos, 0) as negativos, IFNULL(p.positivos, 0) as positivos 
      FROM cometario 
      LEFT OUTER JOIN (
         SELECT `idcomentario`, SUM(voto)*-1 as negativos
         FROM `votoscomentario` WHERE `voto`=-1
         GROUP BY `idcomentario`) k 
      ON cometario.id=k.`idcomentario` 
      LEFT JOIN (
         SELECT `idcomentario`, SUM(voto) as positivos
         FROM `votoscomentario` WHERE `voto`=1
         GROUP BY `idcomentario`) p 
      ON cometario.id=p.`idcomentario`
      WHERE aulaprofessorid = ".$id." 
      AND (IFNULL(p.positivos, 0)-IFNULL(k.negativos, 0))>=-3 
      ORDER BY `time` DESC";
   $comentarios_result = $connection->query($sql);
}
?>
<title>
<?=$info['Pnome'];?> -
<?=$info['codigo'];?>
</title>

<h2>
   <span itemprop="title"><?=$info['Pnome'];?> </span>
   <small> 
      <span itemprop="affiliation"><?=$info['Dnome'];?> - <?=$info['codigo'];?></span>
   </small>
</h2>

<!-- Avaliacoes -->
<hr>
<div class="row">
   <div class="thumbnails">
      <?php while ($row = $medias_result->fetch_assoc()) { ?>
      <div class="col-md-4">
         <div class="thumbnail">
            <div class="caption">
               <h4><?= $arr[$row['tipo']-1]; ?></h4>
               <? if($row['count'] != 0) :?>
                  <h3 style="text-align:center"><?= number_format($row['avg'], 3, ',', ' '); ?></h3>
                  <p class="graph" avg="<?= $row['avg']; ?>" std="<?= $row['std']; ?>"></p>
                  <p><small>Quesito avaliado <?= $row['count']; ?> vezes.</small></p>
               <? else: ?>
                  <p>
                  <br>
                  <h4 style="text-align:center">Sem avaliações</h4>
                  <br>
                  <br> </p>
               <? endif ?>
            </div>
         </div>
      </div>
      <?php } ?>

      <?php $row = $info; ?>
      <div class="col-md-4" style="text-align: center;">
         <button class="btn btn-success btn-block btn-lg" data-toggle="modal" data-target="#modal<?=$row['id'];?>" style="margin-top: 20%;">
            AVALIAR
         </button>
      </div>
      <!-- Modal -->
      <div class="modal fade" id="modal<?=$row['id'];?>" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
         <div class="modal-dialog">
            <div class="modal-content">
               <? include('view/modal.php'); ?>
            </div>
         </div>
      </div>

   </div>
</div>
<hr>

<!-- Comentarios -->
<div class="row">
   <div class="col-md-8">
      <h3>Comentários</h3>
      <p style="text-align: justify;">
      Área destina a comentários referente a disciplina <?=$row['Dnome'];?> - <?=$row['codigo'];?>
      com o professor <?=$row['Pnome'];?>. Os comentários devem ser feitos com o objetivo de auxiliar
      o docente a melhorar a qualidade das aulas. Elogios também são aceitos. Comentários com muitas
      qualificações negativas serão automaticamente retirados. Os comentários também são anônimos.<br>
      Se você ler um comentário verídico e relevante, aperte "Positivo", caso contrário, "Negativo".
      </p>
      <hr>
      <? if($comentarios_result->num_rows > 0): ?>
         <ul class="media-list">
            <?php while ($row = $comentarios_result->fetch_assoc()) { ?>
            <li class="media">
               <div class="media-body">
                  <h4 class="media-heading"><?php echo date('d/m/Y - H:i:s',$row['time']); ?></h4>
                  <div class="row">
                     <div class="col-md-8" style=" text-align:justify">
                        <?= $row['comantario']; ?>
                     </div>
                     <div class="col-md-4" style="opacity: 0.75;">
                        <a class="btn btn-success btn-block" href="?p=votarcomentario&idcomantario=<?= $row['id']; ?>&id=<?= $id; ?>&voto=1">
                           <span class="glyphicon glyphicon-thumbs-up" aria-hidden="true"></span> 
                           Positivo&nbsp;&nbsp;&nbsp;
                           <span class="badge"><?= $row['positivos']; ?></span>
                        </a>
                        <a class="btn btn-danger btn-block" href="?p=votarcomentario&idcomantario=<?= $row['id']; ?>&id=<?= $id; ?>&voto=-1">
                           <span class="glyphicon glyphicon-thumbs-down" aria-hidden="true"></span> 
                           Negativo&nbsp;&nbsp;&nbsp;
                           <span class="badge"><?= $row['negativos']; ?></span>
                        </a>
                     </div>
                  </div>
               </div>
            </li>
            <?php } ?>
         </ul>
      <? else: ?>
         <div class="alert alert-warning" style="opacity: 0.5;">Sem comentátios</div>
      <? endif; ?>
      <hr>
   </div>
   <div class="col-md-4">
      <div class="well">
        <form name="form" role="form" method="GET" action=""> 
            <div class="form-group">
               <label for="exampleInputPassword1">Novo Comentário:</label>
               <textarea class="form-control" name="comentario" rows="6"></textarea>
            </div>
            <button type="submit" class="btn btn-default btn-block">Salvar</button>
            <input type="hidden" name="MM_insert" value="form">
            <input type="hidden" name="p" value="comentar">
            <input type="hidden" name="id" value="<?= $id; ?>">
         </form>
        </div>
   </div>
</div>
<?php if ($info_result) $info_result->close(); ?>
<?php if ($medias_result) $medias_result->close(); ?>
<? require 'view/gdata.php' ?>
