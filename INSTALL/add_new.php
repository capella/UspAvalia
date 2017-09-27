<?php
require __DIR__ . '/../vendor/autoload.php';
require __DIR__ . '/../helpers/connection.php';
require __DIR__ . '/../helpers/sanitizer.php';
use Sunra\PhpSimple\HtmlDomParser;

set_time_limit (1000000);

// Configuracoes
$url_disciplina = 'https://uspdigital.usp.br/jupiterweb/obterTurma?sgldis=';

$json = ["ok"];

$string = file_get_contents("db_usp.txt");
$json_input = json_decode($string, true);

foreach ($json_input['TODOS'] as $val) {
    $codigo = GetSQLValueString($val[0], "text");
    $nome = GetSQLValueString($val[1], "text");
    $sql =   "SELECT codigo FROM disciplinas WHERE codigo = ".$codigo;

    $result = $connection->query($sql);
    if($result && $result->num_rows == 0){

        $html = HtmlDomParser::file_get_html($url_disciplina.$val[0]);
        $disciplina = $html->find('td b font[face="Verdana, Arial, Helvetica, sans-serif"] span[class="txt_arial_10pt_black"]');
        $unidade = $disciplina[0]->plaintext;
        if (!isset($unidade) || $unidade == "") continue;
        $unidade = html_entity_decode ($unidade);
        $unidade = GetSQLValueString(trim($unidade), "text");

        $sql = "
            INSERT INTO disciplinas (nome, codigo, idunidade) 
            VALUES (".$nome.",".$codigo.", (
                SELECT id FROM unidades WHERE NOME = ".$unidade." LIMIT 1
            ))
        ";

        $insert_result = $connection->query($sql);
        // if (!$insert_result) {
        //     $json = array('error' => $connection->error.$val[0]);
        //     break;
        // }
        $html->clear(); 
        unset($html);
    }
    if ($result) {
        $result->close();
    } else {
        $json = array('error' => $connection->error);
        break;
    }
}

echo json_encode($json, JSON_UNESCAPED_UNICODE);
?>