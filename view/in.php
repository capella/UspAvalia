<?
mysql_select_db($database_connection, $connection);

function media(){
	global $connection;
	$query_Media = "SELECT AVG(nota) as m FROM votos WHERE tipo <> 5;";	
	$Media= mysql_query($query_Media, $connection) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	if(mysql_num_rows($Media)==1)
		return number_format($row_Media['m']*2, 2, ',', ' ');
	else
		return '8,38';
}

function avaliacoes(){
	global $connection;
	$query_Media = "SELECT COUNT(*) as m FROM votos;";	
	$Media= mysql_query($query_Media, $connection) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	if(mysql_num_rows($Media)==1)
		return number_format($row_Media['m'], 0, ',', ' ');
	else
		return '0';
}

function pessoas(){
	global $connection;
	$query_Media = "SELECT * FROM votos GROUP BY iduso;";	
	$Media= mysql_query($query_Media, $connection) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	return number_format(mysql_num_rows($Media), 0, ',', ' ');
}


?>
      <div class="starter-template">
      	<? $r = rand(1,3); 
		if($r==1){ ?>
        	<h1>USP Avalia <small style="font-size:10px;" class="hidden-xs"> <?=media();?> - Nota média da  universidade.</small></h1>
        <? } else if($r==2){ ?>
        	<h1>USP Avalia <small style="font-size:10px;" class="hidden-xs"> <?=avaliacoes();?> - Avaliações realizadas</small></h1>
        <? } else if($r==3){ ?>
        	<h1>USP Avalia <small style="font-size:10px;" class="hidden-xs"> Precisamos de  um logo. Alguma Sugestão? </small></h1>
        <? } else if($r==4){ ?>
        	<h1>USP Avalia <small style="font-size:10px;" class="hidden-xs"> <?=pessoas();?>  - Pessoas já ajudaram</small></h1>
        <? } ?>
        <p class="lead">Pesquise uma disciplina, um professor ou uma sigla.
        <br /><small style="font-size:10px;"> Depois avalie e ajude a comunidade.</small>
        </p>
        <p class="lead">
          <form method="get" action="/">
          <div class="input-group">
            <input type="text" class="form-control" name="pesquisa" autofocus="autofocus" autocomplete="off">
            <span class="input-group-btn">
              <input type="hidden" name="p" value="pesquisa" />	
              <button class="btn btn-default" type="submit">Pesquisar</button>
            </span>
          </div><!-- /input-group -->
          </form>
        </p>
      </div>