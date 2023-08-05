<!DOCTYPE html>
<html lang="en">
   <head>
      <meta charset="utf-8">
      <meta http-equiv="X-UA-Compatible" content="IE=edge">
      <meta name="viewport" content="width=device-width, initial-scale=1">
      <meta name="description" content="">
      <meta name="author" content="">
      <!-- Favicons -->
      <link rel="apple-touch-icon" sizes="57x57" href="<?= $url_full; ?>/assets/favicons/apple-icon-57x57.png">
      <link rel="apple-touch-icon" sizes="60x60" href="<?= $url_full; ?>/assets/favicons/apple-icon-60x60.png">
      <link rel="apple-touch-icon" sizes="72x72" href="<?= $url_full; ?>/assets/favicons/apple-icon-72x72.png">
      <link rel="apple-touch-icon" sizes="76x76" href="<?= $url_full; ?>/assets/favicons/apple-icon-76x76.png">
      <link rel="apple-touch-icon" sizes="114x114" href="<?= $url_full; ?>/assets/favicons/apple-icon-114x114.png">
      <link rel="apple-touch-icon" sizes="120x120" href="<?= $url_full; ?>/assets/favicons/apple-icon-120x120.png">
      <link rel="apple-touch-icon" sizes="144x144" href="<?= $url_full; ?>/assets/favicons/apple-icon-144x144.png">
      <link rel="apple-touch-icon" sizes="152x152" href="<?= $url_full; ?>/assets/favicons/apple-icon-152x152.png">
      <link rel="apple-touch-icon" sizes="180x180" href="<?= $url_full; ?>/assets/favicons/apple-icon-180x180.png">
      <link rel="icon" type="image/png" sizes="192x192"  href="<?= $url_full; ?>/assets/favicons/android-icon-192x192.png">
      <link rel="icon" type="image/png" sizes="32x32" href="<?= $url_full; ?>/assets/favicons/favicon-32x32.png">
      <link rel="icon" type="image/png" sizes="96x96" href="<?= $url_full; ?>/assets/favicons/favicon-96x96.png">
      <link rel="icon" type="image/png" sizes="16x16" href="<?= $url_full; ?>/assets/favicons/favicon-16x16.png">
      <link rel="manifest" href="<?= $url_full; ?>/assets/favicons/manifest.json">
      <meta name="msapplication-TileColor" content="#101010">
      <meta name="msapplication-TileImage" content="<?= $url_full; ?>/assets/favicons/ms-icon-144x144.png">
      <meta name="theme-color" content="#101010">
      <!-- Bootstrap core CSS -->
      <link rel="stylesheet" type="text/css" href="<?= $url_full; ?>/assets/css/meu.css">
      <!-- Custom styles for this template -->
      <link rel="stylesheet" type="text/css" href="<?= $url_full; ?>/assets/css/cover.css">
      <!-- Bootstrap core JavaScript
      ================================================== -->
      <!-- Placed at the end of the document so the pages load faster -->
      <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.0/jquery.min.js"></script>
      <script type="text/javascript" src="<?= $url_full; ?>/assets/js/jquery.barrating.min.js"></script>
      <script>
         (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
         (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
         m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
         })(window,document,'script','//www.google-analytics.com/analytics.js','ga');
      
         ga('create', 'UA-49934159-1', 'uspavalia.com');
         ga('send', 'pageview');
      </script>
   </head>
   <body>
      <div class="navbar navbar-inverse navbar-fixed-top" role="navigation">
         <div class="container">
            <div class="navbar-header">
               <button type="button" class="navbar-toggle" data-toggle="collapse" data-target=".navbar-collapse">
                  <span class="sr-only">Toggle navigation</span>
                  <span class="icon-bar"></span>
                  <span class="icon-bar"></span>
                  <span class="icon-bar"></span>
               </button>
               <a class="navbar-brand" href="/">USP Avalia</a>
            </div>
            <div class="collapse navbar-collapse">
               <ul class="nav navbar-nav">
                  <li><a href="/">Pesquisar</a></li>
                  <li><a href="/email">Contato</a></li>
                  <li><a href="/destaques">Destaques</a></li>
                  <li><a href="/sobre">Sobre</a></li>
                  <li><a href="/matrusp">Matrusp</a></li>
                  <?php if ($user): ?>
                  <!-- <li><a href="/?p=add">Adicionar Disciplina/Professor</a></li> -->
                  <li><a href="/logout">Logout</a></li>
                  <?php else: ?>
                  <li><a href="<?=$loginUrl;?>">
                     <img src="/assets/images/facebook.png" style="height: 20px;">
                  </a></li>
                  <?php endif ?>
               </ul>
            </div><!--/.nav-collapse -->
         </div>
      </div>

      <div class="container">
