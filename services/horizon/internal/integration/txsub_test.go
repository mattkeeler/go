package integration

import (
	"sync"
	"testing"

	"github.com/stellar/go/services/horizon/internal/test/integration"
	"github.com/stellar/go/txnbuild"
	"github.com/stretchr/testify/assert"
)

func TestTxsub(t *testing.T) {
	tt := assert.New(t)
	itest := integration.NewTest(t, integration.Config{})
	master := itest.Master()

	t.Run("Sanity", func(t *testing.T) {
		// simplify this to one tx per account, to align with core capabilities of one
		// tx per account per ledger.
		testAccounts := 20
		subsPerAccont := 1
		keys, accounts := itest.CreateAccounts(testAccounts, "1000")

		var wg sync.WaitGroup

		for i := 0; i < testAccounts; i++ {
			for j := 0; j < subsPerAccont; j++ {
				wg.Add(1)

				seq, err := accounts[i].GetSequenceNumber()
				assert.NoError(t, err)

				var account txnbuild.SimpleAccount
				if j == 0 {
					account = txnbuild.SimpleAccount{
						AccountID: keys[i].Address(),
						Sequence:  seq,
					}
				} else {
					account = txnbuild.SimpleAccount{
						AccountID: keys[i].Address(),
						Sequence:  seq + 1,
					}
				}

				go func(i int, j int, account txnbuild.SimpleAccount) {
					defer wg.Done()

					op := txnbuild.Payment{
						Destination: master.Address(),
						Amount:      "10",
						Asset:       txnbuild.NativeAsset{},
					}

					txResp := itest.MustSubmitOperations(&account, keys[i], &op)

					tt.Equal(accounts[i].GetAccountID(), txResp.Account)
					seq, err := account.GetSequenceNumber()
					assert.NoError(t, err)
					tt.Equal(seq, txResp.AccountSequence)
					t.Logf("%d/%d done", i, j)
				}(i, j, account)
			}
		}

		wg.Wait()
	})
}
