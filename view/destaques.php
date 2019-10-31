<?php

$sql = "SELECT * FROM Melhores LIMIT 10";
$melhores_disciplnas_result = $connection->query($sql) or die($connection->error);

$sql = "
  SELECT
    professores.nome,
    SubQuery.expr1 * 2 AS nota,
    SubQuery.expr2 AS votos,
    unidades.NOME AS unidade,
    professores.id as id
  FROM (
    SELECT
      aulaprofessor.idprofessor,
      AVG(votos.nota) AS expr1,
      COUNT(*) AS expr2
    FROM votos
    INNER JOIN aulaprofessor
    ON votos.APid = aulaprofessor.id
    WHERE votos.tipo <> 5
    GROUP BY aulaprofessor.idprofessor) SubQuery
  INNER JOIN professores
  ON SubQuery.idprofessor = professores.id
  INNER JOIN unidades
  ON professores.idunidade = unidades.id
  WHERE SubQuery.expr2 >= 15
  ORDER BY SubQuery.expr1 DESC, SubQuery.expr2 DESC
  LIMIT 10";
$melhores_professores_result = $connection->query($sql) or die($connection->error);

?>
<title>Destaques</title>
<h2>10 Melhores</h2>
<hr />
<h4>Cursos Avaliados</h4>
<table class="table table-bordered">
  <tr>
    <td><div align="left"><strong>Média</strong></div></td>
    <td><div align="left"><strong>Votos</strong></div></td>
    <td><div align="left"><strong>Disciplina</strong></div></td>
    <td><div align="left"><strong>Unidade</strong></div></td>
    <td><div align="left"><strong>Professor</strong></div></td>
    <td><div align="left"></div></td>
  </tr>
  <?php  while ($row = $melhores_disciplnas_result->fetch_assoc()) { ?>
    <tr>
      <td><div align="center"><?php echo number_format($row['media'], 2, ',', ' '); ?></div></td>
      <td><div align="center"><?php echo $row['votos']; ?></div></td>
      <td><?php echo $row['codigo']; ?>-<?php echo $row['materia']; ?></td>
      <td><?php echo $row['unidade']; ?></td>
      <td><?php echo $row['professor']; ?></td>
      <td>
        <div align="center"><a href="<?= $url_full; ?>/?p=ver&id=<?php echo $row['id']; ?>" class="btn btn-success  btn-sm">
          Avaliar
          </a>
      </div></td>
    </tr>
    <?php } ?>
</table>


<h4>Professores Avaliados</h4>
<table class="table table-bordered">
  <tr>
    <td><div align="left"><strong>Média</strong></div></td>
    <td><div align="left"><strong>Votos</strong></div></td>
    <td><div align="left"><strong>Professor</strong></div></td>
    <td><div align="left"><strong>Unidade</strong></div></td>
    <td><div align="left"></div></td>
  </tr>
  <?php  while ($row = $melhores_professores_result->fetch_assoc()) { ?>
    <tr>
      <td><div align="center"><?php echo number_format($row['nota'], 2, ',', ' '); ?></div></td>
      <td><div align="center"><?php echo $row['votos']; ?></div></td>
      <td><?php echo $row['nome']; ?></td>
      <td><?php echo $row['unidade']; ?></td>
      <td>
        <div align="center"><a href="<?= $url_full; ?>/?p=pesquisa2&id=<?php echo $row['id']; ?>&t=2" class="btn btn-success  btn-sm">
          Avaliar
          </a>
      </div></td>
    </tr>
    <?php } ?>
</table>
<small>Só foram avaliados professores com 15 ou mais votos.</small> 

<?php if ($melhores_disciplnas_result) $melhores_disciplnas_result->close(); ?>
<?php if ($melhores_professores_result) $melhores_professores_result->close(); ?>
