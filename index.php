<?php 
require __DIR__ . '/vendor/autoload.php';
require __DIR__ . '/helpers/connection.php'; 
require __DIR__ . '/helpers/sanitizer.php'; 
require __DIR__ . '/config.php';

date_default_timezone_set('America/Sao_Paulo');
session_start();

$fb = new Facebook\Facebook([
   'app_id' => $appId_facebook,
   'app_secret' => $secret_facebook,
   'default_graph_version' => 'v2.9',
]);

$user = null;

if (isset($_SESSION['fb_access_token'])) {
   try {
   // Proceed knowing you have a logged in user who's authenticated.
      $response = $fb->get('/me?fields=id,name', $_SESSION['fb_access_token']);
      $user_profile = $response->getGraphUser();
      $user = $user_profile['id'];
   } catch (FacebookApiException $e) {
      error_log($e);
      $user = null;
   }
} else {
   $helper = $fb->getRedirectLoginHelper();
   $permissions = ['email'];
   $loginUrl = $helper->getLoginUrl($url_full.'/?p=fb-callback&ant='.urlencode($_SERVER[REQUEST_URI]), $permissions);
}

$page = 'index';
if(isset($_GET['p'])&&$_GET['p']!=''){
   $page  = $_GET['p'];
}

// Configuracoes templates
$no_template_pages = array(
   'logout', 
   'votar', 
   'comentar', 
   'votarcomentario',
   'fb-callback',
   'search',
   'vote'
);
$header = 'view/template/header.php';
$footer = 'view/template/footer.php';

$template = !in_array($page, $no_template_pages);
if ($template) include($header);  
if(file_exists("view/" . $page . ".php")) {
  include('view/' . $page . '.php');
} else {
    include('view/404.php');
}
if ($template) include($footer);

if ($connection) $connection->close();
?>