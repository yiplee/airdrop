package cmd

import (
	sdk "github.com/fox-one/fox-wallet-sdk"
	"github.com/fox-one/pkg/store/db"
	"github.com/sirupsen/logrus"
)

func provideDatabase() *db.DB {
	database, err := db.Open(cfg.DB)
	if err != nil {
		logrus.WithError(err).Fatal("open database failed")
	}

	if err := db.Migrate(database); err != nil {
		logrus.WithError(err).Fatal("migrate database failed")
	}

	return database
}

func provideBroker() *sdk.Broker {
	return sdk.NewBroker(
		cfg.Wallet.Endpoint,
		cfg.Wallet.BrokerID,
		cfg.Wallet.BrokerSecret,
		cfg.Wallet.PinSecret,
	)
}
