package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160"
)

const (
	targetAddress    = "13zb1hQbWVsc2S7ZTZnP2G4undNNpdh5so"
	keysFile         = "keys.txt"
	searchRangesFile = "search_ranges.json"
	fixedPrefix      = "00000000000000000000000000000000000000000000000"
	maxKeysPerRange  = 100000
	numGoroutines    = 10 // Número de goroutines no pool
)

type Range struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

func main() {
	ranges, err := loadRanges(searchRangesFile)
	if err != nil {
		log.Fatalf("Falha ao carregar os intervalos: %v", err)
	}

	fmt.Printf("Iniciando a busca...\nNúmero de intervalos a serem testados: %d\n", len(ranges))

	// Limita a quantidade de CPUs utilizadas para 85%
	maxCPUs := runtime.NumCPU()
	desiredCPUs := int(float64(maxCPUs) * 0.85)
	runtime.GOMAXPROCS(desiredCPUs)

	var wg sync.WaitGroup
	found := make(chan struct{})
	results := make(chan string, 1)
	workChan := make(chan Range, len(ranges))

	// Inicia o pool de goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go worker(workChan, found, results, &wg)
	}

	go func() {
		for _, r := range ranges {
			select {
			case <-found:
				close(workChan)
				return
			default:
				workChan <- r
			}
		}
		close(workChan)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	select {
	case result := <-results:
		fmt.Println(result)
	case <-found:
	}

	fmt.Println("Busca finalizada.")
}

func worker(workChan <-chan Range, found chan struct{}, results chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for r := range workChan {
		if searchInRange(r, found, results) {
			close(found)
			return
		}
	}
}

func loadRanges(filename string) ([]Range, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var ranges []Range
	if err := json.Unmarshal(bytes, &ranges); err != nil {
		return nil, err
	}

	return ranges, nil
}

func searchInRange(r Range, found chan struct{}, results chan string) bool {
	select {
	case <-found:
		return false
	default:
	}

	keysGenerated := 0
	startTime := time.Now()

	for keysGenerated < maxKeysPerRange {
		select {
		case <-found:
			return false
		default:
			privateKey, err := generatePrivateKeyWithPrefix(r)
			if err != nil {
				log.Fatalf("Erro ao gerar chave privada: %v", err)
			}

			publicKey := generatePublicKey(privateKey)
			keysGenerated++

			if publicKey == targetAddress {
				elapsedTime := time.Since(startTime).Seconds()
				result := fmt.Sprintf("\nChave encontrada em %.2f segundos!\nChaves testadas: %d\nChave privada: %s\nEndereço público: %s\n", elapsedTime, keysGenerated, privateKey, publicKey)
				results <- result
				saveKeyToFile(privateKey, publicKey)
				return true
			}
		}
	}

	fmt.Printf("Nenhuma chave encontrada no intervalo %s-%s.\n", r.Min, r.Max)
	return false
}

func generatePrivateKeyWithPrefix(r Range) (string, error) {
	min := new(big.Int)
	max := new(big.Int)
	if _, success := min.SetString(r.Min[2:], 16); !success {
		return "", fmt.Errorf("Erro ao converter min para big.Int")
	}
	if _, success := max.SetString(r.Max[2:], 16); !success {
		return "", fmt.Errorf("Erro ao converter max para big.Int")
	}

	rangeSize := new(big.Int).Sub(max, min)

	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	randomNumber := new(big.Int).SetBytes(randomBytes)
	randomNumber.Mod(randomNumber, rangeSize).Add(randomNumber, min)

	privateKey := fmt.Sprintf("%s%s", fixedPrefix, randomNumber.Text(16))
	return privateKey, nil
}

func generatePublicKey(privKeyHex string) string {
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		log.Fatalf("Erro ao decodificar chave privada: %v", err)
	}

	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	pubKey := privKey.PubKey()
	compressedPubKey := pubKey.SerializeCompressed()
	pubKeyHash := hash160(compressedPubKey)
	address := encodeAddress(pubKeyHash, &chaincfg.MainNetParams)
	return address
}

func hash160(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	sha256Hash := h.Sum(nil)
	r := ripemd160.New()
	r.Write(sha256Hash)
	return r.Sum(nil)
}

func encodeAddress(pubKeyHash []byte, params *chaincfg.Params) string {
	versionedPayload := append([]byte{params.PubKeyHashAddrID}, pubKeyHash...)
	checksum := doubleSha256(versionedPayload)[:4]
	fullPayload := append(versionedPayload, checksum...)
	return base58Encode(fullPayload)
}

func doubleSha256(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func base58Encode(input []byte) string {
	var result []byte
	x := new(big.Int).SetBytes(input)
	base := big.NewInt(int64(len(base58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}
	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	for _, b := range input {
		if b != 0 {
			break
		}
		result = append([]byte{base58Alphabet[0]}, result...)
	}
	return string(result)
}

func saveKeyToFile(privateKey, publicKey string) {
	keyData := fmt.Sprintf("Chave privada: %s, Endereço público: %s\n", privateKey, publicKey)
	f, err := os.OpenFile(keysFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo de chaves: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(keyData); err != nil {
		log.Fatalf("Erro ao gravar a chave no arquivo: %v", err)
	}
}
