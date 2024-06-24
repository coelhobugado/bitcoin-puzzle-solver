# Bitcoin Puzzle Solver

Este repositório contém um programa em Go projetado para resolver o puzzle número 66 do Bitcoin em modo aleatorio, com o diferencial de dividir em quantas partes quiser o intervalo de buscas, esta definido atualmente para dividir o intervalo total em 90 mil partes,e vamos testar em cada parte 100 mil vezes ( utilize o arquivo chaves_testadas.js para efetuar a divisao.

## Funcionalidades

- Geração de chaves privadas dentro de intervalos específicos.
- Verificação se a chave privada gera o endereço público alvo.
- Salvamento das chaves privadas e públicas encontradas.
- Uso de múltiplas goroutines para acelerar o processo de busca.
- Configuração do uso de CPUs para otimização de desempenho.

## Como Usar

### Pré-requisitos

- Go 1.16 ou superior.
#### Compile o programa:
go build -o bitcoin-puzzle-solver main.go

##### Instalar as dependencias: 
-` go get github.com/btcsuite/btcd/chaincfg`
-` go get github.com/decred/dcrd/dcrec/secp256k1/v4`
- `go get golang.org/x/crypto/ripemd160`

###### Necessitamos do arquivo search_ranges_json para criar: 
- Execute o arquivo "chaves_testadas.js" com o comando 'node chaves_testadas.js'
- voce pode editar o arquivo e diminuir ou aumentar a quantidade de vezes que a divisao e feita
