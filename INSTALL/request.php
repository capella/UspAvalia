<?php
require __DIR__ . '/../vendor/autoload.php';
require __DIR__ . '/../helpers/connection.php';
require __DIR__ . '/../helpers/sanitizer.php';
use Sunra\PhpSimple\HtmlDomParser;

function startsWith($haystack, $needle){
     $length = strlen($needle);
     return (substr($haystack, 0, $length) === $needle);
}

$options = array(
    "ssl" => array(
        "verify_peer" => false,
        "verify_peer_name" => false,
    ),
);

$url_disciplina = 'https://uspdigital.usp.br/jupiterweb/obterTurma?sgldis=';
$json = ["ok"];

if (isset($_GET['id'])) {
    $id = GetSQLValueString($_GET['id'], "int");

    $sql = "
    SELECT * FROM disciplinas WHERE id =".$id;
    $disciplna_result = $connection->query($sql);
    $disciplna = $disciplna_result->fetch_assoc();

    $a = array();

    $data = file_get_contents($url_disciplina.$disciplna['codigo'], false, stream_context_create($options));
    $htmld = HtmlDomParser::str_get_html($data);

    if ($htmld == "") {
        echo "Erro:".$disciplna['id']."<br>";
        $insertSQL3 = "UPDATE  `disciplinas` SET  `roubo` =  '-1' WHERE  `disciplinas`.`id` =".GetSQLValueString($disciplna['id'],'int').";";
        $Result3 = $connection->query($insertSQL3) or die($connection->error);
    } else {
        foreach($htmld->find('table[cellspacing=1] tr[class="txt_verdana_8pt_gray"] td font[face="Verdana, Arial, Helvetica, sans-serif"]') as $element){
            $value =  str_replace('(R)','',trim(preg_replace("/\r|\n/", "", trim(preg_replace('!\s+!', ' ', $element->plaintext)))));
            if(strlen($value)>7&&$value!='Hor&aacute;rio'&&$value!='Horário'){
                if(!in_array($value, $a) && !startsWith($value, "Aulas") && !startsWith($value, "Atividade")){
                    $a[]=$value;
                }
            }
        }

        foreach ($a as &$value) {
            $insertSQL1 =   "INSERT INTO professores (nome, idunidade)
                            SELECT * FROM (SELECT ".GetSQLValueString($value,'text').", ".GetSQLValueString($disciplna['idunidade'],'int').") AS tmp
                            WHERE NOT EXISTS (
                                SELECT nome FROM professores WHERE nome = ".GetSQLValueString($value,'text')."
                            ) LIMIT 1;";
            $insertSQL2 = "INSERT IGNORE INTO aulaprofessor (idprofessor, idaula) SELECT id as idprofessor, ".GetSQLValueString($disciplna['id'],'int')." as idaula from professores WHERE nome = ".GetSQLValueString($value,'text').";";
            $Result1 = $connection->query($insertSQL1) or die($connection->error);
            $Result2 = $connection->query($insertSQL2) or die($connection->error);
            //echo $insertSQL1;
        }
        $insertSQL3 = "UPDATE  `disciplinas` SET  `roubo` =  '1' WHERE  `disciplinas`.`id` =".GetSQLValueString($disciplna['id'],'int').";";
        $Result3 = $connection->query($insertSQL3) or die($connection->error);
    
    }
    $htmld->clear(); 
    unset($htmld);

    if ($disciplna_result) $disciplna_result->close();
}

echo json_encode($json, JSON_UNESCAPED_UNICODE);