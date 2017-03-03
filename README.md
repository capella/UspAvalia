# UspAvalia

Para mais informações acesse [http://uspavalia.com/sobre](http://uspavalia.com/sobre).

### Instalação

Para executar o sistema localmente a primeira parte é renomear o arquivo `config_example.php` para `config.php`. Nesse arquivo você deve colocar os dados para acesso ao banco de dados e sua chave do facebook. Além disso uma chave de segurança para auxiliar na criptografia dos dados.

Posteriormente a isso é necessário popular as matérias no banco de dados. Para isso utilizamos o seguinte arquivo disponibilizado pelo pessoal do matrusp: `http://bcc.ime.usp.br/matrusp/db/db_usp.txt`. Pra popular é necessário fazer o download desse arquivo e editar o arquivo `testes/index.php`
