<?php
require __DIR__ . '/../vendor/autoload.php';
require __DIR__ . '/../helpers/connection.php';
require __DIR__ . '/../helpers/sanitizer.php';
use voku\helper\HtmlDomParser;

set_time_limit (1000000);

$json = ["ok"];

$string = file_get_contents("../matrusp/db/db.json");
$json_input = json_decode($string, true);

$count = 0;
$count_inserts = 0;

foreach ($json_input as $val) {
    $codigo = GetSQLValueString($val['codigo'], "text");
    $nome = GetSQLValueString($val['nome'], "text");
    $sql =   "SELECT codigo FROM disciplinas WHERE codigo = ".$codigo;

    $result = $connection->query($sql);
    if($result && $result->num_rows == 0){
        $unidade = $val['unidade'];
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

        // tell for next
        $progress = $count/sizeof($json_input);
        $json = array("continue" => $progress);
        if ($count_inserts >= 250) {
            break;
        } else {
            $count_inserts += 1;
        }
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
