<?php
require_once('../Connections/CapellaResumo.php'); 
include('../Connections/simple_html_dom.php');
set_time_limit (1000000);

$templine = '';
// Read in entire file
$lines = file("struct.sql");
// Loop through each line
foreach ($lines as $line){
	// Skip it if it's a comment
	if (substr($line, 0, 2) == '--' || $line == '')
	    continue;
	// Add this line to the current segment
	$templine .= $line;
	// If it has a semicolon at the end, it's the end of the query
	if (substr(trim($line), -1, 1) == ';')
	{
	    // Perform the query
	    mysql_query($templine) or print('Error performing query \'<strong>' . $templine . '\': ' . mysql_error() . '<br /><br />');
	    // Reset temp variable to empty
	    $templine = '';
	}
}
echo "Construção do banco de dados (se já não foi feita): OK";

file_put_contents("db_usp.txt", fopen("http://bcc.ime.usp.br/matrusp/db/db_usp.txt", 'r'));

// $string = file_get_contents("db_usp.txt");
// $json=json_decode($string,true);
// mysql_select_db($database_CapellaResumo, $CapellaResumo);
// //print_r($json);
// foreach ($json['TODOS'] as $val) {
// 	$insertSQL1 = 	"SELECT codigo FROM disciplinas WHERE codigo = '".$val[0]."';";
// 	$Result1 = mysql_query($insertSQL1, $CapellaResumo) or die(mysql_error());
// 	$t =  mysql_num_rows ($Result1);
// 	$h ="";
// 	if($t==0){
// 		echo $val[0];
// 		$html = file_get_html('https://uspdigital.usp.br/jupiterweb/obterTurma?sgldis='.$val[0]);
// 		$disciplina = $html->find('td b font[face="Verdana, Arial, Helvetica, sans-serif"] span[class="txt_arial_10pt_black"]');
// 		echo $disciplina[0]->plaintext;
// 		$insertSQL2 = 	'INSERT INTO capeocom_uspavalia.disciplinas (nome, codigo, idunidade) VALUES ("'.$val[1].'","'.$val[0].'", (SELECT id FROM unidades WHERE NOME = "'.trim(GetSQLValueString($disciplina[0]->plaintext, "text2")).'" LIMIT 1));';
// 		echo $insertSQL2;
// 		echo "<br>";
// 		mysql_query($insertSQL2, $CapellaResumo);
// 	}
// }
?>