package miner

import (
	"crypto/ecdsa"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/fusrodah"
	"github.com/sonm-io/core/fusrodah/hub"
)

const defaultMinerPort = ":30343"

type Server struct {
	PrivateKey *ecdsa.PrivateKey
	Hubs       []hub.HubsType
	Frd        *fusrodah.Fusrodah
	ConfFile   string
	ip         string
}

func NewServer(prv *ecdsa.PrivateKey) *Server {
	if prv == nil {
		var err error
		prv, err = crypto.GenerateKey()
		if err != nil {
			panic(err)
		}
	}

	frd := fusrodah.NewServer(prv, defaultMinerPort, common.BootNodeAddr)

	srv := Server{
		PrivateKey: prv,
		Frd:        frd,
	}

	return &srv
}

func (srv *Server) Start() {
	srv.Frd.Start()
}

func (srv *Server) Stop() {
	srv.Frd.Stop()
}

func (srv *Server) Serve() {
	srv.discovery()
}

func (srv *Server) discovery() {
	var filterID int

	done := make(chan struct{})

	filterID = srv.Frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		srv.ip = string(msg.Payload)
		srv.Frd.RemoveHandling(filterID)
		close(done)
	}, common.TopicMinerDiscover)

	select {
	case <-done:
		return
	default:
		srv.Frd.Send(srv.GetPubKeyString(), true, common.TopicHubDiscover)
		time.Sleep(time.Millisecond * 1000)
	}
}

func (srv *Server) GetHubIp() string {
	if srv.ip == "0.0.0.0" {
		srv.discovery()
	}
	return srv.ip
}

func (srv *Server) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&srv.PrivateKey.PublicKey))
	return pkString
}
