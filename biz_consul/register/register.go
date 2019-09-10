package main

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/pborman/uuid"
	"os"
	"strconv"
)

func Register(consulHost, consulPort, svcHost, svcPort string, logger log.Logger) (registrar sd.Registrar) {
	var client consul.Client
	{
		consulCfg := api.DefaultConfig()
		consulCfg.Address = consulHost + ":" + consulPort
		consultClient, err := api.NewClient(consulCfg)
		if err != nil {
			logger.Log("create consul error:", err)
			os.Exit(1)
		}

		client = consul.NewClient(consultClient)
	}

	check := api.AgentServiceCheck{
		HTTP:     "http://" + svcHost + ":" + svcPort + "/health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "consul check service health status",
	}

	port, _ := strconv.Atoi(svcPort)

	reg := api.AgentServiceRegistration{
		ID:      "biz-" + uuid.New(),
		Name:    "biz",
		Address: svcHost,
		Port:    port,
		Tags:    []string{"biz", "vc"},
		Check:   &check,
	}

	registrar = consul.NewRegistrar(client, &reg, logger)
	return
}
