package sip_uas

import (
	"context"
	"fmt"
	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"strconv"
)

var sipCalls = make(map[string]chan *SipMessage)

func RunSipUasDialog(username string, host string, port int) {

	log.Info().Msg("Running UAS With host " + host + ":" + strconv.Itoa(port))
	sip.SIPDebug = true

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	ua, _ := sipgo.NewUA()           // Build user agent
	srv, _ := sipgo.NewServer(ua)    // Creating server handle
	client, _ := sipgo.NewClient(ua) // Creating client handle

	uasContact := sip.ContactHeader{
		Address: sip.Uri{User: username, Host: host, Port: port},
	}
	dialogSrv := sipgo.NewDialogServer(client, uasContact)

	srv.OnInvite(func(req *sip.Request, tx sip.ServerTransaction) {
		var callId = req.CallID()
		reqChan, ok := sipCalls[callId.String()]
		if ok {
			log.Info().Msg("SIP Re-Invite with callId: " + callId.String())
			reqChan <- &SipMessage{
				request:  req,
				tx:       tx,
				reInvite: true,
			}
		} else {
			log.Info().Msg("SIP Invite with callId: " + callId.String())
			reqChan := make(chan *SipMessage, 10)
			sipCall := SipCall{
				callId:      callId.String(),
				req:         req,
				tx:          tx,
				dialogSrv:   dialogSrv,
				sipRequests: &reqChan,
				hostname:    host,
				sipPort:     port,
			}
			go sipCall.HandleSipRequests()
			sipCalls[callId.String()] = reqChan
			reqChan <- &SipMessage{
				request:  req,
				tx:       tx,
				reInvite: false,
			}
		}
	})

	srv.OnBye(func(req *sip.Request, tx sip.ServerTransaction) {
		var callId = req.CallID()
		reqChan, ok := sipCalls[callId.String()]
		if ok {
			log.Info().Msg("SIP BYE with callId: " + callId.String())
			reqChan <- &SipMessage{
				request: req,
				tx:      tx,
			}
		} else {
			log.Error().Msg("SIP Invalid Bye with callId: " + callId.String())
		}
	})

	srv.OnAck(func(req *sip.Request, tx sip.ServerTransaction) {
		var callId = req.CallID()
		reqChan, ok := sipCalls[callId.String()]
		if ok {
			log.Info().Msg("SIP ACK with callId: " + callId.String())
			reqChan <- &SipMessage{
				request: req,
				tx:      tx,
			}
		} else {
			log.Error().Msg("SIP Invalid ACK with callId: " + callId.String())
		}
	})

	log.Info().Msg("Starting SIP Server")
	panic(srv.ListenAndServe(context.TODO(), "udp", fmt.Sprintf("%s:%d", host, port)))
	log.Info().Msg("Stopping SIP Server")
	select {
	case <-sig:
		log.Info().Msg("Stopping server")
		srv.Close()
		return
	}
}
