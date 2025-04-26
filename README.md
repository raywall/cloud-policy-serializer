# Profile

Para essa atividade você será um especialista em engenharia de software e desenvolvimento em Go

# Context

O objetivo do projeto e criar um framework reutilizável de validações de politicas, que pode ser executado em uma lambda na AWS, respondendo a requisições feitas via API Gateway
No entanto, vamos desenvolver o projeto em partes evolutivas, assim, vou gradualmente realizando solicitações que devem sempre considerar a etapa anterior para que se complementem
Além disso, cada solicitação deve resultar em códigos-fonte com anotação de seus respectivos nomes de arquivo, além dos arquivos de testes com cenários diversos que garantam a eficacia e alcance dos objetivos

Quando estiver pronto posso enviar a primeira solicitação.

1. A primeira parte do framework consiste na instancia do contexto PolicyEngineContext, e nele serão carregados:

- Schema json que representa a estrutura de dados recebidos pela aplicação;
- Schema json que representa a estrutura de dados que será montada na resposta da aplicação;
- Arquivo YAML com as políticas a serem validadas/aplicadas;
  Exemplo:
  {
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
  "id": {
  "type": "string",
  "description": "Identificador único da solicitação"
  },
  "timestamp": {
  "type": "string",
  "format": "date-time",
  "description": "Data e hora da solicitação (opcional)"
  },
  "context": {
  "type": "object",
  "description": "Contexto da solicitação (opcional)",
  "properties": {
  "userId": {
  "type": "string",
  "description": "ID do usuário"
  },
  "source": {
  "type": "string",
  "description": "Origem da solicitação"
  }
  },
  "additionalProperties": false
  },
  "data": {
  "type": "object",
  "description": "Dados para aplicação das políticas (obrigatório)",
  "additionalProperties": true
  },
  "policies": {
  "type": "array",
  "description": "Lista de políticas a serem aplicadas (obrigatório)",
  "items": {
  "type": "string"
  },
  "minItems": 1
  }
  },
  "required": [
  "id",
  "data",
  "policies"
  ],
  "additionalProperties": false
  }

## Contexto

A aplicação deverá instanciar um contexto e carregar os schemas de request, response e as politicas a serem aplicadas.
Ao receber uma requisição, de um API Gateway por exemplo, ele deverá extrair o body da requisição e validar utilizando os metodos de validação do framwork para garantir se os dados estão recebidos estão estruturados corretamente.
Por enquanto, apenas construa as funções para analise da requisição e validação dos dados do request recebido pela aplicação, irei testar, fazer os ajustes necessários e a partir dai farei a segunda solicitação.

2. Como dito anteriormente, e importante lembrar que estamos indo passo a passo, então nao faremos validação de politicas neste ponto. O pkg/core/request/validator.go ira apenas validar a requisição recebida com base no schema json que foi definido para o padrão de entrada.

3. Uma vez que a validação foi feita, vamos a segunda parte do problema. Precisamos criar o pkg/engine/policy_engine.go que ira conter os métodos responsáveis por processar e validar o conteúdo do request recebido com base nas politicas definidas no arquivo de politicas. Nao quero utilizar bibliotecas de terceiros para esta tarefa, então vc deve construir toda lógica de escrita, compreensão e validação das politicas. A ideia e que a partir dos dados recebidos, e guiado pelas políticas definidas, possamos executar a validação do conteúdo recebido e preparar para resposta (que sera nossa próxima etapa), as politicas podem nao apenas validar como comparar e definir valores também, afinal precisaremos responder depois, então funções como SET que definem valor são importantes. Este e um exemplo de como a politica sera escrita em yaml: # Arquivo de Políticas

ValidarIdade:

- data.idade >= 18
- data.tipo == "adulto"

ValidarValorTransacao:

- data.valor > 0
- data.valor <= data.limiteMaximo
- data.moeda == "BRL" || data.moeda == "USD"

CalcularDesconto:

- data.valor > 100
- SET data.desconto = data.valor \* 0.1
- IF data.cliente.tipo == "premium" THEN SET data.desconto = data.valor \* 0.15

AplicarImpostos:

- SET data.impostos = {}
- SET data.impostos.iss = data.valor \* 0.05
- IF data.tipo == "servico" THEN SET data.impostos.pis = data.valor \* 0.0165

ValidarEndereco:

- data.endereco.cep != null
- data.endereco.cidade != null
- data.endereco.estado IN ["SP", "RJ", "MG", "RS"]

VerificarLimites:

- data.transacoes.count() < data.limites.maxTransacoes
- SUM(data.transacoes.map(t => t.valor)) + data.valor <= data.limites.valorTotal

4. Na última etapa vamos responder a requisição, e para isso é importante saber que da mesma forma que utilizamos um schema para validar a estrutura do body da requisição recebida, teremos um schema definindo como devemos montar o body do response da requisição. Assim, após a validação utilizaremos o pkg/core/response/ para montar e devolver o resultado do processamento.

5. Ótimo, agora preciso de duas coisas: a primeira é que monte cenários de teste para os códigos gerados para todas as etapas: request, validação, processamento de politicas e response, de forma a atingir uma cobertura de testes maior ou igual a 90% do código gerado e eliminar qualquer chance de falhas de lógica. E segundo preciso que monte um readme.md para documentar o projeto, explicando objetivo, funcionalidades, métodos, dando exemplos, sintaxe, modo de uso e citando as características principais que mencionou e finalizando com uma seção de boas práticas.

6. Perfeito, agora vamos as melhorias. Se não ouvir dependências entre as políticas indicadas, é recomendável que sejam feitas em paralelo utilizando goroutines, mas lembre-se se o número de politicas é variável e depende do que foi escrito no arquivo yaml de politicas, assim, o código precisa entender a relação entre elas e executar em paralelo o que não houver dependência para otimizar o processo de análise e reduzir o tempo de resposta
# aws-policy-engine-go
