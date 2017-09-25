<?php 


$pesquisa = '';

if (isset($_GET['pesquisa'])) {
    $pesquisa = $_GET['pesquisa'];
}
$startRow_Pesquisa = $pageNum_Pesquisa * $maxRows_Pesquisa;

mysql_select_db($database_connection, $connection);
$query_Pesquisa;
    $query_Pesquisa = "SELECT * FROM (SELECT AP.id, AP.idaula, AP.idprofessor, DIS.nome as 'Dnome', PRO.nome as 'Pnome', codigo FROM aulaprofessor AP INNER JOIN disciplinas DIS ON AP.idaula = DIS.id INNER JOIN professores PRO ON AP.idprofessor = PRO.id) S WHERE id = ".$_GET['id'];
    
$Pesquisa = mysql_query($query_Pesquisa, $connection) or die(mysql_error());
$row_Pesquisa = mysql_fetch_assoc($Pesquisa);

if (isset($_GET['totalRows_Pesquisa'])) {
    $totalRows_Pesquisa = $_GET['totalRows_Pesquisa'];
} else {
    $all_Pesquisa = mysql_query($query_Pesquisa);
    $totalRows_Pesquisa = mysql_num_rows($all_Pesquisa);
}

function mediap($PAid, $tipo)
{
    global $connection;
    $query_Media = "SELECT AVG(nota) as m, COUNT(*) as j, STDDEV_POP(nota) as dp FROM votos WHERE APid = ".$PAid." AND tipo = ".$tipo." GROUP BY APid;";    
    $Media= mysql_query($query_Media, $connection) or die(mysql_error());
    $row_Media = mysql_fetch_assoc($Media);
    $data;
    if(mysql_num_rows($Media)==1) {
        $data[0] = number_format($row_Media['m']*2, 3, ',', ' ');
        $data[1] = $row_Media['m']*2;
        $data[2] = $row_Media['j'];
        $data[3] = $row_Media['dp']*2;
    }else{
        $data[0] = 'XX';
        $data[1] = 'XX';
    }
    return $data;
}

function desvio($tipo)
{
    global $connection;
    $query_Media = "SELECT STDDEV_POP(nota) g FROM votos WHERE tipo = ".$tipo;    
    $Media= mysql_query($query_Media, $connection) or die(mysql_error());
    $row_Media = mysql_fetch_assoc($Media);
    if(mysql_num_rows($Media)==1) {
        return $row_Media['g']*2; 
    }
    else {
        return '0'; 
    }
}

function media($tipo)
{
    global $connection;
    $query_Media = "SELECT AVG(nota) as m  FROM votos WHERE tipo = ".$tipo.";";    
    $Media= mysql_query($query_Media, $connection) or die(mysql_error());
    $row_Media = mysql_fetch_assoc($Media);
    if(mysql_num_rows($Media)==1) {
        return $row_Media['m']*2; 
    }
    else {
        return 'XX'; 
    }
}

function mediageral($PAid)
{
    global $connection;
    $query_Media = "SELECT AVG(nota) as m FROM votos WHERE APid = ".$PAid." AND tipo <> 5 GROUP BY APid;";    
    $Media= mysql_query($query_Media, $connection) or die(mysql_error());
    $row_Media = mysql_fetch_assoc($Media);
    if(mysql_num_rows($Media)==1) {
        return number_format($row_Media['m']*2, 1, '.', ''); 
    }
    else {
        return ''; 
    }
}
?>
    <? require 'view/gdata.php'?>

<h2><span itemprop="title"><?=$row_Pesquisa['Pnome'];?> </span><small>  <span itemprop="affiliation"><?=$row_Pesquisa['Dnome'];?> - <?=$row_Pesquisa['codigo'];?></span></small></h2>


        <title><?=$row_Pesquisa['Pnome'];?> - <?=$row_Pesquisa['codigo'];?></title>
        
    <ul id="myTab" class="nav nav-tabs">
      <li role="presentation" class="active"><a href="?p=ver&id=<?=$_GET['id']?>" role="tab">Avaliações/Notas</a></li>
      <li role="presentation" class=""><a href="?p=ver3&id=<?=$_GET['id']?>" role="tab">Comentários</a></li>
    </ul>


<p>
  <div class="row">
    <?
    $arr = array('Avaliação Geral', 'Didática', 'Empenho/Dedicação');
    reset($arr);
    while (list($key, $value) = each($arr)) {
        $chave = $key+1;
        $mp = mediap($row_Pesquisa['id'], $chave);
        $m = $mp[1];
        $d = $mp[3];
    ?> 
    <div class="col-sm-6 col-md-4">
      <div class="thumbnail">
        <div class="caption">
          <h4><?=$value;?></h4>
            <? if($mp[0]!='XX') :?>
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
        $m = $mp[1];
        $d = $mp[3];
    ?>  
    <div class="col-sm-6 col-md-4">
      <div class="thumbnail">
        <div class="caption">
          <h4><?=$value;?></h4>
            <? if($mp[0]!='XX') :?>
          <p><h3 style="text-align:center"><?= $mp[0]; ?></h3></p>
          <p id="bar<?=$chave;?>">
            <script> bar('#bar<?=$chave;?>',10,0,<?=$d+$m;?>,<?=-$d+$m;?>,<?=$mp[1];?>);</script>
          </p>
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
                    <? include 'view/modal.php'; ?>
                </div>
              </div>
            </div>
           </p>
        <?php } while ($row_Pesquisa = mysql_fetch_assoc($Pesquisa)); ?>
    </div>
  </div>
  
  </p>
<hr />
<?php
mysql_free_result($Pesquisa);
?>
