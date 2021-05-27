package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"go-udy/control-loops/controllers"
	v3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logrus.SetLevel(logrus.TraceLevel)

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	cli, err := v3.New(v3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(),
		},
	})
	if err != nil {
		logrus.WithError(err).Error("failed to connect to etcd")
		return
	}
	defer cli.Close()

	ctx,cancel := context.WithCancel(context.Background())

	backup := &controllers.BackUp{
		KeyPrefix: "/backup",
		Client:    cli,
		Ctx:       ctx,
	}

	deploymentBackup := &controllers.DeploymentBackUp{
		KeyPrefix: "/deployment_backup",
		Client:    cli,
		Ctx:       ctx,
	}

	go backup.Run()

	go deploymentBackup.Run()

	logrus.Info("control loops are running")
	<-interrupt
	cancel()
}
