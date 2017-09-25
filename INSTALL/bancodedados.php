<?php 
require __DIR__ . '/../vendor/autoload.php';
require __DIR__ . '/../helpers/connection.php';
require __DIR__ . '/../helpers/sanitizer.php';
use Sunra\PhpSimple\HtmlDomParser;

function startsWith($haystack, $needle)
{
     $length = strlen($needle);
     return (substr($haystack, 0, $length) === $needle);
}


$hostname_connection = $hostname;
$database_connection = $database;
$username_connection = $username;
$password_connection = $password;
$connection = mysql_pconnect($hostname_connection, $username_connection, $password_connection) or trigger_error(mysql_error(),E_USER_ERROR);
mysql_set_charset('utf8',$connection);


mysql_select_db($database_connection, $connection);

 set_time_limit (1000000);


$editFormAction = $_SERVER['PHP_SELF'];
if (isset($_SERVER['QUERY_STRING'])) {
  $editFormAction .= "?" . htmlentities($_SERVER['QUERY_STRING']);
}

if ((isset($_POST["MM_insert"])) && ($_POST["MM_insert"] == "form1")) {
  $insertSQL = sprintf("INSERT INTO unidades (id, NOME) VALUES (%s, %s)",
                       GetSQLValueString($_POST['id'], "int"),
                       GetSQLValueString($_POST['NOME'], "text"));

  mysql_select_db($database_connection, $connection);
  $Result1 = mysql_query($insertSQL, $connection) or die(mysql_error());
}

mysql_select_db($database_connection, $connection);
$query_Disciplnas = "SELECT * FROM disciplinas WHERE roubo = 0 ORDER BY id ASC LIMIT 0 , 10";
$Disciplnas = mysql_query($query_Disciplnas, $connection) or die(mysql_error());
$row_Disciplnas = mysql_fetch_assoc($Disciplnas);
$totalRows_Disciplnas = mysql_num_rows($Disciplnas);


// Find all images 
/*
$html = file_get_html('https://uspdigital.usp.br/jupiterweb/jupTurmaHorarioBusca');

foreach($html->find('select[name=colegiado]  option') as $element){
	   if($element->value != 0){
       		echo $element->value.'-'.$element->plaintext. '<br>';
			$insertSQL = sprintf("INSERT INTO unidades (id, NOME) VALUES (%s, %s)",
                       GetSQLValueString($element->value, "int"),
                       GetSQLValueString($element->plaintext, "text"));

  mysql_select_db($database_connection, $connection);
  
  $Result1 = mysql_query($insertSQL, $connection) or die(mysql_error());
	   }
}
---------------------------
$html1 = file_get_html('https://uspdigital.usp.br/jupiterweb/jupTurmaHorarioBusca');

foreach($html1->find('select[name=colegiado]  option') as $element1){
	   if($element1->value != 0){

$html = file_get_html('https://uspdigital.usp.br/jupiterweb/jupDisciplinaLista?codcg='.$element1->value.'&letra=A-Z&tipo=T');

foreach($html->find('TABLE[align="center"] TR') as $element){
	$disciplina = $element->find('span[class="txt_arial_8pt_gray"]');
	if(isset($disciplina[0]->plaintext)&&$disciplina[0]->plaintext!=''){
	       		//echo str_replace(' ', '',GetSQLValueString($disciplina[0]->plaintext, "text")).'-'.GetSQLValueString($disciplina[1]->plaintext, "text"). '<br>';
				
			$insertSQL = sprintf("INSERT INTO disciplinas (nome, codigo, idunidade) VALUES (%s, %s, %s)",
                       GetSQLValueString($disciplina[1]->plaintext, "text"),
                       str_replace(' ', '',GetSQLValueString($disciplina[0]->plaintext, "text")),
					    GetSQLValueString($element1->value, "int"));
						echo $insertSQL.'<br>';

  mysql_select_db($database_connection, $connection);
  
  $Result1 = mysql_query($insertSQL, $connection) or die(mysql_error());
	}

}
}
}

*/
if($totalRows_Disciplnas>0){
do { 
	$a=array();
	$htmld = HtmlDomParser::file_get_html('https://uspdigital.usp.br/jupiterweb/obterTurma?sgldis='.$row_Disciplnas['codigo']);
	//echo 'https://uspdigital.usp.br/jupiterweb/obterTurma?sgldis='.$row_Disciplnas['codigo'];
	if ($htmld == "") {
		echo "Erro:".$row_Disciplnas['id']."<br>";
		$insertSQL3 = "UPDATE  `disciplinas` SET  `roubo` =  '-1' WHERE  `disciplinas`.`id` =".GetSQLValueString($row_Disciplnas['id'],'int').";";
		$Result3 = mysql_query($insertSQL3, $connection) or die(mysql_error());
		continue;
	}
	foreach($htmld->find('table[cellspacing=1] tr[class="txt_verdana_8pt_gray"] td font[face="Verdana, Arial, Helvetica, sans-serif"]') as $element){
		$value =  str_replace('(R)','',trim(preg_replace("/\r|\n/", "", trim(preg_replace('!\s+!', ' ', $element->plaintext)))));
		if(strlen($value)>7&&$value!='Hor&aacute;rio'&&$value!='Horário'){
			if(!in_array($value, $a) && !startsWith($value, "Aulas") && !startsWith($value, "Atividade")){
				$a[]=$value;
			}
		}
	}
	
	echo $row_Disciplnas['id']."-".$row_Disciplnas['codigo']."-";
	print_r($a);
	foreach ($a as &$value) {
		$insertSQL1 = 	"INSERT INTO professores (nome, idunidade)
						SELECT * FROM (SELECT ".GetSQLValueString($value,'text').", ".GetSQLValueString($row_Disciplnas['idunidade'],'int').") AS tmp
						WHERE NOT EXISTS (
							SELECT nome FROM professores WHERE nome = ".GetSQLValueString($value,'text')."
						) LIMIT 1;";
		$insertSQL2 = "INSERT IGNORE INTO aulaprofessor (idprofessor, idaula) SELECT id as idprofessor, ".GetSQLValueString($row_Disciplnas['id'],'int')." as idaula from professores WHERE nome = ".GetSQLValueString($value,'text').";";
  		$Result1 = mysql_query($insertSQL1, $connection) or die(mysql_error());
		$Result2 = mysql_query($insertSQL2, $connection) or die(mysql_error());
		//echo $insertSQL1;
	}
	$insertSQL3 = "UPDATE  `disciplinas` SET  `roubo` =  '1' WHERE  `disciplinas`.`id` =".GetSQLValueString($row_Disciplnas['id'],'int').";";
	$Result3 = mysql_query($insertSQL3, $connection) or die(mysql_error());
	echo '<br><br>';
	
	$htmld->clear(); 
	unset($htmld);
	
}while ($row_Disciplnas = mysql_fetch_assoc($Disciplnas));
mysql_free_result($Disciplnas);

?>
<script> window.onload = function () {window.location.reload()} </script>
<?php
} else {
echo "Finalizado!";
}
?>