<?php

require 'facebook.php';
require '../config.php';

$facebook = new Facebook(array(
    'appId' => $appId_facebook,
    'secret' => $secret_facebook,
    'cookie' => true
));
$session = $facebook->getSession();

if (!empty($session)) {
    try {
        $uid = $facebook->getUser();
        $user = $facebook->api('/me');
    } catch (Exception $e) {
    }
    if (!empty($user)) {
        echo '<pre>';
        print_r($user);
    } else {
        die("There was an error.");
    }
} else {
    $login_url = $facebook->getLoginUrl();
    header("Location: " . $login_url);
}
