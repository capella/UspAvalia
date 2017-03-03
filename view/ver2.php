<? include('view/gdata.php'); ?>
<?
$row_Pesquisa2 = $row_Pesquisa;
?>
<p>
  <div class="row">
	<?
    $arr = array('Avaliação Geral', 'Didática', 'Empenho/Dedicação');
    reset($arr);
    while (list($key, $value) = each($arr)) {
        $chave = $key+1;
		$mp = mediap($row_Pesquisa['id'], $chave);
		$m = media($chave);
		$d = desvio($chave);
    ?> 
    <div class="col-sm-6 col-md-4">
      <div class="thumbnail">
        <div class="caption">
          <h4><?=$value;?></h4>
          <? if($mp[0]!='XX'):?>
          <p><h3 style="text-align:center"><?=$mp[0];?></h3></p>
          <p id="bar<?=$chave;?>"><script> bar('#bar<?=$chave;?>',10,0,<?=$d+$m;?>,<?=-$d+$m;?>,<?=$mp[1];?>);</script></p>
          <p><small>Quesito avaliado <?=$mp[2];?> vezes.</small></p>
          <? else: ?>
          <p><br><h4 style="text-align:center">Sem avaliações</h4><br><br></p>
          <? endif ?>
        </div>
      </div>
    </div>
    <?php } ?>
  </div>

  <div class="row">
	<?
    $arr2 = array('Relação com os alunos', 'Dificuldade');
    reset($arr2);
    while (list($key, $value) = each($arr2)) {
        $chave = $key+4;
		$mp = mediap($row_Pesquisa['id'], $chave);
		$m = media($chave);
		$d = desvio($chave);
    ?>  
    <div class="col-sm-6 col-md-4">
      <div class="thumbnail">
        <div class="caption">
          <h4><?=$value;?></h4>
          <? if($mp[0]!='XX'):?>
          <p><h3 style="text-align:center"><?=$mp[0];?></h3></p>
          <p id="bar<?=$chave;?>"><script> bar('#bar<?=$chave;?>',10,0,<?=$d+$m;?>,<?=-$d+$m;?>,<?=$mp[1];?>);</script></p>
          <p><small>Quesito avaliado <?=$mp[2];?> vezes.</small></p>
          <? else: ?>
          <p><br><h4 style="text-align:center">Sem avaliações</h4><br><br></p>
          <? endif ?>
        </div>
      </div> 
    </div>
    <?php } ?>
    <div class="col-sm-6 col-md-4">
    	  <?php do { ?>
          <p align="center">
          <br> <br> <br>
            <button class="btn btn-success btn-block btn-lg" data-toggle="modal" data-target="#modal<?=$row_Pesquisa['id'];?>">
              Avaliar
            </button>
            <!-- Modal -->
            <div class="modal fade" id="modal<?=$row_Pesquisa['id'];?>" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
              <div class="modal-dialog">
                <div class="modal-content">
                    <? include('view/modal.php')?>
                </div>
              </div>
            </div>
           </p>
           <?php } while ($row_Pesquisa = mysql_fetch_assoc($Pesquisa)); ?>
    </div>
  </div>
  

  

  </p>
<hr />