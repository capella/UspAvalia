<?php

$colname_Comentarios = "-1";
if (isset($_GET['id'])) {
  $colname_Comentarios = $_GET['id'];
}
mysql_select_db($database_connection, $connection);
$query_Comentarios = sprintf("
SELECT cometario.*, IFNULL(k.negativos, 0) as negativos, IFNULL(p.positivos, 0) as positivos 
	FROM cometario 
		LEFT OUTER JOIN (SELECT `idcomentario`, SUM(voto)*-1 as negativos FROM `votoscomentario` WHERE `voto`=-1 GROUP BY `idcomentario`) k 
		ON cometario.id=k.`idcomentario` 
		LEFT JOIN (SELECT `idcomentario`, SUM(voto) as positivos FROM `votoscomentario` WHERE `voto`=1 GROUP BY `idcomentario`) p 
		ON cometario.id=p.`idcomentario`
	WHERE aulaprofessorid = %s 
		AND (IFNULL(p.positivos, 0)-IFNULL(k.negativos, 0))>=-3 
	ORDER BY `time` DESC", GetSQLValueString($colname_Comentarios, "int"));
$Comentarios = mysql_query($query_Comentarios, $connection) or die(mysql_error());
$row_Comentarios = mysql_fetch_assoc($Comentarios);
$totalRows_Comentarios = mysql_num_rows($Comentarios);
?>
<div class="row">
	<div class="col-md-8">
        <p><br>
        Área destina a comentários referente a disciplina <?=$row_Pesquisa2['Dnome'];?> - <?=$row_Pesquisa2['codigo'];?> com o professor <?=$row_Pesquisa2['Pnome'];?>. Os comentários devem ser feitos com o objetivo de auxiliar o docente a melhorar a qualidade das aulas. Elogios também são aceitos.
        Comentários com muitas qualificações negativas serão automaticamente retirados. Os comentários também são anônimos.<br>
        Se você ler um comentário verídico e relevante, aperte "Positivo", caso contrário, "Negativo".
        </p>
        <hr>
        
        <? if($totalRows_Comentarios>0): ?>
            <ul class="media-list">
            <?php do { ?>
                <li class="media">
                  <div class="media-body">
                    <h4 class="media-heading"><?php echo date('d/m/Y - H:i:s',$row_Comentarios['time']); ?></h4>
                    <div class="row">
                    	<div class="col-md-8" style=" text-align:justify">
                        <?= $row_Comentarios['comantario']; ?>
                        </div>
                    	<div class="col-md-4">
                        	<a class="btn btn-success btn-block" href="?p=votarcomentario&idcomantario=<?= $row_Comentarios['id']; ?>&id=<?=$_GET['id']?>&voto=1">
                              <span class="glyphicon glyphicon-thumbs-up" aria-hidden="true"></span> Positivo&nbsp;&nbsp;&nbsp; <span class="badge"><?= $row_Comentarios['positivos']; ?></span>
                            </a>
                        	<a class="btn btn-danger btn-block" href="?p=votarcomentario&idcomantario=<?= $row_Comentarios['id']; ?>&id=<?=$_GET['id']?>&voto=-1">
                              <span class="glyphicon glyphicon-thumbs-down" aria-hidden="true"></span> Negativo&nbsp;&nbsp;&nbsp; <span class="badge"><?= $row_Comentarios['negativos']; ?></span>
                            </a>
                        </div>
                    </div>
                    
               
                </li>
            <?php } while ($row_Comentarios = mysql_fetch_assoc($Comentarios)); ?>
            </ul>
		<? else: ?>
        	<div class="alert alert-warning">Sem comentátios</div>
        <? endif; ?>
        <hr>
	</div>
    <div class="col-md-4">
    <br>
    	<div class="well">
        <form name="form" role="form" method="GET" action=""> 
            <div class="form-group">
            	<label for="exampleInputPassword1">Novo Comentário:</label>
            	<textarea class="form-control" name="comentario" rows="6"></textarea>
            </div>
            <button type="submit" class="btn btn-default btn-block">Salvar</button>
            <input type="hidden" name="MM_insert" value="form">
            <input type="hidden" name="p" value="comentar">
            <input type="hidden" name="id" value="<?=$_GET['id']?>">
         </form>
        	
        </div>
	</div>
</div>
<?php
mysql_free_result($Comentarios);
?>
