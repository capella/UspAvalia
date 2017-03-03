<?
$page = 'index';
if(isset($_GET['p'])&&$_GET['p']!=''){
		$page  = $_GET['p'];
}
 include('view/template/header.php');    
	  if(file_exists("view/" . $page . ".php")) {
				include('view/' . $page . '.php');
	  }
	  else {
				  include('view/404.php');
				  echo 'Página não encontrada.';
	  }
include('view/template/footer.php');
?>
