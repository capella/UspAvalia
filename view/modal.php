<form id="form<?=$row_Pesquisa['id'];?>" action="" method="get">  
  <div class="modal-header">
    <button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button>
    <h4 class="modal-title"> <?=$row_Pesquisa['Pnome'];?> - <?=$row_Pesquisa['codigo'];?></h4>
  </div>
  <div class="modal-body">
   Escolha uma nota entre 0 e 5 para avaliar <?=$row_Pesquisa['Pnome'];?>, na disciplina <?=$row_Pesquisa['Dnome'];?> - <?=$row_Pesquisa['codigo'];?> nos seguintes quesitos. Em duficuldade, notas maiores significam maior dificuldade. O voto é secreto.  <br><br>
  <?
  $arr = array('Avaliação Geral', 'Didática', 'Empenho/Dedicação', 'Relação com os alunos', 'Dificuldade');
  reset($arr);
  while (list($key, $value) = each($arr)) {
	  $chave = $key+1;
  ?>
    <hr style="margin-bottom: 6px; margin-top:0;">
    <b><?=$value;?></b>
    <div style=" text-align:center;">
      <div class="input select rating-c" style="margin: auto; width:260px;">
          <select id="select<?=$chave;?><?=$row_Pesquisa['id'];?>" name="nota<?=$chave;?>">
              <option value=""></option>
              <option value="0">0</option>
              <option value="1">1</option>
              <option value="2">2</option>
              <option value="3">3</option>
              <option value="4">4</option>
              <option value="5">5</option>
          </select>
        <script type="text/javascript">
        $('#select<?=$chave;?><?=$row_Pesquisa['id'];?>').barrating('show', {
            showValues:true,
            showSelectedRating:false
        });
        </script>
      </div>
    </div>
	<? } ?>
    <input type="hidden" name="id" value="<?=$row_Pesquisa['id'];?>" />
    <input type="hidden" name="Rid" value="<?=$_GET['id'];?>" />
    <input type="hidden" name="Rt" value="<?=$_GET['t'];?>" />
    <input type="hidden" name="p" value="votar" />
  </div><!--modal body-->
  <div class="modal-footer">
  
    <div class="inline" style="float: left; text-align: center; margin-top: -8px;width: 220px;"><small>Será solicitado validação no Facebook para autenticação do usuário.</small></div>
    <br class="visible-xs">
    <button type="submit" class="btn btn-primary">Salvar</button>
  </div>
</form>
