<title>Contato</title>
<?php
require __DIR__ . '/../config.php';
use PHPMailer\PHPMailer\PHPMailer;

if(isset($_POST['email']) && isset($_SESSION["k1"]) &&
     isset($_SESSION["k1"]) && $_SESSION["k1"] + $_SESSION["k2"] == $_POST['soma']) {

  // EDIT THE 2 LINES BELOW AS REQUIRED
  $email_to = "contato@uspavalia.com";
  $email_subject = "USP AVALIA -  CONTATO";

  function died($error) {
      // your error code can go here
      echo "We are very sorry, but there were error(s) found with the form you submitted. ";
      echo "These errors appear below.<br /><br />";
      echo $error."<br /><br />";
      echo "Please go back and fix these errors.<br /><br />";
      die();
  }

  // validation expected data exists
  if(!(isset($_POST['first_name'])&&isset($_POST['last_name'])&&isset($_POST['email'])&&isset($_POST['comments']))) {
      died('We are sorry, but there appears to be a problem with the form you submitted.');       
  }

  $first_name = $_POST['first_name']; // required 
  $last_name = $_POST['last_name']; // required
  $email_from = $_POST['email']; // required
  $telephone = $_POST['telephone']; // not required
  $comments = $_POST['comments']; // required

  $error_message = "";

  $email_exp = '/^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,4}$/';
  if(!preg_match($email_exp,$email_from)) {
    $error_message .= 'The Email Address you entered does not appear to be valid.<br />';
  }

  $string_exp = "/^[A-Za-z .'-]+$/";

  if(!preg_match($string_exp,$first_name)) {
    $error_message .= 'The First Name you entered does not appear to be valid.<br />';
  }

  if(!preg_match($string_exp,$last_name)) {
    $error_message .= 'The Last Name you entered does not appear to be valid.<br />';
  }

  if(strlen($comments) < 2) {
    $error_message .= 'The Comments you entered do not appear to be valid.<br />';
  }

  if(strlen($error_message) > 0) {
    died($error_message);
  }

  $email_message = "Form details below:\n\n";

  $email_message .= "First Name: ".$first_name."\n";
  $email_message .= "Last Name: ".$last_name."\n";
  $email_message .= "Email: ".$email_from."\n";
  $email_message .= "Comments: ".$comments."\n";
     
  // Settings
  $mail = new PHPMailer();
  $mail->IsSMTP();
  $mail->CharSet = 'UTF-8';

  $mail->Host       = $smtp_host;
  $mail->SMTPDebug  = 0;                     // enables SMTP debug information (for testing)
  $mail->SMTPAuth   = true;                  // enable SMTP authentication
  $mail->Port       = 465;                   // set the SMTP port for the GMAIL server
  $mail->Username   = $smtp_username;         // SMTP account username example
  $mail->Password   = $smtp_password;        // SMTP account password example
  $mail->addAddress($smtp_destination, 'UspAvalia');     // Add a recipient


  // Content
  $mail->isHTML(false);
  $mail->Subject = 'USPAVALIA - Contato';
  $mail->Body    = $email_message;
  $mail->send();

?>

<br><br>
<h4>Email enviado com sucesso.<?= $error_message ?></h4>


<?php

} else {

$_SESSION["k1"] = rand(0,20);
$_SESSION["k2"] = rand(0,20);

?>
<form name="contactform" method="post" action="/email">
  <h3>Contato</h3>
  <hr>
  <p> Aqui você pode solicitar que sua avaliação ou nome seja retirado. Este também é o espaço para dar sugestões ou expressar sua opinião. Se não quiser usar o formulário `contato arroba uspavalia.com` </p>
  <br />
  <table style="width:100%">
    <tr>
      <td width="33%" valign="top">
        <label for="first_name">Nome *</label>
      </td>
      <td width="67%" valign="top">
        <input type="text" name="first_name" maxlength="50" size="30" class="form-control">
        <br> </td>
    </tr>
    <tr>
      <td valign="top">
        <label for="last_name">Sobrenome *</label>
      </td>
      <td valign="top">
        <input type="text" name="last_name" maxlength="50" size="30" class="form-control">
        <br> </td>
    </tr>
    <tr>
      <td valign="top">
        <label for="email">Endereço de Email*</label>
      </td>
      <td valign="top">
        <input type="text" name="email" maxlength="80" size="30" class="form-control">
        <br> </td>
    </tr>
    <tr>
      <td width="33%" valign="top">
        <label for="first_name">Quanto é <?= $_SESSION["k1"] ?> + <?= $_SESSION["k2"] ?>?</label>
      </td>
      <td width="67%" valign="top">
        <input type="number" name="soma" maxlength="50" size="30" class="form-control">
        <br> </td>
    </tr>
    <tr>
      <td valign="top">
        <label for="comments">Mensagem *</label>
      </td>
      <td valign="top">
        <textarea name="comments" maxlength="1000" cols="25" rows="6" class="form-control btn-lg"></textarea>
      </td>
    </tr>
    <tr>
      <td colspan="2" style="text-align:center">
        <br>
        <input type="submit" class="btn btn-default" value="Enviar"> </td>
    </tr>
  </table>
</form>
<?
}
?>
