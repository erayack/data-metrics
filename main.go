package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"
	"io"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ethereum/go-ethereum/params"
)



func main() {
	// Use zerolog for logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Initialize the params value (hexadecimal string)
	blockNumberChan := make(chan int, 20)

	// Define the URL
	url := "http://localhost:8545"

	go func(blockNumbersChannel chan int) {
		blockNumber := 17815200

		for {
			requestData := RequestData{
				Jsonrpc: "2.0",
				Method:  "eth_blockNumber",
				Params:  []interface{}{},
				Id:      0,
			}

			jsonData, err := json.Marshal(requestData)
			if err != nil {
				log.Error().Err(err).Msg("Error marshalling JSON")
				time.Sleep(1 * time.Second)
				continue
			}

			req, err := http.NewRequest("POST", "http://localhost:8545", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Error().Err(err).Msg("Error creating HTTP request")
				time.Sleep(1 * time.Second)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Error().Err(err).Msg("Error sending HTTP request")
				time.Sleep(1 * time.Second)
				continue
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Error().Err(err).Msg("Error reading HTTP response")
				resp.Body.Close()
				time.Sleep(1 * time.Second)
				continue
			}
			resp.Body.Close()

			var blockNumberResponse BlockNumberResponse
			err = json.Unmarshal(body, &blockNumberResponse)
			if err != nil {
				log.Error().Err(err).Msg("Error decoding response JSON")
				time.Sleep(1 * time.Second)
				continue
			}

			newBlockNumber, err := strconv.ParseUint(blockNumberResponse.Result[2:], 16, 64)
			if err != nil {
				log.Error().Err(err).Msgf("Error converting block number %s to decimal", blockNumberResponse.Result)
				continue
			}
			for blockNumber < int(newBlockNumber) {
				blockNumber = blockNumber + 1
				blockNumbersChannel <- blockNumber
			}
			time.Sleep(1 * time.Second)

		}
	}(blockNumberChan)

	for blockNumber := range blockNumberChan {
		// Prepare the request data
		requestData := RequestData{
			Jsonrpc: "2.0",
			Method:  "eth_getHeaderByNumber",
			Params:  []interface{}{fmt.Sprintf("0x%x", blockNumber)},
			Id:      0,
		}

		headerBody := processData(requestData, url)

		// Decode the response JSON
		var responseDataHeader ResponseData
		err := json.Unmarshal(headerBody, &responseDataHeader)
		if err != nil {
			log.Fatal().Err(err).Msg("Error decoding response JSON")
		}

		// Decode the extraData field from hex to a string
		extraDataBytes, err := hex.DecodeString(responseDataHeader.Result.ExtraData[2:]) // skip the '0x' prefix
		if err != nil {
			log.Fatal().Err(err).Msg("Error decoding extraData")
		}
		extraData := string(extraDataBytes)

		// Log the extraData

		// Prepare the request data for BlockByNumber
		requestDataBlockByNumber := RequestData{
			Jsonrpc: "2.0",
			Method:  "eth_getBlockByNumber",
			Params:  []interface{}{fmt.Sprintf("0x%x", blockNumber), true},
			Id:      0,
		}

		responseDataBody := processData(requestDataBlockByNumber, url)

		// Decode the response JSON for the Body
		var responseBlock Payload
		err = json.Unmarshal(responseDataBody, &responseBlock)
		if err != nil {
			log.Fatal().Err(err).Msg("Error decoding response JSON")
		}
		if len(responseBlock.Result.Transactions) == 0 {
			log.Info().Int("block_number", blockNumber).Int("txn_count", len(responseBlock.Result.Transactions)).Msg("Encountered a Block with no transactions. Skipping analysis")
			blockNumber++
			continue
		}

		gas, _ := strconv.ParseInt(responseBlock.Result.GasUsed[2:], 16, 64)
		blockValueWei, _ := strconv.ParseInt(responseBlock.Result.Transactions[len(responseBlock.Result.Transactions)-1].Value[2:], 16, 64)
		blockValueEth, _ := weiToEther(big.NewInt(blockValueWei)).Float64()
		// txns := fmt.Sprintf("%v", responseBlock.Result.Transactions)
		// log.Debug().Str("transactions", txns).Msg("Dumping txns")
		log.Info().Int("block_number", blockNumber).Int64("gas_used", gas).Int("txn_count", len(responseBlock.Result.Transactions)).Float64("block_value", blockValueEth).Msg(extraData)

		// Get the transaction reciept for all the transactions https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_gettransactionreceipt
		// Total Tip Reward = gasUsed * tipFeePerGas
		// tipFeePerGas = gasprice - baseFeePerGas
		// the baseFeePerGas is set in the block

		// Increment the params value
		blockNumber++
	}
}

// WeiToEther converts wei to ether
func weiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}