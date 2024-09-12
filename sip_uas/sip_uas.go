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
	"time"
)

var sipCalls = make(map[string]chan *SipMessage)
var sipServer *sipgo.Server

func StartSipServer(username string, localHost string, port int, publicHost string) {
	ua, _ := sipgo.NewUA()             // Build user agent
	sipServer, _ = sipgo.NewServer(ua) // Creating server handle
	client, _ := sipgo.NewClient(ua)   // Creating client handle
	go runSipUasDialog(username, localHost, port, publicHost, sipServer, client)
}

func runSipUasDialog(username string, localHost string, port int, publicHost string,
	srv *sipgo.Server, client *sipgo.Client) (server *sipgo.Server) {

	log.Info().Msg("Running UAS With localHost " + localHost + ":" + strconv.Itoa(port) +
		" PublicHost: " + publicHost)
	sip.SIPDebug = true

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	uasContact := sip.ContactHeader{
		Address: sip.Uri{User: username, Host: publicHost, Port: port},
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
				hostname:    publicHost,
				sipPort:     port,
			}
			go sipCall.HandleSipRequests(&sipCalls)
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
	panic(srv.ListenAndServe(context.TODO(), "udp", fmt.Sprintf("%s:%d", localHost, port)))
	log.Info().Msg("Stopping SIP Server")
	select {
	case <-sig:
		log.Info().Msg("Stopping server")
		srv.Close()
		return
	}
}

func StopSipServer() {
	log.Info().Msg("Stopping SIP Server")
	for callId := range sipCalls {
		sipCalls[callId] <- &SipMessage{request: &sip.Request{Method: "SEND_BYE"}, tx: nil}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() {
		for len(sipCalls) > 0 {
		}
		done <- struct{}{}
	}()
	select {
	case <-done:
		log.Info().Msg("All SIP Calls Closed!")
		break
	case <-ctx.Done():
		log.Warn().Msg("Time out before closing SIP Calls!")
		break
	}
	sipServer.Close()
	log.Info().Msg("Stopped SIP server")
}
