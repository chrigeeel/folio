package main

import (
	"context"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("starting...")

	owner := solana.MustPublicKeyFromBase58("DvxJjNdTVPSknYDVPQbWZyZZ5uqNRLEiYDHoW1U7FbhC")
	account := solana.MustPublicKeyFromBase58("DJ9aKByfVrNzW1r7MTKCBVED9izCVat9gENXLQn4fq9x")

	rpcClient := rpc.New(os.Getenv("RPC_URL"))

	before := solana.Signature{}

	balanceRes, err := rpcClient.GetTokenAccountBalance(
		context.TODO(),
		account,
		rpc.CommitmentConfirmed,
	)
	if err != nil {
		panic(err)
	}

	balance, _ := strconv.ParseUint(balanceRes.Value.Amount, 10, 64)

	log.Printf("balance now: %v SOL", math.Round(float64(balance)/float64(solana.LAMPORTS_PER_SOL)*100)/100)

	for {
		signatures, err := rpcClient.GetSignaturesForAddressWithOpts(
			context.TODO(),
			account,
			&rpc.GetSignaturesForAddressOpts{
				Before: before,
			},
		)
		if err != nil {
			panic(err)
		}

		before = signatures[len(signatures)-1].Signature

		for _, signature := range signatures {
			if time.Since(signature.BlockTime.Time()) > time.Minute*5 {
				v := uint64(0)
				txData, err := rpcClient.GetTransaction(
					context.TODO(),
					signature.Signature,
					&rpc.GetTransactionOpts{
						MaxSupportedTransactionVersion: &v,
					},
				)
				if err != nil {
					panic(err)
				}

				var postBalance uint64
				for _, balance := range txData.Meta.PostTokenBalances {
					if balance.Owner.Equals(owner) && balance.Mint.Equals(solana.SolMint) {
						postBalance, _ = strconv.ParseUint(balance.UiTokenAmount.Amount, 10, 64)
					}
				}

				log.Printf("last hour: %v SOL", math.Round(float64(balance-postBalance)/float64(solana.LAMPORTS_PER_SOL)*100)/100)
				return
			}
		}
	}
}
