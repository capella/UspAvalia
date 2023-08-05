<?php
require __DIR__ . '/../vendor/autoload.php';
require __DIR__ . '/../helpers/connection.php';
require __DIR__ . '/../helpers/sanitizer.php';
use voku\helper\HtmlDomParser;

$json = ["ok"];

if (isset($_GET['id'])) {
    $id = GetSQLValueString($_GET['id'], "int");

    $sql = "SELECT * FROM disciplinas WHERE id = ".$id;
    $result = $connection->query($sql);

    if($result && $result->num_rows == 0) {
        return;
    } else {
        $disciplina = mysqli_fetch_assoc($result);
    }

    $data = file_get_contents("../matrusp/db/".$disciplina['codigo'].".json");
    $json_input = json_decode($data, true);

    $professores = [];
    foreach ($json_input['turmas'] as $turma) {
        if ($turma['horario'] == null) {
            continue;
        }
        foreach ($turma['horario'] as $horario) {
            foreach ($horario['professores'] as $professor) {
                $professores[] = trim(preg_replace('/\(.*\)/','',$professor));
            }
        }
    }
    $professores = array_filter(array_unique($professores));
    foreach ($professores as $professor) {
        $insertSQL1 =   "INSERT INTO professores (nome, idunidade)
                        SELECT * FROM (SELECT ".GetSQLValueString($professor,'text').", ".GetSQLValueString($disciplina['idunidade'],'int').") AS tmp
                        WHERE NOT EXISTS (
                            SELECT nome FROM professores WHERE nome = ".GetSQLValueString($professor,'text')."
                        ) LIMIT 1;";
        $insertSQL2 = "INSERT IGNORE INTO aulaprofessor (idprofessor, idaula) SELECT id as idprofessor, ".GetSQLValueString($disciplina['id'],'int')." as idaula from professores WHERE nome = ".GetSQLValueString($professor,'text').";";
        $Result1 = $connection->query($insertSQL1) or die($connection->error);
        $Result2 = $connection->query($insertSQL2) or die($connection->error);
    }
    $insertSQL3 = "UPDATE  `disciplinas` SET  `roubo` =  '1' WHERE  `disciplinas`.`id` =".GetSQLValueString($disciplina['id'],'int').";";
    $Result3 = $connection->query($insertSQL3) or die($connection->error);

    if ($disciplna_result) $disciplna_result->close();
}

echo json_encode($json, JSON_UNESCAPED_UNICODE);
