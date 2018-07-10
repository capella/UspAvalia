<?php
require __DIR__ . '/../vendor/autoload.php';
require __DIR__ . '/../helpers/connection.php';
require __DIR__ . '/../helpers/sanitizer.php';
use Sunra\PhpSimple\HtmlDomParser;

set_time_limit (1000000);

// Configuracoes
$url_disciplina = 'https://uspdigital.usp.br/jupiterweb/obterTurma?sgldis=';

$json = ["ok"];

$options = array(
    "ssl" => array(
        "verify_peer" => false,
        "verify_peer_name" => false,
    ),
);

$string = file_get_contents("db_usp.txt");
$json_input = json_decode($string, true);

$count = 0;

foreach ($json_input['TODOS'] as $val) {
    $codigo = GetSQLValueString($val[0], "text");
    $nome = GetSQLValueString($val[1], "text");
    $sql =   "SELECT codigo FROM disciplinas WHERE codigo = ".$codigo;

    $result = $connection->query($sql);
    if($result && $result->num_rows == 0){

        $data = file_get_contents($url_disciplina.$val[0], false, stream_context_create($options));
        $html = HtmlDomParser::str_get_html($data);
        $disciplina = $html->find('td b font[face="Verdana, Arial, Helvetica, sans-serif"] span[class="txt_arial_10pt_black"]');
        $unidade = $disciplina[0]->plaintext;
        if (!isset($unidade) || $unidade == "") continue;
        $unidade = html_entity_decode ($unidade);
        $unidade = GetSQLValueString(trim($unidade), "text");

        // check for unidade
        $sql =   "SELECT id FROM unidades WHERE NOME = ".$unidade;
        $result = $connection->query($sql);
        $unidadeID = NULL;
        if($result && $result->num_rows == 0) {
            $sql =   "INSERT INTO unidades (NOME) VALUES (".$unidade.")";
            $connection->query($sql);
            $unidadeID = $connection->insert_id;
        } else {
            $row = mysqli_fetch_assoc($result);
            $unidadeID = $row['id'];
        }
        if (!$result) {
            $json = array('error' => $connection->error.$val[0]);
            break;
        }

        // Insert new discipline
        $sql = "
            INSERT INTO disciplinas (nome, codigo, idunidade) 
            VALUES (".$nome.",".$codigo.", ".$unidadeID.")
        ";

        $insert_result = $connection->query($sql);
        if (!$insert_result) {
            $json = array('error' => $connection->error.$val[0]);
            break;
        }
        $html->clear(); 
        unset($html);

        // tell for next
        $progress = $count/sizeof($json_input['TODOS']);
        $json = array("continue" => $progress);
        break;
    }
    if ($result) {
        $result->close();
    } else {
        $json = array('error' => $connection->error);
        break;
    }
    $count += 1;
}

echo json_encode($json, JSON_UNESCAPED_UNICODE);
?>