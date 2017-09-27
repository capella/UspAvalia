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
      <title>USP Avalia</title>
   </head>
   <body>
        <div class="container">
            <h1>Instalação</h1>
            <hr>
            <table class="table table-hover table-bordered">
                <tr>
                    <td id="create_db" style="text-align: center;">
                       <span class="glyphicon glyphicon-minus" aria-hidden="true"></span>
                    </td>
                    <td>Criando DB</td>
                </tr>
                <tr>
                    <td id="copy_data" style="text-align: center;">
                       <span class="glyphicon glyphicon-minus" aria-hidden="true"></span>
                    </td>
                    <td>Copiando Arquivo de Matérias</td>
                </tr>
                <tr>
                    <td id="add_new" style="text-align: center;">
                       <span class="glyphicon glyphicon-minus" aria-hidden="true"></span>
                    </td>
                    <td>Verificando Matérias Novas</td>
                </tr>
                <tr>
                    <td id="add_teacher" style="text-align: center;">
                       <span class="glyphicon glyphicon-minus" aria-hidden="true"></span>
                    </td>
                    <td>Buscando Professores <span id="por"></span></td>
                </tr>
            </table>
         </div><!-- /.container -->
      <script src="<?= $url_full; ?>/assets/js/bootstrap.min.js"></script>
      <script>
         function mark_done (d) {
            d.html('<span class="glyphicon glyphicon-ok" aria-hidden="true"></span>');
         }
         function mark_doing (d) {
            d.html('<div class="progress"> <div class="progress-bar progress-bar-striped active" role="progressbar" style="width: 45%"> <span class="sr-only">45% Complete</span> </div> </div>');            
         }
         function mark_error (d) {
            d.html('<span class="glyphicon glyphicon-remove" aria-hidden="true"></span>');            
         }
         function mark_progress (d, done, max) {
            d.html('<div class="progress"> <div class="progress-bar progress-bar-striped active" role="progressbar"  style="width: '+(done/max*100)+'%"> <span class="sr-only">'+done+'/'+max+'</span> </div> </div>');            
         }

         function create_db () {
            mark_doing ($("#create_db"));
            $.get( "create_db.php", function( data ) {
               if (data == "ok") {
                  mark_done ($("#create_db"));
                  copy_data ();
               } else {
                  mark_error ($("#create_db"));
               }
            }, "json").fail(function() {
               mark_error ($("#create_db"));
            });
         }

         function copy_data () {
            mark_doing ($("#copy_data"));
            $.get( "copy_data.php", function( data ) {
               if (data == "ok") {
                  mark_done ($("#copy_data"));
                  add_new();
                  add_teacher ();
               } else {
                  mark_error ($("#copy_data"));
               }
            }, "json").fail(function() {
               mark_error ($("#copy_data"));
            });
         }

         function add_new () {
            mark_doing ($("#add_new"));
            $.get( "add_new.php", function( data ) {
               if (data == "ok") {
                  mark_done ($("#add_new"));
               } else {
                  mark_error ($("#add_new"));
               }
            }, "json").fail(function() {
               mark_error ($("#add_new"));
            });
         }
         function add_teacher () {
            mark_doing ($("#add_teacher"));
            $.get( "list.php", function( data ) {
               request_one(data, 0);
            }, "json").fail(function() {
               mark_error ($("#add_teacher"));
            });
         }

         function request_one (array, id) {
            if (id >= array.length) {
               mark_done ($("#add_teacher"));
            } else {
               $.get( "request.php?id="+array[id], function( data ) {
                  if (data == "ok") {
                     mark_progress ($("#add_teacher"), id+1, array.length);
                     $("#por").text((id+1)+'/'+array.length);
                     request_one(array, id+1);
                  } else {
                     mark_error ($("#add_teacher"));
                  }
               }, "json").fail(function() {
                  mark_error ($("#add_teacher"));
               });
            }
         }

         create_db ();
      </script>
   </body>
</html>
