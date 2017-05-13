<?php
    session_start();
    unset($_SESSION['id']);
    unset($_SESSION['username']);
    unset($_SESSION['oauth_provider']);
    session_destroy();
    header("location: ".$url_full; ?."/");
?>