<?php
   $json = [];
   if (isset($_POST['query'])) {
      $data = GetSQLValueString($_POST['query'], "text2");
      $sql = "
      SELECT * FROM (
         (SELECT id, nome, 0 as type FROM professores
         WHERE MATCH(nome) AGAINST('*".$data."*' IN BOOLEAN MODE)
         ORDER BY MATCH(nome) AGAINST('*".$data."*' IN BOOLEAN MODE) DESC
         LIMIT 10)
         UNION
         (SELECT id, CONCAT(codigo, ' - ', nome) as nome, 1 as type from disciplinas
         WHERE MATCH(nome, codigo) AGAINST('*".$data."*' IN BOOLEAN MODE)
         ORDER BY MATCH(nome, codigo) AGAINST('*".$data."*' IN BOOLEAN MODE) DESC
         LIMIT 10)
      ) NAMES
      LIMIT 10"; 

      $result = $connection->query($sql);
      
      while($row = $result->fetch_assoc()){
         $json[] = array(
            "name" => $row['nome'],
            "type" => $row['type'],
            "id" => $row['id'],
         );
      }
      $result->close();
   }
   echo json_encode($json, JSON_UNESCAPED_UNICODE);
?>