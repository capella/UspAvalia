<?php require_once('Connections/CapellaResumo.php'); 
require 'view/facebook.php';
require 'config.php';
date_default_timezone_set('America/Sao_Paulo');


session_start();

$facebook = new Facebook(array(
	'appId' => $appId_facebook,
	'secret' => $secret_facebook
));
$url = "uspavalia.com";

$user = $facebook->getUser();

if ($user) {
  try {
    // Proceed knowing you have a logged in user who's authenticated.
    $user_profile = $facebook->api('/me');
  } catch (FacebookApiException $e) {
    error_log($e);
    $user = null;
  }
}

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
	  }
include('view/template/footer.php');
?>
