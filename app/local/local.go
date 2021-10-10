package local

import (
	"github.com/adrianliechti/loop/app"
	"github.com/adrianliechti/loop/pkg/cli"
)

var Command = &cli.Command{
	Name:  "local",
	Usage: "local development servers",

	Category: app.CategoryDevelopment,

	HideHelpCommand: true,

	Subcommands: []*cli.Command{
		mariadbCommand,
		mongoDBCommand,
		postgresCommand,

		influxdbCommand,
		redisCommand,
		elasticsearchCommand,

		minioCommand,
		vaultCommand,

		natsCommand,
		rabbitmqCommand,

		mailtrapCommand,

		codeCommand,
		grafanaCommand,
		jupyterCommand,
	},
}
