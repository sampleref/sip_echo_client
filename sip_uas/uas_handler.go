package sip_uas

import (
	"context"
	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/rs/zerolog/log"
	"sip_echo_client/sip_util"
	"time"
)

var (
	contentTypeHeaderSDP = sip.ContentTypeHeader("application/sdp")
)

type SipCall struct {
	callId           string
	req              *sip.Request
	tx               sip.ServerTransaction
	dialogSrv        *sipgo.DialogServer
	sipRequests      *chan *SipMessage
	dialogSrvSession *sipgo.DialogServerSession
	hostname         string
	sipPort          int

	toTag      string
	rtpSession RTPSession
	sdpSession SDPSession
}

type SipMessage struct {
	request  *sip.Request
	tx       sip.ServerTransaction
	reInvite bool
}

func (sipCall *SipCall) HandleSipRequests(sipCalls *map[string]chan *SipMessage) {
	for sipMsg := range *sipCall.sipRequests {
		switch sipMsg.request.Method {
		case sip.INVITE:
			if sipMsg.reInvite {
				sipCall.HandleUasReInvite(sipMsg.request, sipMsg.tx)
			} else {
				sipCall.HandleUasInvite(sipMsg.request, sipMsg.tx)
			}
			break
		case sip.BYE:
			sipCall.HandleBye(sipMsg.request, sipMsg.tx)
			break
		case sip.ACK:
			sipCall.HandleAck(sipMsg.request, sipMsg.tx)
			break
		default:
			log.Info().Msg("Unhandled SIP Method " + sipMsg.request.Method.String())
		}
	}
	delete(*sipCalls, sipCall.callId)
	sip_util.StopCallStatus(sipCall.callId)
	log.Info().Msg("Stopped reading messages for " + sipCall.callId)
}

func (sipCall *SipCall) HandleUasInvite(req *sip.Request, tx sip.ServerTransaction) {
	dlgSrvSession, err := sipCall.dialogSrv.ReadInvite(req, tx)
	if err != nil {
		log.Err(err)
		sipCall.SendBye(req, tx)
		return
	}
	sipCall.dialogSrvSession = dlgSrvSession
	sipCall.dialogSrvSession.Respond(sip.StatusTrying, "Trying", nil)
	sipCall.dialogSrvSession.Respond(sip.StatusRinging, "Ringing", nil)

	sip_util.NewCallStats(sipCall.callId)

	sipCall.rtpSession = RTPSession{callId: sipCall.callId, reInvite: false}
	sipCall.sdpSession = SDPSession{callId: sipCall.callId}
	aPort, vPort, err := sipCall.rtpSession.StartRTPListenerAudioVideo()
	if err != nil {
		log.Err(err)
		sipCall.SendBye(req, tx)
		return
	}

	answer, err := sipCall.sdpSession.generateAnswer(req.Body(), sipCall.hostname, aPort, vPort)
	if err != nil {
		log.Err(err)
		sipCall.SendBye(req, tx)
		return
	}
	res := sip.NewResponseFromRequest(req, 200, "OK", answer)
	res.AppendHeader(&sip.ContactHeader{Address: sip.Uri{Host: sipCall.hostname, Port: sipCall.sipPort}})
	res.AppendHeader(&contentTypeHeaderSDP)
	if err = tx.Respond(res); err != nil {
		log.Err(err)
		sipCall.SendBye(req, tx)
		return
	}
	log.Info().Msg("Accepting SIP Invite: " + req.Contact().String() + "\n")
}

func (sipCall *SipCall) HandleUasReInvite(req *sip.Request, tx sip.ServerTransaction) {
	sipCall.toTag = req.To().Params["tag"] //Hack to preserve tag in To header
	if len(sipCall.toTag) == 0 {
		log.Error().Msg("Not a reInvite, Invalid to tag for callId: " + sipCall.callId)
		return
	}
	dlgSrvSession, err := sipCall.dialogSrv.ReadInvite(req, tx)
	if err != nil {
		log.Err(err)
		return
	}
	sipCall.dialogSrvSession = dlgSrvSession

	rtpSession := RTPSession{callId: sipCall.callId, reInvite: true}
	existingOffer := sipCall.sdpSession.isCurrentOffer(req.Body())

	var answer []byte
	if existingOffer {
		log.Info().Msg("SIP Re-Invite, Existing Offer, Reusing Answer " +
			req.Contact().String() +
			" To tag " + sipCall.toTag +
			"\n")
		answer, err = sipCall.sdpSession.MarshalSessionDescription(sipCall.sdpSession.currAnswer)
	} else {
		aPort, vPort, err := sipCall.rtpSession.StartRTPListenerAudioVideo()
		if err != nil {
			log.Err(err)
			sipCall.SendBye(req, tx)
			return
		}

		answer, err = sipCall.sdpSession.generateAnswer(req.Body(), sipCall.hostname, aPort, vPort)
		if err != nil {
			log.Err(err)
			sipCall.SendBye(req, tx)
			return
		}
	}

	res := sip.NewResponseFromRequest(req, 200, "OK", answer)
	log.Info().Msg("ReInvite Updating to tag from: " + res.To().Params["tag"] +
		" to: " + sipCall.toTag + " for callId: " + sipCall.callId)
	res.To().Params["tag"] = sipCall.toTag
	res.AppendHeader(&sip.ContactHeader{Address: sip.Uri{Host: sipCall.hostname, Port: sipCall.sipPort}})
	res.AppendHeader(&contentTypeHeaderSDP)
	if err = tx.Respond(res); err != nil {
		log.Err(err)
		sipCall.SendBye(req, tx)
		return
	}
	if existingOffer {
		log.Info().Msg("Accepting SIP Re-Invite with Existing Offer: " + req.Contact().String())
		return
	}
	sipCall.rtpSession.StopRTPListeners() //Stop Previous RTP
	sipCall.rtpSession = rtpSession       // Keep New RTP Session
	log.Info().Msg("Accepting SIP Re-Invite: " + req.Contact().String() + "\n")
}

func (sipCall *SipCall) HandleBye(req *sip.Request, tx sip.ServerTransaction) {
	log.Info().Msg("Accepting SIP BYE: " + req.From().String() + "\n")
	sipCall.rtpSession.StopRTPListeners()
	sipCall.dialogSrv.ReadBye(req, tx)
	if err := tx.Respond(sip.NewResponseFromRequest(req, sip.StatusOK, "", nil)); err != nil {
		log.Err(err)
	}
	close(*sipCall.sipRequests)
}

func (sipCall *SipCall) HandleAck(req *sip.Request, tx sip.ServerTransaction) {
	log.Info().Msg("Accepting SIP ACK: " + req.Contact().String() + "\n")
	sipCall.dialogSrv.ReadAck(req, tx)
	if err := tx.Respond(sip.NewResponseFromRequest(req, sip.StatusOK, "", nil)); err != nil {
		log.Err(err)
	}
}

func (sipCall *SipCall) SendBye(req *sip.Request, tx sip.ServerTransaction) {
	log.Info().Msg("Sending SIP BYE: " + req.Contact().String() + "\n")
	sipCall.rtpSession.StopRTPListeners()
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	sipCall.dialogSrvSession.Bye(ctx)
	close(*sipCall.sipRequests)
}

func (sipCall *SipCall) SendAck(req *sip.Request, tx sip.ServerTransaction) {
	log.Info().Msg("Sending SIP ACK: " + req.Contact().String() + "\n")
	sipCall.dialogSrvSession.Respond(sip.StatusOK, "OK", nil)
}
