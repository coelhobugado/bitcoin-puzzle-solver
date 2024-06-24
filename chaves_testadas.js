const fs = require('fs');

function dividirIntervalo(min, max, divisoes) {
  // Converte as strings hexadecimais para números decimais
  let minDecimal = parseInt(min, 16);
  let maxDecimal = parseInt(max, 16);

  // Converte os valores mínimos e máximos para BigIntegers
  let minBigInt = BigInt(minDecimal);
  let maxBigInt = BigInt(maxDecimal);

  // Verifica se min é menor que max
  if (minBigInt >= maxBigInt) {
    console.error("O valor mínimo deve ser menor que o valor máximo.");
    return [];
  }

  // Calcula o tamanho do passo
  let range = maxBigInt - minBigInt;
  let passo = range / BigInt(divisoes);

  // Array para armazenar os intervalos
  let intervalos = [];

  // Divide o intervalo em partes iguais
  for (let i = 0; i < divisoes; i++) {
    let intervaloMin = minBigInt + (BigInt(i) * passo);
    let intervaloMax = minBigInt + (BigInt(i + 1) * passo);

    // Converte BigInts para strings hexadecimais
    intervalos.push({
      min: '0x' + intervaloMin.toString(16).padStart(13, '0'),
      max: '0x' + intervaloMax.toString(16).padStart(13, '0')
    });
  }

  return intervalos;
}

// Definição dos parâmetros
let minimo = '21000000000000000';
let maximo = '2ffffffffffffffff';
let divisoes = 90000; // Quantidade de divisões desejada

// Dividir o intervalo
let intervalos = dividirIntervalo(minimo, maximo, divisoes);

// Estrutura do arquivo JSON
let data = JSON.stringify(intervalos, null, 2);

// Salvar no arquivo JSON
fs.writeFile('search_ranges.json', data, (err) => {
  if (err) {
    console.error('Erro ao salvar o arquivo:', err);
    return;
  }
  console.log('Arquivo search_ranges.json salvo com sucesso.');
});