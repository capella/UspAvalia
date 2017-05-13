<title>Contato</title>
<?php



if(isset($_POST['email']) && isset($_SESSION["k1"]) && isset($_SESSION["k1"]) && $_SESSION["k1"] + $_SESSION["k2"] == $_POST['soma']) {
 
     
 
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
 
    $email_message = "Form details below.\n\n";
 
     
 
    function clean_string($string) {
 
      $bad = array("content-type","bcc:","to:","cc:","href");
 
      return str_replace($bad,"",$string);
 
    }
 
     
 
    $email_message .= "First Name: ".clean_string($first_name)."\n";
 
    $email_message .= "Last Name: ".clean_string($last_name)."\n";
 
    $email_message .= "Email: ".clean_string($email_from)."\n";
 
    $email_message .= "Telephone: ".clean_string($telephone)."\n";
 
    $email_message .= "Comments: ".clean_string($comments)."\n";
 
     
 
     
 
// create email headers
 
$headers = 'From: '.$email_from."\r\n".
 
'Reply-To: '.$email_from."\r\n" .
 
'X-Mailer: PHP/' . phpversion();
 
@mail($email_to, $email_subject, $email_message, $headers);  
 
?>

 
 
 <br><br>
 

<h4>Email enviado com sucesso.</h4>
 
 
 
<?php
 
} else {

$_SESSION["k1"] = rand(0,20);
$_SESSION["k2"] = rand(0,20);

 
?>
<form name="contactform" method="post" action="/email">

<h3>Contato</h3><hr>
 <p>
 Aqui você pode solicitar  que sua avaliação ou  nome seja retirado. Este  também é o espaço para dar sugestões ou expressar sua opinião.
 </p><br />
<table style="width:100%">
 
<tr>
 
 <td width="33%" valign="top">
 
  <label for="first_name">Nome *</label>
 
 </td>
 
 <td width="67%" valign="top">
 
  <input  type="text" name="first_name" maxlength="50" size="30" class="form-control"><br>
 
 </td>
 
</tr>
 
<tr>
 
 <td valign="top">
 
  <label for="last_name">Sobrenome *</label>
 
 </td>
 
 <td valign="top">
 
  <input  type="text" name="last_name" maxlength="50" size="30" class="form-control"><br>
 
 </td>
 
</tr>
 
<tr>
 
 <td valign="top">
 
  <label for="email">Endereço de Email*</label>
 
 </td>
 
 <td valign="top">
 
     <input  type="text" name="email" maxlength="80" size="30" class="form-control"><br>
 
 </td>
 
</tr> 

<tr>
 
 <td width="33%" valign="top">
 
  <label for="first_name">Quanto é <?= $_SESSION["k1"] ?> + <?= $_SESSION["k2"] ?>?</label>
 
 </td>
 
 <td width="67%" valign="top">
 
  <input  type="number" name="soma" maxlength="50" size="30" class="form-control"><br>
 
 </td>
 
</tr>
 
<tr>
  
  <td valign="top">
    
    <label for="comments">Mensagem *</label></td>
  
  <td valign="top">
    
    <textarea  name="comments" maxlength="1000" cols="25" rows="6" class="form-control btn-lg"></textarea>
    
    </td>
  
</tr>
 
<tr>
 
 <td colspan="2" style="text-align:center">
 
  <br>
  <input type="submit" class="btn btn-default" value="Enviar">  
 
 </td>
 
</tr>
 
</table>
 
</form>
<?php
 
} 
 
?>