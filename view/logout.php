<?php

session_start();
unset($_SESSION['fb_access_token']);
session_destroy();
header("location: /");
