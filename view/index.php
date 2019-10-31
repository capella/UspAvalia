<?php

function getNumber($conn, $query, $field, $cachename) {
   if (!is_file($cachename) || filemtime($cachename) < time()-1*300) {
      $result = $conn->query($query);
      $number = 0;
      if($result) {
         $result_row = $result->fetch_assoc();
         $number = $result_row[$field];
         $result->close();
      }
      file_put_contents($cachename, serialize($number));
      return $number;
   } else {
     return unserialize(file_get_contents($cachename));
   }
}


function media ($conn) {
   $query = "SELECT AVG(nota) as m FROM votos WHERE tipo <> 5;";
   $n = getNumber($conn, $query, 'm', 'media.cache');
   return number_format($n*2, 2, ',', ' ');
}

function avaliacoes ($conn) {
   $query = "SELECT COUNT(*) as m FROM votos;";
   $n = getNumber($conn, $query, 'm', 'avaliacoes.cache');
   return number_format($n, 0, ',', ' ');
}

function pessoas ($conn) {
   $query = "SELECT count(*) as m FROM (
               SELECT iduso FROM votos GROUP BY iduso
            ) as S;";
   $n = getNumber($conn, $query, 'm', 'pessoas.cache');
   return number_format($n, 0, ',', ' ');
}

?>
<div class="starter-template">
   <h1>USP Avalia</h1>
   <p class="lead">Pesquise uma disciplina, um professor ou uma sigla.
      <br /><small style="font-size:10px;"> Depois avalie e ajude a comunidade.</small>
   </p>
   <p class="lead">
      <form method="get" action="/">
      <div class="input-group">
         <input type="text" class="form-control typeahead" name="pesquisa" autofocus autocomplete="off">
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
               <h3><?= media($connection); ?></h3>
               <p>Nota média da <br> universidade</p>
            </div>
            <a href="/" class="small-box-footer"></a>
         </div>
      </div><!-- ./col -->

      <div class="col-lg-4 col-xs-6">
         <!-- small box -->
         <div class="small-box bg-green">
            <div class="inner">
               <h3><?= avaliacoes($connection); ?></h3>
               <p>Avaliações realizadas</p><br>
            </div>
            <a href="/" class="small-box-footer"></a>
         </div>
      </div><!-- ./col -->

      <div class="col-lg-4 col-xs-6">
         <!-- small box -->
         <div class="small-box bg-yellow">
            <div class="inner">
               <h3><?= pessoas($connection); ?></h3>
               <p>Pessoas já ajudaram</p><br>
            </div>
            <a href="/" class="small-box-footer"></a>
         </div>
      </div><!-- ./col -->

   </div><!-- /.row -->
   <p>Ajude a divulgar o USP Avalia em sua unidade com <a href="<?= $url_full; ?>/assets/images/poster.pdf" onclick="ga('send', 'event', 'Poster', 'click', 'pagina_inicial');">esse poster</a>.</p>

</div><!-- /.row -->

<script type="application/ld+json">
   {
      "@context": "http://schema.org",
      "@type": "Organization",
      "url": "<?= $url_full; ?>/",
      "logo": "<?= $url_full; ?>/assets/images/poster.jpg"
   }
</script>
<script type="application/ld+json">
   {
      "@context": "http://schema.org",
      "@type": "WebSite",
      "name": "UspAvalia",
      "url": "<?= $url_full; ?>/",
      "potentialAction": {
         "@type": "SearchAction",
         "target": "<?= $url_full; ?>/?p=pesquisa&pesquisa={search_term_string}",
         "query-input": "required name=search_term_string"
      }
   }
</script>
