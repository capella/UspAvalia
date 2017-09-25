<?php 
if ($user){
	mysql_select_db($database_connection, $connection);
	if(
	isset($_POST['v'])&&$_POST['v']=='verificador'&&
	((isset($_POST['sigla'])&&$_POST['sigla']!='')||
	(isset($_POST['ndisciplina'])&&$_POST['ndisciplina']!=''))&&
	isset($_POST['nprofessor'])&&$_POST['nprofessor']!=''){
		
		function novolaco($idpro,$iddis){
			global $connection, $user_profile;
			$query_AP = "SELECT * FROM `aulaprofessor` WHERE `idaula` =".$iddis." AND `idprofessor` =".$idpro;
			$AP = mysql_query($query_AP, $connection) or die(mysql_error());
			$row_AP = mysql_fetch_assoc($AP);
			$totalRows_AP = mysql_num_rows($AP);
			if($totalRows_AP>0){
				header('Location: ?p=ver&id='.$row_AP['id']);
			} else {
				$insertSQL = sprintf("INSERT INTO aulaprofessor (idaula, idprofessor, time, uso) VALUES (%s, %s, %s, %s)",
						 GetSQLValueString($iddis, "int"),
						 GetSQLValueString($idpro, "int"),
						 GetSQLValueString(time(), "int"),
						 GetSQLValueString(hash($hash , $secret_key.$user_profile['id']), "text"));
				
				$Result1 = mysql_query($insertSQL, $connection) or die(mysql_error());
				header('Location: ?p=ver&id='.mysql_insert_id());
			}
		}
		$query_dis = "SELECT * FROM `disciplinas` WHERE `nome` LIKE '%".$_POST['ndisciplina']."%' OR `codigo` LIKE '%".$_POST['sigla']."%'";// AND idunidade = ".$_POST['unidade'];
		$query_pro = "SELECT * FROM `professores` WHERE `nome` LIKE '%".$_POST['nprofessor']."%'";// AND idunidade = ".$_POST['unidade'];
				
		$DIS = mysql_query($query_dis, $connection) or die(mysql_error());
		$row_DIS = mysql_fetch_assoc($DIS);
		$totalRows_DIS = mysql_num_rows($DIS);
		
		$PRO = mysql_query($query_pro, $connection) or die(mysql_error());
		$row_PRO = mysql_fetch_assoc($PRO);
		$totalRows_PRO = mysql_num_rows($PRO);
		
		$busca_disid=$row_DIS['id'];
		$busca_proid=$row_PRO['id'];
		
		if($totalRows_PRO>0&&$totalRows_DIS>0){
			novolaco($busca_proid,$busca_disid);
		}
		if($totalRows_PRO == 0){
			$insertSQL = sprintf("INSERT INTO professores (nome, idunidade, time, uso) VALUES (%s, %s, %s, %s)",
					 GetSQLValueString(ucwords($_POST['nprofessor']), "text"),
					 GetSQLValueString($_POST['unidade'], "int"),
					 GetSQLValueString(time(), "int"),
					 GetSQLValueString(hash($hash , $secret_key.$user_profile['id']), "text"));
			
			$Result1 = mysql_query($insertSQL, $connection) or die(mysql_error());
			$busca_proid = mysql_insert_id();
		}
		if($totalRows_DIS == 0){
			$insertSQL = sprintf("INSERT INTO disciplinas (nome, idunidade, codigo, time, uso) VALUES (%s, %s, %s, %s, %s)",
					 GetSQLValueString(ucwords($_POST['ndisciplina']), "text"),
					 GetSQLValueString($_POST['unidade'], "int"),
					 GetSQLValueString(strtoupper($_POST['sigla']), "text"),
					 GetSQLValueString(time(), "int"),
					 GetSQLValueString(hash($hash , $secret_key.$user_profile['id']), "text"));
			
			$Result1 = mysql_query($insertSQL, $connection) or die(mysql_error());
			$busca_disid = mysql_insert_id();
		}
		novolaco($busca_proid,$busca_disid);
	}
	if(isset($_POST['v'])&&$_POST['v']=='verificador'): ?>
    	<br><br>
  		<div class="alert alert-danger">Preencha corretamente o formulário.</div>
  	<? endif;
	
	
	$query_Pesquisa = "SELECT * FROM (SELECT AP.id, AP.idaula, AP.idprofessor, DIS.nome as 'Dnome', PRO.nome as 'Pnome', codigo FROM aulaprofessor AP INNER JOIN disciplinas DIS ON AP.idaula = DIS.id INNER JOIN professores PRO ON AP.idprofessor = PRO.id) S WHERE idprofessor =".$_GET['id'];
	
	
$Pesquisa = mysql_query($query_Pesquisa, $connection) or die(mysql_error());
$row_Pesquisa = mysql_fetch_assoc($Pesquisa);
?>
<h2>Adionar Disciplina</h2>
<hr>

                
<form role="form" method="post" action="?p=add"> 
  <div class="form-group">
    <label for="exampleInputPassword1">Sigla da disciplina:</label>
    <input class="form-control" name="sigla" placeholder="Nomalmente 7 caracteres no formato XXX9999)">
  </div>
  <div class="form-group">
    <label for="exampleInputPassword1">Nome da disciplina:</label>
    <input class="form-control" name="ndisciplina" placeholder="">
  </div>
  <div class="form-group">
    <label for="exampleInputPassword1">Nome do professor:</label>
    <input class="form-control" placeholder="De preferência completo." disabled value="<?=$row_Pesquisa['Pnome']?>">
  </div>
  <input type="hidden" name="nprofessor" value="<?=$_GET['id']?>">
  <input type="hidden" value="verificador" name="v">
  <button type="submit" class="btn btn-default">Salvar</button>
</form>
<?php
} else {
  //$statusUrl = $facebook->getLoginStatusUrl();
  $loginUrl = $facebook->getLoginUrl();
  header('Location: '.$loginUrl);
}
?>
