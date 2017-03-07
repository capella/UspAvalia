<?
/*
if (mt_rand(0,1) == 0) {
    header('Location: v2/');
    exit;
}
*/

mysql_select_db($database_CapellaResumo, $CapellaResumo);

function media(){
	global $CapellaResumo;
	$query_Media = "SELECT AVG(nota) as m FROM votos WHERE tipo <> 5;";	
	$Media= mysql_query($query_Media, $CapellaResumo) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	if(mysql_num_rows($Media)==1)
		return number_format($row_Media['m']*2, 2, ',', ' ');
	else
		return '8,38';
}

function avaliacoes(){
	global $CapellaResumo;
	$query_Media = "SELECT COUNT(*) as m FROM votos;";	
	$Media= mysql_query($query_Media, $CapellaResumo) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	if(mysql_num_rows($Media)==1)
		return number_format($row_Media['m'], 0, ',', ' ');
	else
		return '0';
}

function pessoas(){
	global $CapellaResumo;
	$query_Media = "SELECT * FROM votos GROUP BY iduso;";	
	$Media= mysql_query($query_Media, $CapellaResumo) or die(mysql_error());
	$row_Media = mysql_fetch_assoc($Media);
	return number_format(mysql_num_rows($Media), 0, ',', ' ');
}


?>
      <div class="starter-template">
      	<? //$r = rand(1,3); 
		$r = 9;
		if($r==1){ ?>
        	<h1>USP Avalia <small style="font-size:10px;" class="hidden-xs"> <?=media();?> - Nota média da  universidade.</small></h1>
        <? } else if($r==2){ ?>
        	<h1>USP Avalia <small style="font-size:10px;" class="hidden-xs"> <?=avaliacoes();?> - Avaliações realizadas</small></h1>
        <? } else if($r==3){ ?>
        	<h1>USP Avalia <small style="font-size:10px;" class="hidden-xs"> Precisamos de  um logo. Alguma Sugestão? </small></h1>
        <? } else if($r==4){ ?>
        	<h1>USP Avalia <small style="font-size:10px;" class="hidden-xs"> <?=pessoas();?>  - Pessoas já ajudaram</small></h1>
        <? } else if($r==9){ ?>
        	<h1>USP Avalia</h1>
        <? } ?>
        <p class="lead">Pesquise uma disciplina, um professor ou uma sigla.
        <br /><small style="font-size:10px;"> Depois avalie e ajude a comunidade.</small>
        </p>
        <p class="lead">
          <form method="get" action="/">
          <div class="input-group">
            <input type="text" class="form-control" name="pesquisa" autofocus autocomplete="off">
            <span class="input-group-btn">
              <input type="hidden" name="p" value="pesquisa" />	
              <button class="btn btn-default" type="submit">Pesquisar</button>
            </span>
          </div><!-- /input-group -->
          </form>
        </p>
      </div>
      
       <div align="center">
                  <div class="row" style="  max-width: 620px;">
                        <div class="col-lg-4 col-xs-6">
                            <!-- small box -->
                            <div class="small-box bg-aqua">
                                <div class="inner">
                                    <h3>
                                         <?=media();?>
                                    </h3>
                                    <p>
                                        Nota média da <br /> universidade.
                                    </p>
                                </div>
                                <a href="#" class="small-box-footer">
                                </a>
                            </div>
                        </div><!-- ./col -->
                        <div class="col-lg-4 col-xs-6">
                            <!-- small box -->
                            <div class="small-box bg-green">
                                <div class="inner">
                                    <h3>
                                        <?=avaliacoes();?>
                                    </h3>
                                    <p>
                                        Avaliações realizadas
                                    </p><br />
                                </div>
                                <a href="#" class="small-box-footer">
                                </a>
                            </div>
                        </div><!-- ./col -->
                        <div class="col-lg-4 col-xs-6">
                            <!-- small box -->
                            <div class="small-box bg-yellow">
                                <div class="inner">
                                    <h3>
                                        <?=pessoas();?> 
                                    </h3>
                                    <p>
                                        Pessoas já ajudaram
                                    </p><br />
                                </div>
                                <a href="#" class="small-box-footer">
                                </a>
                            </div>
                        </div><!-- ./col -->
                      
                    </div><!-- /.row -->
    	 <p>Ajude a divulgar o USP Avalia em sua unidade com <a href="http://uspavalia.com/assets/images/poster.pdf" onclick="ga('send', 'event', 'Poster', 'click', 'pagina_inicial');">esse poster</a>.</p>

                     </div><!-- /.row -->


    <script type="application/ld+json">
        {
          "@context": "http://schema.org",
          "@type": "WebSite",
          "name": "UspAvalia",
          "url": "http://uspavalia.com/",
          "logo": "http://uspavalia.com/assets/images/poster.jpg",
          "potentialAction": {
            "@type": "SearchAction",
            "target": "http://uspavalia.com/?p=pesquisa&pesquisa={search_term_string}",
            "query-input": "required name=search_term_string"
          }
        }
    </script>

