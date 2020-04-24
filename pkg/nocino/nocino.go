package nocino

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/frapposelli/nocino/pkg/gif"

	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type Nocino struct {
	API         *tgbotapi.BotAPI
	BotUsername string
	Numw        int
	Plen        int
	GIFmaxsize  int
	TrustedMap  map[int]bool
	Log         *logrus.Entry
}

func NewNocino(tgtoken string, trustedIDs string, numw int, plen int, gifmaxsize int, logger *logrus.Logger) *Nocino {
	trustedMap := make(map[int]bool)
	if trustedIDs != "" {
		ids := strings.Split(trustedIDs, ",")
		for i := 0; i < len(ids); i++ {
			j, _ := strconv.Atoi(ids[i])
			trustedMap[j] = true
		}
	}
	logfields := logger.WithField("component", "nocino")

	bot, err := tgbotapi.NewBotAPI(tgtoken)
	if err != nil {
		logfields.Fatal("Cannot log in, exiting...")
	}
	botUsername := fmt.Sprintf("@%s", bot.Self.UserName)
	logfields.Infof("Authorized on account %s", botUsername)

	return &Nocino{
		API:         bot,
		BotUsername: botUsername,
		Numw:        numw,
		Plen:        plen,
		GIFmaxsize:  gifmaxsize,
		TrustedMap:  trustedMap,
		Log:         logfields,
	}
}

func (n *Nocino) RunStatsTicker(markov *bbolt.DB, gifdb *gif.GIFDB) {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			bucketStats := 0
			err := markov.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("Chain"))
				if b != nil {
					bucketStats = b.Stats().KeyN
				}
				return nil
			})
			if err != nil {
				n.Log.Errorf("boltdb transaction failed with: '%s'", err)
			}
			n.Log.Infof("Nocino Stats: %d Markov suffixes, %d GIF in Database", bucketStats, len(gifdb.List))
		}
	}()
}
