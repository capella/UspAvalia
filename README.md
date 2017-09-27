# UspAvalia

Para mais informações acesse [<?= $url_full; ?>/sobre](<?= $url_full; ?>/sobre).

### Instalação

Para executar o sistema localmente a primeira parte é renomear o arquivo `config_example.php` para `config.php`. Nesse arquivo você deve colocar os dados para acesso ao banco de dados e sua chave do facebook. Além disso uma chave de segurança para auxiliar na criptografia dos dados.

Posteriormente a isso é necessário ir no endereço `/INSTALL/index.php`. Para atualizar o banco de dados, também basta ir nesse endereço.

### Para Finalizar

Abaixo lista de partes do sistema que necessita reparo:

- Email: formulário de contato não está enviando todos os emails.

- Adicionar Disciplina: adicionar disciplna de maneira manual precisa ser feito.

### Nota

Esse sistema foi feito sem nenhum framework. Por isso apresenta uma estrutura não convencional. Existem uma pasta chamada view, onde estão todas as páginas. As páginas são envocadas pelo index.php.

### Atualizações

Esse código foi escrito da noite para o dia em 2014, apresentando vários erros. Em Setembro de 2017 arrumei algumas coisas, no entanto falta arrumar outras.


