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

const (
	checkTime = time.Hour
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("starting...")

	owner := solana.MustPublicKeyFromBase58("AD65fgYti96iSSzSPaNazV9Bs29m7JbNomGjG4Cp5WFS")
	account, _, _ := solana.FindAssociatedTokenAddress(owner, solana.SolMint)

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
			if time.Since(signature.BlockTime.Time()) < checkTime {
				continue
			}
			v := uint64(0)
			slot, err := rpcClient.GetBlockWithOpts(
				context.TODO(),
				signature.Slot,
				&rpc.GetBlockOpts{
					TransactionDetails:             rpc.TransactionDetailsFull,
					MaxSupportedTransactionVersion: &v,
				},
			)
			if err != nil {
				panic(err)
			}

			var postBalance uint64
			var lastTx rpc.TransactionWithMeta
			for _, tx := range slot.Transactions {
				for _, balance := range tx.Meta.PostTokenBalances {
					if balance.Owner.Equals(owner) && balance.Mint.Equals(solana.SolMint) {
						postBalance, _ = strconv.ParseUint(balance.UiTokenAmount.Amount, 10, 64)
						lastTx = tx
					}
				}
			}
			log.Println(postBalance, lastTx.MustGetTransaction().Signatures[0])
			log.Println(signature.Signature.String())
			profit := math.Round(float64(balance-postBalance)/float64(solana.LAMPORTS_PER_SOL)*100) / 100

			profitPerHour := profit / checkTime.Hours()

			log.Printf("last %s: %v SOL", checkTime.String(), profit)
			log.Printf("%v SOL/h", profitPerHour)
			return
		}
	}
}
