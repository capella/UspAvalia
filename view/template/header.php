<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="description" content="">
    <meta name="author" content="">
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
            <?php if ($user): ?>
            <li><a href="/?p=add">Adicionar Disciplina/Professor</a></li>
            <li><a href="/logout">Logout</a></li>
            <? else: ?>
            <li><a href="<?=$loginUrl;?>">Login</a></li>
            <?php endif ?>
          </ul>
        </div><!--/.nav-collapse -->
      </div>
    </div>

    <div class="container">