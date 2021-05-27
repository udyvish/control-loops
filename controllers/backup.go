package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	v3 "go.etcd.io/etcd/client/v3"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type BackUp struct {
	KeyPrefix string
	Ctx       context.Context
	Client    *v3.Client
}

type BackUpSpec struct {
	Name   string `json:"name"`
	Status bool   `json:"status"`
}

func (b *BackUp) Run() {
	go b.WatchForEvents()
	go b.Reconcile()
}

func (b *BackUp) WatchForEvents() {
	log := logrus.WithField("controller", "backup").WithField("method", "WatchForEvents")
	termSignal := make(chan os.Signal)
	signal.Notify(termSignal, os.Interrupt, syscall.SIGTERM)
	for {
		log.Debug("waiting for item")
		select {
		case watchRes := <-b.Client.Watch(b.Ctx, b.KeyPrefix, v3.WithPrefix()):
			{
				for _, event := range watchRes.Events {
					var spec BackUpSpec
					if event.Type==v3.EventTypeDelete{
						log.Debugf("%+v",event)
					}
					err := json.Unmarshal(event.Kv.Value, &spec)
					if err != nil {
						log.WithError(err).Error("failed to unmarshal event")
						continue
					}
					err = spec.HandleEvent(b.Ctx, event, b.Client)
					if err != nil {
						log.WithError(err).Error("failed to handle event")
						continue
					}
				}
			}
		case <-termSignal:
			return
		}

	}
}

func (b *BackUp) Reconcile() {
	log := logrus.WithField("controller", "backup").WithField("method", "Reconcile")

	for range time.Tick(2 * time.Second) {
		log.Debug("reconciling at ", time.Now().Format(time.RFC3339))
		time.Sleep(5 * time.Second)
	}
}

func (s *BackUpSpec) HandleEvent(ctx context.Context, e *v3.Event, c *v3.Client) error {
	switch e.Type {
	case v3.EventTypePut:
		deploymentBackupSpec := &DeploymentBackUpSpec{
			Name:      "deployment_one",
			OwnerName: s.Name,
			Status:    false,
		}
		d, err := json.Marshal(deploymentBackupSpec)
		if err != nil {
			return err
		}
		_, err = c.Put(ctx, fmt.Sprintf("/deployment_backup/%s", deploymentBackupSpec.Name), string(d))
		if err != nil {
			return err
		}
		deploymentBackupSpec.Name = "deployment_two"
		d, err = json.Marshal(deploymentBackupSpec)
		if err != nil {
			return err
		}
		_, err = c.Put(ctx, fmt.Sprintf("/deployment_backup/%s", deploymentBackupSpec.Name), string(d))
		if err != nil {
			return err
		}
	case v3.EventTypeDelete:
		//	should have the child ref as well?
		_, err := c.Delete(ctx, "/deployment_backup/deployment_one")
		if err != nil {
			return err
		}
		_, err = c.Delete(ctx, "/deployment_backup/deployment_two")
		if err != nil {
			return err
		}
	}
	return nil
}
