package controllers

import (
	"context"
	"github.com/sirupsen/logrus"
	v3 "go.etcd.io/etcd/client/v3"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type DeploymentBackUp struct {
	KeyPrefix string
	Ctx       context.Context
	Client    *v3.Client
}

type DeploymentBackUpSpec struct {
	Name      string `json:"name"`
	OwnerName string `json:"owner_name"`
	Status    bool   `json:"status"`
}

func (db *DeploymentBackUp) Run() {
	go db.WatchForEvents()
	go db.Reconcile()
}

func (db *DeploymentBackUp) WatchForEvents() {
	log := logrus.WithField("controller", "deployment_backup").WithField("method", "WatchForEvents")
	termSignal := make(chan os.Signal)
	signal.Notify(termSignal, os.Interrupt, syscall.SIGTERM)
	for {
		log.Debug("waiting for item")
		select {
		case watchRes := <-db.Client.Watch(db.Ctx, db.KeyPrefix, v3.WithPrefix()):
			log.Debug("got item")
			for _, event := range watchRes.Events {
				log.Debugf("handling event for %s", string(event.Kv.Key))
			}
		case <-termSignal:
			return
		}
	}
}

func (db *DeploymentBackUp) Reconcile() {
	log := logrus.WithField("controller", "deployment_backup").WithField("method", "Reconcile")

	for range time.Tick(2 * time.Second) {
		log.Debug("reconciling at ", time.Now().Format(time.RFC3339))
		time.Sleep(5 * time.Second)
	}
}
